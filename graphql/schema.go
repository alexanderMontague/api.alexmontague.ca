package graphql

import (
	"am.ca-server/data"
	"am.ca-server/helpers"
	"errors"
	// "fmt"
	GQL "github.com/graphql-go/graphql"
	"log"
	"math/rand"
	"time"
)

var bookType = GQL.NewObject(
	GQL.ObjectConfig{
		Name: "Book",
		Fields: GQL.Fields{
			"id": &GQL.Field{
				Type: GQL.Int,
			},
			"title": &GQL.Field{
				Type: GQL.String,
			},
			"isbn": &GQL.Field{
				Type: GQL.String,
			},
			"authorID": &GQL.Field{
				Type: GQL.Int,
			},
		},
	},
)

var authorType = GQL.NewObject(
	GQL.ObjectConfig{
		Name: "Author",
		Fields: GQL.Fields{
			"id": &GQL.Field{
				Type: GQL.Int,
			},
			"name": &GQL.Field{
				Type: GQL.String,
			},
			"dateOfBirth": &GQL.Field{
				Type: GQL.String,
			},
			"books": &GQL.Field{
				Type: GQL.NewList(bookType),
			},
		},
	},
)

var queryType = GQL.NewObject(
	GQL.ObjectConfig{
		Name: "Query",
		Fields: GQL.Fields{
			// hello world!
			"hello": &GQL.Field{
				Type:        GQL.String,
				Description: "Hello World!",
				Resolve: func(params GQL.ResolveParams) (interface{}, error) {
					return "world", nil
				},
			},

			// fetch a book by id
			"getBook": &GQL.Field{
				Type:        bookType,
				Description: "Get a Book by id",
				Args: GQL.FieldConfigArgument{
					"id": &GQL.ArgumentConfig{
						Type: GQL.Int,
					},
				},
				Resolve: func(params GQL.ResolveParams) (interface{}, error) {
					id, ok := params.Args["id"].(int)
					if !ok {
						return nil, errors.New("error getting book argument")
					}

					book := data.GetSpecificBook(int64(id))
					if book == nil {
						return nil, errors.New("Could not find book based on ID")
					}

					return *book, nil
				},
			},

			// fetch an author by id
			"getAuthor": &GQL.Field{
				Type:        authorType,
				Description: "Get an Author by id",
				Args: GQL.FieldConfigArgument{
					"id": &GQL.ArgumentConfig{
						Type: GQL.Int,
					},
				},
				Resolve: func(params GQL.ResolveParams) (interface{}, error) {
					id, ok := params.Args["id"].(int)
					if !ok {
						return nil, errors.New("error getting author argument")
					}

					author := data.GetSpecificAuthor(int64(id))
					if author == nil {
						return nil, errors.New("Could not find author based on ID")
					}

					return *author, nil
				},
			},

			// fetch all books
			"getBooks": &GQL.Field{
				Type:        GQL.NewList(bookType),
				Description: "Get all Books",
				Resolve: func(p GQL.ResolveParams) (interface{}, error) {
					return data.GetBooks(), nil
				},
			},

			// fetch all books
			"getAuthors": &GQL.Field{
				Type:        GQL.NewList(authorType),
				Description: "Get all Authors",
				Resolve: func(p GQL.ResolveParams) (interface{}, error) {
					return data.GetAuthors(), nil
				},
			},
		},
	})

var mutationType = GQL.NewObject(GQL.ObjectConfig{
	Name: "Mutation",
	Fields: GQL.Fields{
		// add book to DB
		"addBook": &GQL.Field{
			Type:        bookType,
			Description: "Create a new Book",
			Args: GQL.FieldConfigArgument{
				"title": &GQL.ArgumentConfig{
					Type: GQL.NewNonNull(GQL.String),
				},
				"isbn": &GQL.ArgumentConfig{
					Type: GQL.NewNonNull(GQL.String),
				},
				"authorID": &GQL.ArgumentConfig{
					Type: GQL.NewNonNull(GQL.Int),
				},
				// "author": &GQL.ArgumentConfig{
				// 	Type: authorType,
				// },
			},
			Resolve: func(params GQL.ResolveParams) (interface{}, error) {
				rand.Seed(time.Now().UnixNano())

				book := helpers.Book{
					ID:       int64(rand.Intn(100000)),
					Title:    params.Args["title"].(string),
					ISBN:     params.Args["isbn"].(string),
					AuthorID: int64(params.Args["authorID"].(int)),
					// Author:   params.Args["author"].(helpers.Author),
				}

				data.AddBook(book)
				return book, nil
			},
		},

		// add author to DB
		"addAuthor": &GQL.Field{
			Type:        authorType,
			Description: "Create a new Author",
			Args: GQL.FieldConfigArgument{
				"name": &GQL.ArgumentConfig{
					Type: GQL.NewNonNull(GQL.String),
				},
				"dateOfBirth": &GQL.ArgumentConfig{
					Type: GQL.NewNonNull(GQL.String),
				},
				// "books": &GQL.ArgumentConfig{
				// 	Type: GQL.NewList(bookType),
				// },
			},
			Resolve: func(params GQL.ResolveParams) (interface{}, error) {
				rand.Seed(time.Now().UnixNano())

				author := helpers.Author{
					ID:          int64(rand.Intn(100000)),
					Name:        params.Args["name"].(string),
					DateOfBirth: params.Args["dateOfBirth"].(string),
					// Books:       params.Args["books"].([]helpers.Book),
				}

				data.AddAuthor(author)
				return author, nil
			},
		},

		// remove book from DB
		"removeBook": &GQL.Field{
			Type:        GQL.String,
			Description: "Remove a book given an id",
			Args: GQL.FieldConfigArgument{
				"id": &GQL.ArgumentConfig{
					Type: GQL.NewNonNull(GQL.Int),
				},
			},
			Resolve: func(params GQL.ResolveParams) (interface{}, error) {
				id, ok := params.Args["id"].(int)
				if !ok {
					return nil, errors.New("error getting book id")
				}

				err := data.RemoveBook(int64(id))
				if err != nil {
					return nil, err
				}

				return "Successfully removed Book", nil
			},
		},

		// remove author from DB
		"removeAuthor": &GQL.Field{
			Type:        GQL.String,
			Description: "Remove an author given an id",
			Args: GQL.FieldConfigArgument{
				"id": &GQL.ArgumentConfig{
					Type: GQL.NewNonNull(GQL.Int),
				},
			},
			Resolve: func(params GQL.ResolveParams) (interface{}, error) {
				id, ok := params.Args["id"].(int)
				if !ok {
					return nil, errors.New("error getting author id")
				}

				err := data.RemoveAuthor(int64(id))
				if err != nil {
					return nil, err
				}

				return "Successfully removed Book", nil
			},
		},
	},
})

// GetSchema - function that returns our GQL schema
func GetSchema() GQL.Schema {
	// Add circular referenced fields
	bookType.AddFieldConfig("author", &GQL.Field{
		Type: authorType,
	})

	var schema, err = GQL.NewSchema(
		GQL.SchemaConfig{
			Query:    queryType,
			Mutation: mutationType,
		},
	)
	if err != nil {
		log.Fatalf("Failed to create GQL schema, error: %v", err)
	}

	return schema
}
