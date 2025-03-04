package go_gin_pages

import (
	"math"
	"net/http"
	"sync"

	"tick_test/utils/sorting"

	"github.com/gin-gonic/gin"
)

func incrementalSort(c *gin.Context) {
	var wg sync.WaitGroup
	var arr []float64
	c.ShouldBindBodyWithJSON(&arr)
	wg.Add(1)
	sorting.SimpleSort(
		arr,
		func(a0 float64, a1 float64) bool {
			return a0 < a1
		},
		&wg,
	)
	wg.Wait()
	c.JSON(http.StatusOK, arr)
}

func decrementalSort(c *gin.Context) {
	var wg sync.WaitGroup
	var arr []float64
	c.ShouldBindBodyWithJSON(&arr)
	wg.Add(1)
	sorting.SimpleSort(
		arr,
		func(a0 float64, a1 float64) bool {
			return a0 > a1
		},
		&wg,
	)
	wg.Wait()
	c.JSON(http.StatusOK, arr)
}

func absoluteIncrementalSort(c *gin.Context) {
	var wg sync.WaitGroup
	var arr []float64
	c.ShouldBindBodyWithJSON(&arr)
	wg.Add(1)
	sorting.SimpleSort(
		arr,
		func(a0 float64, a1 float64) bool {
			return math.Abs(a0) < math.Abs(a1)
		},
		&wg,
	)
	wg.Wait()
	c.JSON(http.StatusOK, arr)
}

func absoluteDecrementalSort(c *gin.Context) {
	var wg sync.WaitGroup
	var arr []float64
	c.ShouldBindBodyWithJSON(&arr)
	wg.Add(1)
	sorting.SimpleSort(
		arr,
		func(a0 float64, a1 float64) bool {
			return math.Abs(a0) > math.Abs(a1)
		},
		&wg,
	)
	wg.Wait()
	c.JSON(http.StatusOK, arr)
}

func prepareSort(route *gin.RouterGroup) {
	route.POST("", incrementalSort)
	route.POST("/reverse", decrementalSort)
	route.POST("/abs", absoluteIncrementalSort)
	route.POST("/abs-reverse", absoluteDecrementalSort)
}
