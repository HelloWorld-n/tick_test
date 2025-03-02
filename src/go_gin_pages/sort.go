package go_gin_pages

import (
	"math"
	"net/http"
	"sync"

	"tick_test/utils/sorting"

	"github.com/gin-gonic/gin"
	"golang.org/x/exp/constraints"
)

type ComparableElement[T any, ord constraints.Ordered] struct {
	Element *T
	Cmp     ord
}

func makeSortFunction[T any](fn func(a0 T, a1 T) bool) func(c *gin.Context) {
	return func(c *gin.Context) {
		var wg sync.WaitGroup
		var arr []T
		c.ShouldBindBodyWithJSON(&arr)
		wg.Add(1)
		sorting.SimpleSort(arr, fn, &wg)
		wg.Wait()
		c.JSON(http.StatusOK, arr)
	}
}

var incrementalSort = makeSortFunction(
	func(a0 float64, a1 float64) bool {
		return a0 < a1
	},
)

var decrementalSort = makeSortFunction(
	func(a0 float64, a1 float64) bool {
		return a0 > a1
	},
)

var absoluteIncrementalSort = makeSortFunction(
	func(a0 float64, a1 float64) bool {
		return math.Abs(a0) < math.Abs(a1)
	},
)

var absoluteDecrementalSort = makeSortFunction(
	func(a0 float64, a1 float64) bool {
		return math.Abs(a0) < math.Abs(a1)
	},
)

var intensiveSort = makeSortFunction(
	func(a0 []float64, a1 []float64) bool {
		return doIntensiveCalculation(a0) < doIntensiveCalculation(a1)
	},
)

func doIntensiveCalculation(val []float64) float64 {
	var calc float64 = 1
	for _, item := range val {
		calc *= item
	}
	return calc
}

func intensiveCalculation(val *ComparableElement[[]float64, float64]) {
	val.Cmp = doIntensiveCalculation(*val.Element)
}

func sortIntensivelyCalculatedObjectForComparation(c *gin.Context) {
	var wg sync.WaitGroup
	var arr []*[]float64
	c.ShouldBindBodyWithJSON(&arr)

	var cmpArr = make([]ComparableElement[[]float64, float64], 0)
	for _, item := range arr {
		cmpArr = append(cmpArr, ComparableElement[[]float64, float64]{item, 0})
	}
	for index, item := range cmpArr {
		intensiveCalculation(&item)
		cmpArr[index] = item
	}

	wg.Add(1)
	sorting.SimpleSort(
		cmpArr,
		func(a0, a1 ComparableElement[[]float64, float64]) bool {
			return a0.Cmp < a1.Cmp
		},
		&wg,
	)
	wg.Wait()
	arr = make([]*[]float64, 0)
	for _, item := range cmpArr {
		arr = append(arr, item.Element)
	}

	c.JSON(http.StatusOK, arr)
}

func prepareSort(route *gin.RouterGroup) {
	route.POST("", incrementalSort)
	route.POST("/reverse", decrementalSort)
	route.POST("/abs", absoluteIncrementalSort)
	route.POST("/abs-reverse", absoluteDecrementalSort)
	route.POST("/calculative/prioritize-memory", intensiveSort)
	route.POST("/calculative/prioritize-speed", sortIntensivelyCalculatedObjectForComparation)
}
