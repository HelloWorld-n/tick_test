package go_gin_pages

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"tick_test/types"

	"gopkg.in/go-playground/assert.v1"
)

type bookData struct {
	Title  string `json:"title"`
	Author string `json:"author"`
	Code   string `json:"code"`
}

var books []bookData

func setupBook() {
	setupIndex()
	books = make([]bookData, 0)
}

func setAuthHeaders(req *http.Request, t *testing.T) {
	username := adminCredentials.Username
	password := adminCredentials.Password
	if username == "" || password == "" {
		t.Fatal("TEST_USERNAME and TEST_PASSWORD variables must be set in file ", adminCredentialsPath)
	}
	req.Header.Set("Username", username)
	req.Header.Set("Password", password)
}

func bookCreator(title, author string) func(*testing.T) {
	return func(t *testing.T) {
		book := bookData{
			Title:  title,
			Author: author,
		}
		books = append(books, book)
		body, _ := json.Marshal(book)
		req, err := http.NewRequest(http.MethodPost, "http://"+url+"/v1/books/create", bytes.NewBuffer(body))
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		setAuthHeaders(req, t)
		fmt.Println(req)
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		fmtPrintlnRespone(resp)
		if title == "" || author == "" {
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		} else {
			assert.Equal(t, http.StatusCreated, resp.StatusCode)
		}
		if resp.StatusCode == http.StatusCreated {
			codeBytes, _ := io.ReadAll(resp.Body)
			if err := json.Unmarshal(codeBytes, &book); err != nil && (title != "" || author != "") {
				t.Fatalf("failed to decode response: %v", err)
				return
			}
			books[len(books)-1] = book
		}
	}
}

func bookFetcher(book bookData) func(*testing.T) {
	return func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "http://"+url+"/v1/books/code/"+book.Code, nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		setAuthHeaders(req, t)
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		fmtPrintlnRespone(resp)

		if resp.StatusCode == http.StatusOK {
			var newBook types.Book
			if err := json.NewDecoder(resp.Body).Decode(&newBook); err != nil {
				t.Fatalf("failed to decode book: %v", err)
			}
			assert.Equal(t, book.Code, newBook.Code)
			assert.Equal(t, book.Author, newBook.Author)
			assert.Equal(t, book.Title, newBook.Title)
		}
	}
}

func bookUpdater(code, newTitle, newAuthor string) func(*testing.T) {
	return func(t *testing.T) {
		updates := types.Book{
			Title:  newTitle,
			Author: newAuthor,
		}
		body, _ := json.Marshal(updates)
		req, err := http.NewRequest(http.MethodPatch, "http://"+url+"/v1/books/code/"+code, bytes.NewBuffer(body))
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		setAuthHeaders(req, t)
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		fmtPrintlnRespone(resp)

		if resp.StatusCode == http.StatusOK {
			var updatedBook types.Book
			if err := json.NewDecoder(resp.Body).Decode(&updatedBook); err != nil {
				t.Fatalf("failed to decode updated book: %v", err)
			}
			if newTitle != "" {
				assert.Equal(t, newTitle, updatedBook.Title)
			}
			if newAuthor != "" {
				assert.Equal(t, newAuthor, updatedBook.Author)
			}
		}
	}
}

func bookDeleter(code string) func(*testing.T) {
	return func(t *testing.T) {
		req, err := http.NewRequest(http.MethodDelete, "http://"+url+"/v1/books/code/"+code, nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		setAuthHeaders(req, t)
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		fmtPrintlnRespone(resp)
	}
}

func TestCreateBook(t *testing.T) {
	setupBook()

	t.Run("Success", bookCreator("SPELL", "CASTER"))
	t.Run("MissingTitle", bookCreator("", "WRITER"))
	t.Run("MissingAuthor", bookCreator("TEXT", ""))
}

func TestGetBook(t *testing.T) {
	setupBook()

	t.Run("CreateBook", bookCreator("BOOK", "WRITER"))
	createdBook := books[0]

	t.Run("GetExisting", bookFetcher(createdBook))

	t.Run("GetNonExisting", bookFetcher(bookData{
		Code: "FICTION",
	}))
}

func TestUpdateBook(t *testing.T) {
	setupBook()

	t.Run("Success/CreateBook", bookCreator("TRICK", "TRICKSTER"))
	createdBook := books[0]

	t.Run("Success/UpdateTitle", bookUpdater(createdBook.Code, "UPDATE", ""))
	t.Run("Success/UpdateAuthor", bookUpdater(createdBook.Code, "", "CREDIT-TAKER"))
	t.Run("Success/UpdateBoth", bookUpdater(createdBook.Code, "MODIFICATION", "CREDIT-TAKER"))

	t.Run("Failure/NoFields", bookUpdater(createdBook.Code, "", ""))

	t.Run("Failure/InvalidCode", bookUpdater("invalid_code", "NEW-TITLE", "NEW-AUTHOR"))
}

func TestDeleteBook(t *testing.T) {
	setupBook()

	t.Run("CreateBook", bookCreator("TEST", "TEST"))
	t.Run("DeleteExisting", bookDeleter("valid_code"))

	t.Run("DeleteNonExisting", bookDeleter("invalid_code"))
}
