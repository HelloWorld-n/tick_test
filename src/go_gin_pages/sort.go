package go_gin_pages

import (
	"net/http"
	"sync"

	"tick_test/utils/sorting"

	"github.com/gin-gonic/gin"
)

func receiveSort(c *gin.Context) {
	var wg sync.WaitGroup
	var arr []int
	c.ShouldBindBodyWithJSON(&arr)
	wg.Add(1)
	sorting.SimpleSort(arr, &wg)
	wg.Wait()
	c.JSON(http.StatusOK, arr)
}

func prepareSort(route *gin.RouterGroup) {
	route.POST("", receiveSort)
}
