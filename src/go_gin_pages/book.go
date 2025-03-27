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
	books, err := FindAllBooks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, books)
}

func getBook(c *gin.Context) {
	code := c.Param("code")
	book, err := FindBookByCode(code)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"Error": "book not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, book)
}

func postBook(c *gin.Context) {
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

	if err := CreateBook(&book); err != nil {
		c.JSON(http.StatusConflict, gin.H{"Error": "code already exists"})
		return
	}

	c.JSON(http.StatusCreated, book)
}

func patchBook(c *gin.Context) {
	code := c.Param("code")

	_, err := FindBookByCode(code)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"Error": "book not found"})
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

	updatedBook, err := UpdateBookByCode(code, updates)
	if err != nil {
		if err.Error() == "no fields to update" {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, updatedBook)
}

func deleteBook(c *gin.Context) {
	code := c.Param("code")

	rowsAffected, err := RemoveBookByCode(code)
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

func requireBookKeeperRole(handler gin.HandlerFunc) gin.HandlerFunc {
	return RoleRequirer(handler, []string{"Admin", "BookKeeper"})
}

func prepareBook(route *gin.RouterGroup) {
	doPostgresPreparationForBook()

	route.GET("/all", EnsureDatabaseIsOK(getAllBooks))
	route.GET("/code/:code", EnsureDatabaseIsOK(getBook))
	route.POST("/create", EnsureDatabaseIsOK(requireBookKeeperRole(postBook)))
	route.PATCH("/code/:code", EnsureDatabaseIsOK(requireBookKeeperRole(patchBook)))
	route.DELETE("/code/:code", EnsureDatabaseIsOK(requireBookKeeperRole(deleteBook)))
}
