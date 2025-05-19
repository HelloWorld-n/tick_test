package go_gin_pages_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"tick_test/go_gin_pages"
	"tick_test/go_gin_pages/mocks"
	"tick_test/types"
	"tick_test/utils/errDefs"
	"tick_test/utils/random"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetAllBooksHandler(t *testing.T) {
	testCases := []struct {
		name            string
		repo            *mocks.BookRepositoryMock
		expectedStatus  int
		expectedPayload string
	}{
		{
			name: "Success",
			repo: &mocks.BookRepositoryMock{
				FindAllBooksFn: func() ([]types.Book, error) {
					return []types.Book{
						{Code: "CODE_ZERO", Title: "BOOK", Author: "WRITER"},
						{Code: "CODE_ONE", Title: "MEDIA", Author: "AUTHOR"},
					}, nil
				},
			},
			expectedStatus:  http.StatusOK,
			expectedPayload: `[{"code":"CODE_ZERO","title":"BOOK","author":"WRITER"},{"code":"CODE_ONE","title":"MEDIA","author":"AUTHOR"}]`,
		},
		{
			name: "Empty result set",
			repo: &mocks.BookRepositoryMock{
				FindAllBooksFn: func() ([]types.Book, error) {
					return []types.Book{}, nil
				},
			},
			expectedStatus:  http.StatusOK,
			expectedPayload: `[]`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bh := go_gin_pages.NewBookHandler(tc.repo)
			handler := bh.GetAllBooksHandler()

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			handler(c)

			assert.Equal(t, tc.expectedStatus, w.Code)
			assert.JSONEq(t, tc.expectedPayload, w.Body.String())
		})
	}
}

func TestGetPaginatedBooksHandler(t *testing.T) {
	testCases := []struct {
		name            string
		repo            *mocks.BookRepositoryMock
		pageSize        string
		pageNumber      string
		expectedStatus  int
		expectedPayload string
	}{
		{
			name:       "Success",
			pageSize:   "2",
			pageNumber: "2",
			repo: &mocks.BookRepositoryMock{
				FindPaginatedBooksFn: func(pageSize, pageNumber int) ([]types.Book, error) {
					return []types.Book{
						{Code: "CODE_ZERO", Title: "BOOK", Author: "WRITER"},
						{Code: "CODE_ONE", Title: "MEDIA", Author: "AUTHOR"},
					}, nil
				},
			},
			expectedStatus:  http.StatusOK,
			expectedPayload: `[{"code":"CODE_ZERO","title":"BOOK","author":"WRITER"},{"code":"CODE_ONE","title":"MEDIA","author":"AUTHOR"}]`,
		},
		{
			name:       "Fail - invalid page size",
			pageSize:   "invalid",
			pageNumber: "1",
			repo: &mocks.BookRepositoryMock{
				FindPaginatedBooksFn: func(pageSize, pageNumber int) ([]types.Book, error) {
					return nil, nil
				},
			},
			expectedStatus:  http.StatusBadRequest,
			expectedPayload: `{"Error":"strconv.Atoi: parsing \"invalid\": invalid syntax"}`,
		},
		{
			name:       "Fail - invalid page number",
			pageSize:   "1",
			pageNumber: "invalid",
			repo: &mocks.BookRepositoryMock{
				FindPaginatedBooksFn: func(pageSize, pageNumber int) ([]types.Book, error) {
					return nil, nil
				},
			},
			expectedStatus:  http.StatusBadRequest,
			expectedPayload: `{"Error":"strconv.Atoi: parsing \"invalid\": invalid syntax"}`,
		},
		{
			name:       "Fail - page number less than 1",
			pageSize:   "6",
			pageNumber: "0",
			repo: &mocks.BookRepositoryMock{
				FindPaginatedBooksFn: func(pageSize, pageNumber int) ([]types.Book, error) {
					return nil, nil
				},
			},
			expectedStatus:  http.StatusBadRequest,
			expectedPayload: `{"Error":"pageNumber must be greater than 0"}`,
		},
		{
			name:       "Fail - page size less than 1",
			pageSize:   "0",
			pageNumber: "3",
			repo: &mocks.BookRepositoryMock{
				FindPaginatedBooksFn: func(pageSize, pageNumber int) ([]types.Book, error) {
					return nil, nil
				},
			},
			expectedStatus:  http.StatusBadRequest,
			expectedPayload: `{"Error":"pageSize must be greater than 0"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bh := go_gin_pages.NewBookHandler(tc.repo)
			handler := bh.GetPaginatedBooksHandler()

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			if c.Request == nil {
				c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
			}
			q := c.Request.URL.Query()
			q.Add("pageSize", tc.pageSize)
			q.Add("pageNumber", tc.pageNumber)
			c.Request.URL.RawQuery = q.Encode()

			handler(c)

			assert.Equal(t, tc.expectedStatus, w.Code)
			assert.JSONEq(t, tc.expectedPayload, w.Body.String())
		})
	}
}

func TestGetBookHandler(t *testing.T) {
	testCases := []struct {
		name            string
		repo            *mocks.BookRepositoryMock
		code            string
		expectedStatus  int
		expectedPayload string
	}{
		{
			name: "Success",
			code: "C123",
			repo: &mocks.BookRepositoryMock{
				FindBookByCodeFn: func(code string) (types.Book, error) {
					return types.Book{Code: code, Title: "BOOK", Author: "WRITER"}, nil
				},
			},
			expectedStatus:  http.StatusOK,
			expectedPayload: `{"code":"C123","title":"BOOK","author":"WRITER"}`,
		},
		{
			name: "Fail - Book not found",
			code: "C123",
			repo: &mocks.BookRepositoryMock{
				FindBookByCodeFn: func(code string) (types.Book, error) {
					return types.Book{}, sql.ErrNoRows
				},
			},
			expectedStatus:  http.StatusConflict,
			expectedPayload: `{"Error":"book not found"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bh := go_gin_pages.NewBookHandler(tc.repo)
			handler := bh.GetBookHandler()

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = []gin.Param{{Key: "code", Value: tc.code}}
			handler(c)

			assert.Equal(t, tc.expectedStatus, w.Code)
			assert.JSONEq(t, tc.expectedPayload, w.Body.String())
		})
	}
}

func TestPostBookHandler(t *testing.T) {
	testCases := []struct {
		name            string
		repo            *mocks.BookRepositoryMock
		inputPayload    string
		expectedStatus  int
		expectedPayload string
	}{
		{
			name: "Success",
			repo: &mocks.BookRepositoryMock{
				CreateBookFn: func(book *types.Book) error {
					return nil
				},
			},
			inputPayload:    `{"title":"BOOK","author":"WRITER"}`,
			expectedStatus:  http.StatusCreated,
			expectedPayload: `{"code":"` + random.RandSeq(80) + `","title":"BOOK","author":"WRITER"}`,
		},
		{
			name: "Fail - Bind error",
			repo: &mocks.BookRepositoryMock{
				CreateBookFn: func(book *types.Book) error { return nil },
			},
			inputPayload:    `invalid json`,
			expectedStatus:  http.StatusBadRequest,
			expectedPayload: `{"Error":"json: cannot unmarshal invalid json into Go value of type types.Book"}`,
		},
		{
			name: "Fail - Missing title",
			repo: &mocks.BookRepositoryMock{
				CreateBookFn: func(book *types.Book) error { return nil },
			},
			inputPayload:    `{"author":"WRITER"}`,
			expectedStatus:  http.StatusBadRequest,
			expectedPayload: `{"Error":"Missing required field; field Title"}`,
		},
		{
			name: "Fail - Missing author",
			repo: &mocks.BookRepositoryMock{
				CreateBookFn: func(book *types.Book) error { return nil },
			},
			inputPayload:    `{"title":"BOOK"}`,
			expectedStatus:  http.StatusBadRequest,
			expectedPayload: `{"Error":"Missing required field; field Author"}`,
		},
		{
			name: "Fail - CreateBook error",
			repo: &mocks.BookRepositoryMock{
				CreateBookFn: func(book *types.Book) error {
					return errDefs.ErrConflict
				},
			},
			inputPayload:    `{"title":"BOOK","author":"WRITER"}`,
			expectedStatus:  http.StatusConflict,
			expectedPayload: `{"Error":"code already exists"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bh := go_gin_pages.NewBookHandler(tc.repo)
			handler := bh.PostBookHandler()

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/books", bytes.NewBufferString(tc.inputPayload))
			c.Request.Header.Set("Content-Type", "application/json")
			handler(c)

			assert.Equal(t, tc.expectedStatus, w.Code)
			if tc.expectedStatus != http.StatusBadRequest {
				if tc.expectedStatus == http.StatusCreated {
					var expected types.Book
					errExcepted := json.Unmarshal([]byte(tc.expectedPayload), &expected)
					assert.NoError(t, errExcepted)
					var actual types.Book
					errActual := json.Unmarshal(w.Body.Bytes(), &actual)
					assert.NoError(t, errActual)
					assert.Equal(t, expected.Title, actual.Title)
					assert.Equal(t, expected.Author, actual.Author)
					assert.NotEmpty(t, actual.Code)
				} else {
					assert.JSONEq(t, tc.expectedPayload, w.Body.String())
				}
			}
		})
	}
}

func TestPatchBookHandler(t *testing.T) {
	testCases := []struct {
		name            string
		repo            *mocks.BookRepositoryMock
		code            string
		inputPayload    string
		expectedStatus  int
		expectedPayload string
	}{
		{
			name: "Success",
			code: "123",
			repo: &mocks.BookRepositoryMock{
				FindBookByCodeFn: func(code string) (types.Book, error) {
					return types.Book{Code: code, Title: "Old Title", Author: "Old Author"}, nil
				},
				UpdateBookByCodeFn: func(code string, updates types.Book) (types.Book, error) {
					updatedBook := types.Book{Code: code, Title: updates.Title, Author: updates.Author}
					return updatedBook, nil
				},
			},
			inputPayload:    `{"title":"New Title","author":"New Author"}`,
			expectedStatus:  http.StatusOK,
			expectedPayload: `{"code":"123","title":"New Title","author":"New Author"}`,
		},
		{
			name: "Fail - Book not found",
			code: "123",
			repo: &mocks.BookRepositoryMock{
				FindBookByCodeFn: func(code string) (types.Book, error) {
					return types.Book{}, errors.New("sql: no rows in result set")
				},
				UpdateBookByCodeFn: func(code string, updates types.Book) (types.Book, error) {
					return types.Book{}, nil
				},
			},
			inputPayload:    `{"title":"New Title","author":"New Author"}`,
			expectedStatus:  http.StatusConflict,
			expectedPayload: `{"Error":"book not found"}`,
		},
		{
			name: "Fail - Bind error",
			code: "123",
			repo: &mocks.BookRepositoryMock{
				FindBookByCodeFn: func(code string) (types.Book, error) {
					return types.Book{Code: code, Title: "Old Title", Author: "Old Author"}, nil
				},
				UpdateBookByCodeFn: func(code string, updates types.Book) (types.Book, error) {
					return types.Book{}, nil
				},
			},
			inputPayload:   `invalid json`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Fail - UpdateBookByCode error",
			code: "123",
			repo: &mocks.BookRepositoryMock{
				FindBookByCodeFn: func(code string) (types.Book, error) {
					return types.Book{Code: code, Title: "Old Title", Author: "Old Author"}, nil
				},
				UpdateBookByCodeFn: func(code string, updates types.Book) (types.Book, error) {
					return types.Book{}, errors.New("DB error")
				},
			},
			inputPayload:   `{"title":"New Title","author":"New Author"}`,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bh := go_gin_pages.NewBookHandler(tc.repo)
			handler := bh.PatchBookHandler()

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = []gin.Param{{Key: "code", Value: tc.code}}
			c.Request = httptest.NewRequest(http.MethodPatch, "/books/123", bytes.NewBufferString(tc.inputPayload))
			c.Request.Header.Set("Content-Type", "application/json")
			handler(c)

			assert.Equal(t, tc.expectedStatus, w.Code)
			if len(tc.expectedPayload) > 0 {
				assert.JSONEq(t, tc.expectedPayload, w.Body.String())
			}
		})
	}
}

func TestDeleteBookHandler(t *testing.T) {
	testCases := []struct {
		name            string
		repo            *mocks.BookRepositoryMock
		code            string
		expectedStatus  int
		expectedPayload string
	}{
		{
			name: "Success - 1 row affected",
			code: "123",
			repo: &mocks.BookRepositoryMock{
				RemoveBookByCodeFn: func(code string) (int64, error) {
					return 1, nil
				},
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name: "Success - 0 rows affected",
			code: "123",
			repo: &mocks.BookRepositoryMock{
				RemoveBookByCodeFn: func(code string) (int64, error) {
					return 0, nil
				},
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			bh := go_gin_pages.NewBookHandler(tc.repo)
			handler := bh.DeleteBookHandler()

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = []gin.Param{{Key: "code", Value: tc.code}}
			handler(c)

			assert.Equal(t, tc.expectedStatus, w.Code)
			if tc.expectedPayload != "" {
				assert.JSONEq(t, tc.expectedPayload, w.Body.String())
			}
		})
	}
}
