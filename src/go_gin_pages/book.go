package go_gin_pages

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"tick_test/repository"
	"tick_test/types"
	"tick_test/utils/errDefs"
	"tick_test/utils/random"

	"github.com/gin-gonic/gin"
)

type bookHandler struct {
	repo           repository.BookRepository
	accountHandler *accountHandler
}

func NewBookHandler(bookRepo repository.BookRepository) (res *bookHandler) {
	return &bookHandler{
		repo: bookRepo,
	}
}

func (bh *bookHandler) GetAllBooksHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		books, err := bh.repo.FindAllBooks()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, books)
	}
}

func (bh *bookHandler) GetPaginatedBooksHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		pageSize, err := strconv.Atoi(c.Query("pageSize"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}
		pageNumber, err := strconv.Atoi(c.Query("pageNumber"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}
		if pageNumber <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "pageNumber must be greater than 0"})
			return
		}
		if pageSize <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "pageSize must be greater than 0"})
			return
		}

		books, err := bh.repo.FindPaginatedBooks(pageSize, pageNumber)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, books)
	}
}

func (bh *bookHandler) GetBookHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.Param("code")
		book, err := bh.repo.FindBookByCode(code)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusConflict, gin.H{"Error": "book not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, book)
	}
}

func (bh *bookHandler) PostBookHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var book types.Book
		if err := c.ShouldBindJSON(&book); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}

		book.Code = random.RandSeq(80)

		if book.Title == "" {
			c.JSON(http.StatusBadRequest, fmt.Errorf("%w; field Title", errDefs.ErrMissingField))
			return
		}
		if book.Author == "" {
			c.JSON(http.StatusBadRequest, fmt.Errorf("%w; field Author", errDefs.ErrMissingField))
			return
		}

		if err := bh.repo.CreateBook(&book); err != nil {
			c.JSON(errDefs.DetermineStatus(err), gin.H{"Error": "code already exists"})
			return
		}

		c.JSON(http.StatusCreated, book)
	}
}

func (bh *bookHandler) PatchBookHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.Param("code")

		_, err := bh.repo.FindBookByCode(code)
		if err != nil {
			if err.Error() == sql.ErrNoRows.Error() {
				c.JSON(http.StatusConflict, gin.H{"Error": "book not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}

		var updates types.Book
		if err := c.ShouldBindJSON(&updates); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}

		updatedBook, err := bh.repo.UpdateBookByCode(code, updates)
		if err != nil {
			c.JSON(errDefs.DetermineStatus(err), gin.H{"Error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, updatedBook)
	}
}

func (bh *bookHandler) DeleteBookHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.Param("code")
		rowsAffected, err := bh.repo.RemoveBookByCode(code)
		if err != nil {
			returnError(c, err)
			return
		}

		if rowsAffected > 0 {
			c.JSON(http.StatusAccepted, nil)
		} else {
			c.JSON(http.StatusOK, nil)
		}
	}
}

func (bh *bookHandler) RoleRequirer(handler gin.HandlerFunc, roles []string) func(c *gin.Context) {
	return func(c *gin.Context) {
		claims, err := bh.accountHandler.ConfirmAccountFromGinContext(c)
		if err != nil {
			returnError(c, err)
			c.Abort()
			return
		}

		authorized := false
		for _, role := range roles {
			if role == claims.Role {
				authorized = true
				break
			}
		}

		if !authorized {
			c.JSON(http.StatusForbidden, gin.H{
				"error":      "user does not have the required role",
				"validRoles": roles,
			})
			c.Abort()
			return
		}

		handler(c)
	}
}

func (bh *bookHandler) requireBookKeeperRole(handler gin.HandlerFunc) gin.HandlerFunc {
	return bh.RoleRequirer(handler, []string{"Admin", "BookKeeper"})
}

func (bh *bookHandler) prepareBook(route *gin.RouterGroup) {
	route.GET("/all", bh.GetAllBooksHandler())
	route.GET("/", bh.GetPaginatedBooksHandler())
	route.GET("/code/:code", bh.GetBookHandler())
	route.POST("/create", bh.requireBookKeeperRole(bh.PostBookHandler()))
	route.PATCH("/code/:code", bh.requireBookKeeperRole(bh.PatchBookHandler()))
	route.DELETE("/code/:code", bh.requireBookKeeperRole(bh.DeleteBookHandler()))
}
