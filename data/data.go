package data

import (
	"am.ca-server/helpers"
	"errors"
)

// AUTHORS
var author1 = helpers.Author{
	ID:          1,
	Name:        "Harper Lee",
	DateOfBirth: "April 28, 1926",
}
var author2 = helpers.Author{
	ID:          2,
	Name:        "Mark Twain",
	DateOfBirth: "November 30, 1835",
}

// BOOKS
var book1 = helpers.Book{
	ID:       1,
	Title:    "To Kill a Mockingbird",
	ISBN:     "978-0446310789",
	AuthorID: 1,
	Author:   author1,
}
var book2 = helpers.Book{
	ID:       2,
	Title:    "Go Set A Watchman",
	ISBN:     "978-0-06-240985-0",
	AuthorID: 1,
	Author:   author1,
}
var book3 = helpers.Book{
	ID:       3,
	Title:    "Adventures of Tom Sawyer",
	ISBN:     "9781094015231",
	AuthorID: 2,
	Author:   author2,
}
var book4 = helpers.Book{
	ID:       4,
	Title:    "Adventures of Huckleberry Finn",
	ISBN:     "9780001047570",
	AuthorID: 2,
	Author:   author2,
}

var authorDB []helpers.Author
var bookDB []helpers.Book

// SeedData : Initialize and seed starting GQL data
func SeedData() {
	// Add Circular References
	author1.Books = []helpers.Book{
		book1, book2,
	}
	author2.Books = []helpers.Book{
		book3, book4,
	}

	// Seed Data
	authorDB = append(authorDB, author1, author2)
	bookDB = append(bookDB, book1, book2, book3, book4)
}

// GETTERS

// GetAuthors : Get the author DB
func GetAuthors() []helpers.Author {
	return authorDB
}

// GetBooks : Get the book DB
func GetBooks() []helpers.Book {
	return bookDB
}

// GetSpecificAuthor : Get a specific Author based on ID
func GetSpecificAuthor(authorID int64) *helpers.Author {
	for _, author := range authorDB {
		if authorID == author.ID {
			return &author
		}
	}
	return nil
}

// GetSpecificBook : Get a specific Book based on ID
func GetSpecificBook(bookID int64) *helpers.Book {
	for _, book := range bookDB {
		if bookID == book.ID {
			return &book
		}
	}
	return nil
}

// SETTERS

// AddAuthor : Add an author to DB
func AddAuthor(author helpers.Author) {
	authorDB = append(authorDB, author)
}

// AddBook : Add a book to DB
func AddBook(book helpers.Book) {
	bookDB = append(bookDB, book)
}

// RemoveAuthor : Remove an author based on ID
func RemoveAuthor(authorID int64) error {
	for i, author := range authorDB {
		if authorID == author.ID {
			authorDB = append(authorDB[:i], authorDB[i+1:]...)
			return nil
		}
	}
	return errors.New("Could not find Author to be Removed")
}

// RemoveBook : Remove a book based on ID
func RemoveBook(bookID int64) error {
	for i, book := range bookDB {
		if bookID == book.ID {
			bookDB = append(bookDB[:i], bookDB[i+1:]...)
			return nil
		}
	}
	return errors.New("Could not find Book to be Removed")
}
