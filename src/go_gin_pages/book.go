package go_gin_pages

import (
	"database/sql"
	"fmt"
	"net/http"
	"tick_test/utils/random"

	"github.com/gin-gonic/gin"
)

type Book struct {
	Code   string `json:"Code"`
	Title  string `json:"Title"`
	Author string `json:"Author"`
}

func getAllBooks(c *gin.Context) {
	query := `SELECT code, title, author FROM book`
	rows, err := database.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	defer rows.Close()

	books := make([]Book, 0)
	for rows.Next() {
		var book Book
		if err := rows.Scan(&book.Code, &book.Title, &book.Author); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
		books = append(books, book)
	}

	c.JSON(http.StatusOK, books)
}

func getBook(c *gin.Context) {
	code := c.Param("code")
	var book Book
	err := database.QueryRow(
		`SELECT code, title, author FROM book WHERE code = $1`,
		code,
	).Scan(&book.Code, &book.Title, &book.Author)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNoContent, gin.H{"Error": "book not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, book)
}

func createBook(c *gin.Context) {
	var book Book
	if err := c.ShouldBindJSON(&book); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	book.Code = random.RandSeq(80)

	if book.Title == "" {
		c.JSON(http.StatusBadRequest, fmt.Errorf("%w; field Title", ErrMissingField))
		return
	}
	if book.Author == "" {
		c.JSON(http.StatusBadRequest, fmt.Errorf("%w; field Author", ErrMissingField))
		return
	}

	_, err := database.Exec(
		`INSERT INTO book (code, title, author) VALUES ($1, $2, $3)`,
		book.Code, book.Title, book.Author,
	)

	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"Error": "code already exists"})
		return
	}

	c.JSON(http.StatusCreated, book)
}

func patchBook(c *gin.Context) {
	code := c.Param("code")

	var book Book
	err := database.QueryRow(
		`SELECT code, title, author FROM book WHERE code = $1`,
		code,
	).Scan(&book.Code, &book.Title, &book.Author)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNoContent, gin.H{"Error": "book not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}

	var updates Book
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	query := `UPDATE book SET `
	params := []interface{}{}
	paramCount := 1

	if updates.Title != "" {
		query += fmt.Sprintf("title = $%d, ", paramCount)
		params = append(params, updates.Title)
		paramCount++
	}
	if updates.Author != "" {
		query += fmt.Sprintf("author = $%d, ", paramCount)
		params = append(params, updates.Author)
		paramCount++
	}

	if len(params) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "no fields to update"})
		return
	}

	query = query[:len(query)-2] + fmt.Sprintf(" WHERE code = $%d", paramCount)
	params = append(params, code)

	_, err = database.Exec(query, params...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}

	_ = database.QueryRow(
		`SELECT code, title, author FROM book WHERE code = $1`,
		code,
	).Scan(&book.Code, &book.Title, &book.Author)

	c.JSON(http.StatusOK, book)
}

func deleteBook(c *gin.Context) {
	code := c.Param("code")

	result, err := database.Exec(`DELETE FROM book WHERE code = $1`, code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}

	if rowsAffected > 0 {
		c.JSON(http.StatusAccepted, nil)
	} else {
		c.JSON(http.StatusOK, nil)
	}
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

func getUserRole(username string) (string, error) {
	var role string
	query := `
		SELECT r.name 
		FROM account a 
		JOIN role r ON a.role_id = r.id 
		WHERE a.username = $1
	`
	err := database.QueryRow(query, username).Scan(&role)
	if err != nil {
		return "", err
	}
	return role, nil
}

func roleRequirer(handler gin.HandlerFunc, roles []string) (fn func(c *gin.Context)) {
	return func(c *gin.Context) {
		username, err := confirmUserFromGinContext(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"Error": err.Error()})
			c.Abort()
			return
		}

		userRole, err := getUserRole(username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "failed to retrieve user role"})
			c.Abort()
			return
		}

		ok := false
		for _, role := range roles {
			if role == userRole {
				ok = true
				break
			}
		}

		if !ok {
			c.JSON(http.StatusForbidden, gin.H{
				"Error":      "user does not have the required role",
				"ValidRoles": roles,
			})
			c.Abort()
			return
		}

		handler(c)
	}
}

func requireBookKeeperRole(handler gin.HandlerFunc) gin.HandlerFunc {
	return roleRequirer(handler, []string{"Admin", "BookKeeper"})
}

func prepareBook(route *gin.RouterGroup) {
	doPostgresPreparationForBook()

	route.GET("/all", ensureDatabaseIsOK(getAllBooks))
	route.GET("/code/:code", ensureDatabaseIsOK(getBook))
	route.POST("/create", ensureDatabaseIsOK(requireBookKeeperRole(createBook)))
	route.PATCH("/code/:code", ensureDatabaseIsOK(requireBookKeeperRole(patchBook)))
	route.DELETE("/code/:code", ensureDatabaseIsOK(requireBookKeeperRole(deleteBook)))
}
