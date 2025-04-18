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

func (bh *bookHandler) getAllBooksHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		books, err := bh.repo.FindAllBooks()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, books)
	}
}

func (bh *bookHandler) getPaginatedBooksHandler() gin.HandlerFunc {
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

func (bh *bookHandler) getBookHandler() gin.HandlerFunc {
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

func (bh *bookHandler) postBookHandler() gin.HandlerFunc {
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

func (bh *bookHandler) patchBookHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.Param("code")

		_, err := bh.repo.FindBookByCode(code)
		if err != nil {
			if err == sql.ErrNoRows {
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

func (bh *bookHandler) deleteBookHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.Param("code")
		rowsAffected, err := bh.repo.RemoveBookByCode(code)
		if err != nil {
			c.JSON(errDefs.DetermineStatus(err), gin.H{"Error": err.Error()})
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
		username, err := bh.accountHandler.confirmUserFromGinContext(c)
		if err != nil {
			c.JSON(errDefs.DetermineStatus(err), gin.H{"Error": err.Error()})
			c.Abort()
			return
		}

		userRole, err := bh.accountHandler.repo.FindUserRole(username)
		if err != nil {
			c.JSON(errDefs.DetermineStatus(err), gin.H{"Error": "failed to retrieve user role"})
			c.Abort()
			return
		}

		authorized := false
		for _, role := range roles {
			if role == userRole {
				authorized = true
				break
			}
		}

		if !authorized {
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

func (bh *bookHandler) requireBookKeeperRole(handler gin.HandlerFunc) gin.HandlerFunc {
	return bh.RoleRequirer(handler, []string{"Admin", "BookKeeper"})
}

func (bh *bookHandler) prepareBook(route *gin.RouterGroup) {
	route.GET("/all", bh.repo.EnsureDatabaseIsOK(bh.getAllBooksHandler()))
	route.GET("/", bh.repo.EnsureDatabaseIsOK(bh.getPaginatedBooksHandler()))
	route.GET("/code/:code", bh.repo.EnsureDatabaseIsOK(bh.getBookHandler()))
	route.POST("/create", bh.repo.EnsureDatabaseIsOK(bh.requireBookKeeperRole(bh.postBookHandler())))
	route.PATCH("/code/:code", bh.repo.EnsureDatabaseIsOK(bh.requireBookKeeperRole(bh.patchBookHandler())))
	route.DELETE("/code/:code", bh.repo.EnsureDatabaseIsOK(bh.requireBookKeeperRole(bh.deleteBookHandler())))
}
