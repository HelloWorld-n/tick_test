package repository

import (
	"fmt"
	"tick_test/types"
	errDefs "tick_test/utils/errDefs"
)

func FindAllBooks() (books []types.Book, err error) {
	if database == nil {
		err = errDefs.ErrDatabaseOffline
		return
	}
	query := `SELECT code, title, author FROM book`
	rows, err := database.Query(query)
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

func FindBookByCode(code string) (book types.Book, err error) {
	if database == nil {
		err = errDefs.ErrDatabaseOffline
		return
	}
	err = database.QueryRow(
		`SELECT code, title, author FROM book WHERE code = $1`,
		code,
	).Scan(&book.Code, &book.Title, &book.Author)

	if err != nil {
		return types.Book{}, err
	}
	return book, nil
}

func CreateBook(book *types.Book) (err error) {
	if database == nil {
		err = errDefs.ErrDatabaseOffline
		return
	}
	_, err = database.Exec(
		`INSERT INTO book (code, title, author) VALUES ($1, $2, $3)`,
		book.Code, book.Title, book.Author,
	)
	return err
}

func UpdateBookByCode(code string, updates types.Book) (book types.Book, err error) {
	if database == nil {
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

	_, err = database.Exec(query, params...)
	if err != nil {
		return types.Book{}, err
	}

	var updatedBook types.Book
	err = database.QueryRow(
		`SELECT code, title, author FROM book WHERE code = $1`,
		code,
	).Scan(&updatedBook.Code, &updatedBook.Title, &updatedBook.Author)
	if err != nil {
		return types.Book{}, err
	}

	return updatedBook, nil
}

func RemoveBookByCode(code string) (n int64, err error) {
	if database == nil {
		err = errDefs.ErrDatabaseOffline
		return
	}
	result, err := database.Exec(`DELETE FROM book WHERE code = $1`, code)
	if err != nil {
		return 0, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return rowsAffected, nil
}

func doPostgresPreparationForBook() {
	if database != nil {
		result, err := database.Exec(`
			CREATE TABLE IF NOT EXISTS book (
				id SERIAL PRIMARY KEY,
				code varchar(100) UNIQUE NOT NULL,
				title varchar(200) NOT NULL,
				author varchar(100) NOT NULL
			);
		`)
		fmt.Println(result, err)
	}
}
