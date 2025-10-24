package repository

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"sync"

	"github.com/invopop/jsonschema"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

var SYSTEM_PROMPT_FILE = "assets/prompts/classify_transaction/system.md"
var USER_PROMPT_FILE = "assets/prompts/classify_transaction/user.md"
var MODEL = openai.ChatModelGPT4_1Mini2025_04_14
var TEMPERATURE = openai.Float(0.0)
var TOP_P = openai.Float(1.0)
var BATCH_SIZE = 100

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

func buildPrompt(categories []Category, transactions []Transaction) (string, string) {
	systemPromptBytes, systemErr := os.ReadFile(SYSTEM_PROMPT_FILE)
	userPromptBytes, userErr := os.ReadFile(USER_PROMPT_FILE)
	if systemErr != nil || userErr != nil {
		return "", ""
	}
	systemPrompt := string(systemPromptBytes)
	userPrompt := string(userPromptBytes)

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
		return "", ""
	}

	return systemPrompt, userPrompt + "\n" + string(xmlOutput)
}

func categorizeBatch(client openai.Client, categories []Category, transactions []Transaction, batchNum int) ([]CategorizedTransaction, error) {
	systemPrompt, userPrompt := buildPrompt(categories, transactions)

	if systemPrompt == "" || userPrompt == "" {
		return nil, fmt.Errorf("failed to build prompt for batch %d", batchNum)
	}

	fmt.Printf("\n=== Batch %d: Processing %d transactions ===\n", batchNum, len(transactions))
	fmt.Println("\n=== System Prompt Sent to OpenAI ===")
	fmt.Println(systemPrompt)
	fmt.Println("=============================")
	fmt.Println("\n=== User Prompt Sent to OpenAI ===")
	fmt.Println(userPrompt)
	fmt.Println("=============================")

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
							OfString: openai.String(systemPrompt),
						},
					},
				},
				{
					OfUser: &openai.ChatCompletionUserMessageParam{
						Content: openai.ChatCompletionUserMessageParamContentUnion{
							OfString: openai.String(userPrompt),
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
		return nil, fmt.Errorf("failed to call OpenAI API for batch %d: %w", batchNum, err)
	}

	if len(chatCompletion.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI for batch %d", batchNum)
	}

	responseContent := chatCompletion.Choices[0].Message.Content
	if responseContent == "" {
		return nil, fmt.Errorf("empty response from OpenAI for batch %d", batchNum)
	}

	fmt.Printf("\n=== Batch %d: OpenAI Raw Response ===\n", batchNum)
	fmt.Println(responseContent)
	fmt.Println("=============================")

	var result CategorizationResponse
	if err := json.Unmarshal([]byte(responseContent), &result); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAI response for batch %d: %w", batchNum, err)
	}

	return result.Transactions, nil
}

func CategorizeTransactions(categories []Category, transactions []Transaction) ([]CategorizedTransaction, error) {
	apiKey := os.Getenv("OPEN_AI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPEN_AI_API_KEY environment variable not set")
	}

	client := openai.NewClient(option.WithAPIKey(apiKey))

	totalTransactions := len(transactions)
	if totalTransactions == 0 {
		return []CategorizedTransaction{}, nil
	}

	numBatches := (totalTransactions + BATCH_SIZE - 1) / BATCH_SIZE
	fmt.Printf("\n=== Categorizing %d transactions in %d batches ===\n", totalTransactions, numBatches)

	type batchResult struct {
		transactions []CategorizedTransaction
		err          error
	}

	results := make(chan batchResult, numBatches)
	var wg sync.WaitGroup

	for i := 0; i < numBatches; i++ {
		wg.Add(1)
		go func(batchNum int) {
			defer wg.Done()

			start := batchNum * BATCH_SIZE
			end := start + BATCH_SIZE
			if end > totalTransactions {
				end = totalTransactions
			}

			batch := transactions[start:end]
			categorized, err := categorizeBatch(client, categories, batch, batchNum+1)
			results <- batchResult{transactions: categorized, err: err}
		}(i)
	}

	wg.Wait()
	close(results)

	var allCategorized []CategorizedTransaction
	for result := range results {
		if result.err != nil {
			return nil, result.err
		}
		allCategorized = append(allCategorized, result.transactions...)
	}

	return allCategorized, nil
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
