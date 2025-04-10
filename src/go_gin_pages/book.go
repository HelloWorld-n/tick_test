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

func getAllBooks(c *gin.Context) {
	books, err := repository.FindAllBooks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, books)
}

func getPaginatedBooks(c *gin.Context) {
	var pageSize, pageNumber int
	var err error

	pageSize, err = strconv.Atoi(c.Query("pageSize"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	pageNumber, err = strconv.Atoi(c.Query("pageNumber"))
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

	books, err := repository.FindPaginatedBooks(pageSize, pageNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, books)
}

func getBook(c *gin.Context) {
	code := c.Param("code")
	book, err := repository.FindBookByCode(code)
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

func postBook(c *gin.Context) {
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

	if err := repository.CreateBook(&book); err != nil {
		c.JSON(errDefs.DetermineStatus(err), gin.H{"Error": "code already exists"})
		return
	}

	c.JSON(http.StatusCreated, book)
}

func patchBook(c *gin.Context) {
	code := c.Param("code")

	_, err := repository.FindBookByCode(code)
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

	updatedBook, err := repository.UpdateBookByCode(code, updates)
	if err != nil {
		c.JSON(errDefs.DetermineStatus(err), gin.H{"Error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedBook)
}

func deleteBook(c *gin.Context) {
	code := c.Param("code")

	rowsAffected, err := repository.RemoveBookByCode(code)
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

func RoleRequirer(handler gin.HandlerFunc, roles []string) (fn func(c *gin.Context)) {
	return func(c *gin.Context) {
		username, err := confirmUserFromGinContext(c)
		if err != nil {
			c.JSON(errDefs.DetermineStatus(err), gin.H{"Error": err.Error()})
			c.Abort()
			return
		}

		userRole, err := repository.FindUserRole(username)
		if err != nil {
			c.JSON(errDefs.DetermineStatus(err), gin.H{"Error": "failed to retrieve user role"})
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
	return RoleRequirer(handler, []string{"Admin", "BookKeeper"})
}

func prepareBook(route *gin.RouterGroup) {
	route.GET("/all", repository.EnsureDatabaseIsOK(getAllBooks))
	route.GET("/", repository.EnsureDatabaseIsOK(getPaginatedBooks))
	route.GET("/code/:code", repository.EnsureDatabaseIsOK(getBook))
	route.POST("/create", repository.EnsureDatabaseIsOK(requireBookKeeperRole(postBook)))
	route.PATCH("/code/:code", repository.EnsureDatabaseIsOK(requireBookKeeperRole(patchBook)))
	route.DELETE("/code/:code", repository.EnsureDatabaseIsOK(requireBookKeeperRole(deleteBook)))
}
