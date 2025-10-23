package repository

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"

	"github.com/invopop/jsonschema"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

var PROMPT_FILE = "assets/prompts/classify_transaction_v2.md"
var MODEL = openai.ChatModelGPT4oMini
var TEMPERATURE = openai.Float(0.0)
var TOP_P = openai.Float(1.0)

type CategorizedTransaction struct {
	ID       string `json:"id" jsonschema_description:"The unique ID of the transaction"`
	Category string `json:"category" jsonschema_description:"The category assigned to this transaction"`
}

type CategorizationResponse struct {
	Transactions []CategorizedTransaction `json:"transactions" jsonschema_description:"List of categorized transactions"`
}

func GenerateSchema[T any]() interface{} {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T
	schema := reflector.Reflect(v)
	return schema
}

var CategorizationResponseSchema = GenerateSchema[CategorizationResponse]()

type XMLCategory struct {
	XMLName xml.Name `xml:"category"`
	Name    string   `xml:"name"`
}

type XMLTransaction struct {
	XMLName     xml.Name `xml:"transaction"`
	ID          string   `xml:"id"`
	Description string   `xml:"description"`
	Merchant    string   `xml:"merchant"`
}

type XMLInput struct {
	XMLName      xml.Name         `xml:"input"`
	Categories   []XMLCategory    `xml:"categories>category"`
	Transactions []XMLTransaction `xml:"transactions>transaction"`
}

func buildPrompt(categories []Category, transactions []Transaction) string {
	promptBytes, err := os.ReadFile(PROMPT_FILE)
	if err != nil {
		return ""
	}
	prompt := string(promptBytes)

	xmlCategories := make([]XMLCategory, len(categories))
	for i, cat := range categories {
		xmlCategories[i] = XMLCategory{
			Name: cat.Name,
		}
	}

	xmlTransactions := make([]XMLTransaction, len(transactions))
	for i, tx := range transactions {
		xmlTransactions[i] = XMLTransaction{
			ID:          tx.ID,
			Description: tx.Description,
			Merchant:    tx.Merchant,
		}
	}

	input := XMLInput{
		Categories:   xmlCategories,
		Transactions: xmlTransactions,
	}

	xmlOutput, err := xml.MarshalIndent(input, "", "  ")
	if err != nil {
		return prompt
	}

	return prompt + "\n\n" + string(xmlOutput)
}

func CategorizeTransactions(categories []Category, transactions []Transaction) ([]CategorizedTransaction, error) {
	apiKey := os.Getenv("OPEN_AI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPEN_AI_API_KEY environment variable not set")
	}

	client := openai.NewClient(option.WithAPIKey(apiKey))

	prompt := buildPrompt(categories, transactions)

	fmt.Println("\n=== Prompt Sent to OpenAI ===")
	fmt.Println(prompt)
	fmt.Println("=============================\n")

	schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        "categorization_response",
		Description: openai.String("Categorized list of transactions with assigned category IDs"),
		Schema:      CategorizationResponseSchema,
		Strict:      openai.Bool(true),
	}

	chatCompletion, err := client.Chat.Completions.New(
		context.Background(),
		openai.ChatCompletionNewParams{
			Model:       MODEL,
			Temperature: TEMPERATURE,
			TopP:        TOP_P,
			Messages: []openai.ChatCompletionMessageParamUnion{
				{
					OfSystem: &openai.ChatCompletionSystemMessageParam{
						Content: openai.ChatCompletionSystemMessageParamContentUnion{
							OfString: openai.String(``),
						},
					},
				},
				{
					OfUser: &openai.ChatCompletionUserMessageParam{
						Content: openai.ChatCompletionUserMessageParamContentUnion{
							OfString: openai.String(prompt),
						},
					},
				},
			},
			ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
				OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
					JSONSchema: schemaParam,
				},
			},
		},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to call OpenAI API: %w", err)
	}

	if len(chatCompletion.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	responseContent := chatCompletion.Choices[0].Message.Content
	if responseContent == "" {
		return nil, fmt.Errorf("empty response from OpenAI")
	}

	fmt.Println("\n=== OpenAI Raw Response ===")
	fmt.Println(responseContent)
	fmt.Println("===========================\n")

	var result CategorizationResponse
	if err := json.Unmarshal([]byte(responseContent), &result); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	return result.Transactions, nil
}

func ApplyCategorizationToTransactions(categories []Category, categorized []CategorizedTransaction, transactions []Transaction) []Transaction {
	// map category name to category ID
	categoryNameMap := make(map[string]string)
	for _, cat := range categories {
		categoryNameMap[cat.Name] = cat.ID
	}

	// map transaction ID to category ID
	categoryMap := make(map[string]string)
	for _, ct := range categorized {
		categoryMap[ct.ID] = categoryNameMap[ct.Category]
	}

	result := make([]Transaction, len(transactions))
	for i, tx := range transactions {
		result[i] = tx
		if categoryID, exists := categoryMap[tx.ID]; exists {
			result[i].CategoryID = &categoryID
		}
	}

	return result
}
