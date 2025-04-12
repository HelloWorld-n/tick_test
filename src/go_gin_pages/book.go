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

func getAllBooksHandler(repo *repository.Repo) gin.HandlerFunc {
	return func(c *gin.Context) {
		books, err := repo.FindAllBooks()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, books)
	}
}

func getPaginatedBooksHandler(repo *repository.Repo) gin.HandlerFunc {
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

		books, err := repo.FindPaginatedBooks(pageSize, pageNumber)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, books)
	}
}

func getBookHandler(repo *repository.Repo) gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.Param("code")
		book, err := repo.FindBookByCode(code)
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

func postBookHandler(repo *repository.Repo) gin.HandlerFunc {
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

		if err := repo.CreateBook(&book); err != nil {
			c.JSON(errDefs.DetermineStatus(err), gin.H{"Error": "code already exists"})
			return
		}

		c.JSON(http.StatusCreated, book)
	}
}

func patchBookHandler(repo *repository.Repo) gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.Param("code")

		_, err := repo.FindBookByCode(code)
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

		updatedBook, err := repo.UpdateBookByCode(code, updates)
		if err != nil {
			c.JSON(errDefs.DetermineStatus(err), gin.H{"Error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, updatedBook)
	}
}

func deleteBookHandler(repo *repository.Repo) gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.Param("code")
		rowsAffected, err := repo.RemoveBookByCode(code)
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

func RoleRequirer(handler gin.HandlerFunc, roles []string, repo *repository.Repo) func(c *gin.Context) {
	return func(c *gin.Context) {
		username, err := confirmUserFromGinContext(c, repo)
		if err != nil {
			c.JSON(errDefs.DetermineStatus(err), gin.H{"Error": err.Error()})
			c.Abort()
			return
		}

		userRole, err := repo.FindUserRole(username)
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

func requireBookKeeperRole(handler gin.HandlerFunc, repo *repository.Repo) gin.HandlerFunc {
	return RoleRequirer(handler, []string{"Admin", "BookKeeper"}, repo)
}

func prepareBook(route *gin.RouterGroup, repo *repository.Repo) {
	route.GET("/all", repo.EnsureDatabaseIsOK(getAllBooksHandler(repo)))
	route.GET("/", repo.EnsureDatabaseIsOK(getPaginatedBooksHandler(repo)))
	route.GET("/code/:code", repo.EnsureDatabaseIsOK(getBookHandler(repo)))
	route.POST("/create", repo.EnsureDatabaseIsOK(requireBookKeeperRole(postBookHandler(repo), repo)))
	route.PATCH("/code/:code", repo.EnsureDatabaseIsOK(requireBookKeeperRole(patchBookHandler(repo), repo)))
	route.DELETE("/code/:code", repo.EnsureDatabaseIsOK(requireBookKeeperRole(deleteBookHandler(repo), repo)))
}
