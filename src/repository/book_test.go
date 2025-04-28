package repository_test

import (
	"errors"
	"regexp"
	"testing"
	"tick_test/repository"
	"tick_test/types"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestFindAllBooks(t *testing.T) {
	tests := []struct {
		name          string
		mockRows      *sqlmock.Rows
		mockError     error
		expectedBooks []types.Book
		expectError   bool
	}{
		{
			name: "Success with multiple books",
			mockRows: sqlmock.NewRows([]string{"code", "title", "author"}).
				AddRow("123", "Title 1", "Author 1").
				AddRow("456", "Title 2", "Author 2"),
			expectedBooks: []types.Book{
				{Code: "123", Title: "Title 1", Author: "Author 1"},
				{Code: "456", Title: "Title 2", Author: "Author 2"},
			},
			expectError: false,
		},
		{
			name:          "Success with no books",
			mockRows:      sqlmock.NewRows([]string{"code", "title", "author"}),
			expectedBooks: []types.Book{},
			expectError:   false,
		},
		{
			name:        "Database error",
			mockError:   errors.New("db error"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rMock, mock := setupMock(t)
			r := repository.NewRepo(&repository.Database{Conn: rMock.DB})
			defer r.DB.Conn.Close()

			query := regexp.QuoteMeta(`SELECT code, title, author FROM book`)
			expect := mock.ExpectQuery(query)

			if tt.mockError != nil {
				expect.WillReturnError(tt.mockError)
			} else {
				expect.WillReturnRows(tt.mockRows)
			}

			books, err := r.FindAllBooks()

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedBooks, books)
			}
		})
	}
}

func TestFindPaginatedBooks(t *testing.T) {
	tests := []struct {
		name          string
		pageSize      int
		pageNumber    int
		mockRows      *sqlmock.Rows
		mockError     error
		expectedBooks []types.Book
		expectError   bool
		errorMsg      string
	}{
		{
			name:       "Success with valid page and size",
			pageSize:   2,
			pageNumber: 1,
			mockRows: sqlmock.NewRows([]string{"code", "title", "author"}).
				AddRow("1", "Book 1", "Author A").
				AddRow("2", "Book 2", "Author B"),
			expectedBooks: []types.Book{
				{Code: "1", Title: "Book 1", Author: "Author A"},
				{Code: "2", Title: "Book 2", Author: "Author B"},
			},
			expectError: false,
		},
		{
			name:          "Success with empty result",
			pageSize:      2,
			pageNumber:    2,
			mockRows:      sqlmock.NewRows([]string{"code", "title", "author"}),
			expectedBooks: []types.Book{},
			expectError:   false,
		},
		{
			name:        "Invalid page number",
			pageSize:    2,
			pageNumber:  0,
			mockError:   errors.New("wrong page"),
			expectError: true,
			errorMsg:    "parameter pageNumbers needs to be 1 or greater",
		},
		{
			name:        "Invalid page size",
			pageSize:    0,
			pageNumber:  1,
			mockError:   errors.New("wrong size"),
			expectError: true,
			errorMsg:    "parameter pageSize needs to be 1 or greater",
		},
		{
			name:        "Database error",
			pageSize:    2,
			pageNumber:  1,
			mockError:   errors.New("db error"),
			expectError: true,
			errorMsg:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rMock, mock := setupMock(t)
			r := repository.NewRepo(&repository.Database{Conn: rMock.DB})
			defer r.DB.Conn.Close()

			query := regexp.QuoteMeta(`SELECT code, title, author FROM book ORDER BY id LIMIT $1 OFFSET $2`)
			expect := mock.ExpectQuery(query).WithArgs(tt.pageSize, (tt.pageNumber-1)*tt.pageSize)

			if tt.mockError != nil {
				expect.WillReturnError(tt.mockError)
			} else {
				expect.WillReturnRows(tt.mockRows)
			}

			books, err := r.FindPaginatedBooks(tt.pageSize, tt.pageNumber)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					require.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedBooks, books)
			}
		})
	}
}

func TestFindBookByCode(t *testing.T) {
	tests := []struct {
		name         string
		code         string
		mockRow      *sqlmock.Rows
		mockError    error
		expectedBook types.Book
		expectError  bool
		errorMsg     string
	}{
		{
			name: "Success",
			code: "123",
			mockRow: sqlmock.NewRows([]string{"code", "title", "author"}).
				AddRow("123", "Title 1", "Author 1"),
			expectedBook: types.Book{Code: "123", Title: "Title 1", Author: "Author 1"},
			expectError:  false,
		},
		{
			name:        "No record found",
			code:        "123",
			mockError:   errors.New("sql: no rows in result set"),
			expectError: true,
			errorMsg:    "no rows in result set",
		},
		{
			name:        "Database error",
			code:        "123",
			mockError:   errors.New("db error"),
			expectError: true,
			errorMsg:    "db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rMock, mock := setupMock(t)
			r := repository.NewRepo(&repository.Database{Conn: rMock.DB})
			defer r.DB.Conn.Close()

			query := regexp.QuoteMeta(`SELECT code, title, author FROM book WHERE code = $1`)
			expect := mock.ExpectQuery(query).WithArgs(tt.code)

			if tt.mockError != nil {
				expect.WillReturnError(tt.mockError)
			} else {
				expect.WillReturnRows(tt.mockRow)
			}

			book, err := r.FindBookByCode(tt.code)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					require.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedBook, book)
			}
		})
	}
}

func TestCreateBook(t *testing.T) {
	tests := []struct {
		name        string
		book        *types.Book
		mockError   error
		expectError bool
		errorMsg    string
	}{
		{
			name: "Success",
			book: &types.Book{
				Code:   "123",
				Title:  "Title 1",
				Author: "Author 1",
			},
			expectError: false,
		},
		{
			name: "Database error",
			book: &types.Book{
				Code:   "123",
				Title:  "Title 1",
				Author: "Author 1",
			},
			mockError:   errors.New("db error"),
			expectError: true,
			errorMsg:    "db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rMock, mock := setupMock(t)
			r := repository.NewRepo(&repository.Database{Conn: rMock.DB})
			defer r.DB.Conn.Close()

			query := regexp.QuoteMeta(`INSERT INTO book (code, title, author) VALUES ($1, $2, $3)`)
			expect := mock.ExpectExec(query).
				WithArgs(tt.book.Code, tt.book.Title, tt.book.Author)

			if tt.mockError != nil {
				expect.WillReturnError(tt.mockError)
			} else {
				expect.WillReturnResult(sqlmock.NewResult(1, 1))
			}

			err := r.CreateBook(tt.book)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					require.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRemoveBookByCode(t *testing.T) {
	tests := []struct {
		name                  string
		code                  string
		mockExecError         error
		mockRowsAffectedError error
		expectedRowsAffected  int64
		expectError           bool
		errorMsg              string
	}{
		{
			name:                 "Success",
			code:                 "123",
			expectedRowsAffected: 1,
			expectError:          false,
		},
		{
			name:                 "No rows affected",
			code:                 "123",
			expectedRowsAffected: 0,
			expectError:          false,
		},
		{
			name:          "Database error on exec",
			code:          "123",
			mockExecError: errors.New("exec error"),
			expectError:   true,
			errorMsg:      "exec error",
		},
		{
			name:                  "Error getting rows affected",
			code:                  "123",
			mockRowsAffectedError: errors.New("rows affected error"),
			expectError:           true,
			errorMsg:              "rows affected error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rMock, mock := setupMock(t)
			r := repository.NewRepo(&repository.Database{Conn: rMock.DB})
			defer r.DB.Conn.Close()

			query := regexp.QuoteMeta(`DELETE FROM book WHERE code = $1`)
			exec := mock.ExpectExec(query).WithArgs(tt.code)

			if tt.mockExecError != nil {
				exec.WillReturnError(tt.mockExecError)
			} else {
				exec.WillReturnResult(sqlmock.NewResult(0, tt.expectedRowsAffected))
				if tt.mockRowsAffectedError != nil {
					exec.WillReturnError(tt.mockRowsAffectedError)
				}
			}

			rowsAffected, err := r.RemoveBookByCode(tt.code)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					require.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedRowsAffected, rowsAffected)
			}
		})
	}
}
