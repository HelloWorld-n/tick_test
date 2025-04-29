package repository

import (
	"fmt"
	"tick_test/types"
	errDefs "tick_test/utils/errDefs"
)

type BookRepository interface {
	FindAllBooks() (books []types.Book, err error)
	FindPaginatedBooks(pageSize int, pageNumber int) (books []types.Book, err error)
	FindBookByCode(code string) (book types.Book, err error)
	CreateBook(book *types.Book) (err error)
	UpdateBookByCode(code string, updates types.Book) (book types.Book, err error)
	RemoveBookByCode(code string) (n int64, err error)
}

func (r *repo) FindAllBooks() (books []types.Book, err error) {
	if r.DB.Conn == nil {
		err = errDefs.ErrDatabaseOffline
		return
	}
	query := `SELECT code, title, author FROM book`
	rows, err := r.DB.Conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	books = make([]types.Book, 0)
	for rows.Next() {
		var book types.Book
		if err := rows.Scan(&book.Code, &book.Title, &book.Author); err != nil {
			return nil, err
		}
		books = append(books, book)
	}
	return books, nil
}

func (r *repo) FindPaginatedBooks(pageSize int, pageNumber int) (books []types.Book, err error) {
	if r.DB.Conn == nil {
		err = errDefs.ErrDatabaseOffline
		return
	}

	offset := (pageNumber - 1) * pageSize
	if pageNumber < 1 {
		return nil, fmt.Errorf("%w: parameter pageNumbers needs to be 1 or greater but it is %v", errDefs.ErrBadRequest, pageNumber)
	}
	if pageSize < 1 {
		return nil, fmt.Errorf("%w: parameter pageSize needs to be 1 or greater but it is %v", errDefs.ErrBadRequest, pageSize)
	}

	query := `SELECT code, title, author FROM book ORDER BY id LIMIT $1 OFFSET $2`
	rows, err := r.DB.Conn.Query(query, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	books = make([]types.Book, 0)
	for rows.Next() {
		var book types.Book
		if err := rows.Scan(&book.Code, &book.Title, &book.Author); err != nil {
			return nil, err
		}
		books = append(books, book)
	}
	return books, nil
}

func (r *repo) FindBookByCode(code string) (book types.Book, err error) {
	if r.DB.Conn == nil {
		err = errDefs.ErrDatabaseOffline
		return
	}
	err = r.DB.Conn.QueryRow(
		`SELECT code, title, author FROM book WHERE code = $1`,
		code,
	).Scan(&book.Code, &book.Title, &book.Author)

	if err != nil {
		return types.Book{}, err
	}
	return book, nil
}

func (r *repo) CreateBook(book *types.Book) (err error) {
	if r.DB.Conn == nil {
		err = errDefs.ErrDatabaseOffline
		return
	}
	_, err = r.DB.Conn.Exec(
		`INSERT INTO book (code, title, author) VALUES ($1, $2, $3)`,
		book.Code, book.Title, book.Author,
	)
	return err
}

func (r *repo) UpdateBookByCode(code string, updates types.Book) (book types.Book, err error) {
	if r.DB.Conn == nil {
		err = errDefs.ErrDatabaseOffline
		return
	}
	var queryFields string
	params := []interface{}{}
	paramCount := 1

	if updates.Title != "" {
		queryFields += fmt.Sprintf("title = $%d, ", paramCount)
		params = append(params, updates.Title)
		paramCount++
	}
	if updates.Author != "" {
		queryFields += fmt.Sprintf("author = $%d, ", paramCount)
		params = append(params, updates.Author)
		paramCount++
	}

	if len(params) == 0 {
		return types.Book{}, fmt.Errorf("%w: no fields to update", errDefs.ErrBadRequest)
	}

	query := fmt.Sprintf("UPDATE book SET %s WHERE code = $%d", queryFields[:len(queryFields)-2], paramCount)
	params = append(params, code)

	_, err = r.DB.Conn.Exec(query, params...)
	if err != nil {
		return types.Book{}, err
	}

	var updatedBook types.Book
	err = r.DB.Conn.QueryRow(
		`SELECT code, title, author FROM book WHERE code = $1`,
		code,
	).Scan(&updatedBook.Code, &updatedBook.Title, &updatedBook.Author)
	if err != nil {
		return types.Book{}, err
	}

	return updatedBook, nil
}

func (r *repo) RemoveBookByCode(code string) (n int64, err error) {
	if r.DB.Conn == nil {
		err = errDefs.ErrDatabaseOffline
		return
	}
	result, err := r.DB.Conn.Exec(`DELETE FROM book WHERE code = $1`, code)
	if err != nil {
		return 0, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return rowsAffected, nil
}

func (r *repo) doPostgresPreparationForBook() {
	if r.DB.Conn != nil {
		_, err := r.DB.Conn.Exec(`
			CREATE TABLE IF NOT EXISTS book (
				id SERIAL PRIMARY KEY,
				code varchar(100) UNIQUE NOT NULL,
				title varchar(200) NOT NULL,
				author varchar(100) NOT NULL
			);
		`)
		logPossibleError(err)
	}
}
