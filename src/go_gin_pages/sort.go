package go_gin_pages

import (
	"math"
	"net/http"
	"sync"

	"tick_test/utils/random"
	"tick_test/utils/sorting"

	"github.com/gin-gonic/gin"
	"golang.org/x/exp/constraints"
)

type intensiveCalculationResult struct {
	Code   string
	Result any
}

var intensiveCalculationResults []*intensiveCalculationResult

type comparableElement[T any, ord constraints.Ordered] struct {
	Element *T
	Cmp     ord
}

func makeSortFunction[T any](fn func(a0 T, a1 T) bool) func(c *gin.Context) {
	return func(c *gin.Context) {
		var arr []T
		var code = random.RandSeq(80)
		var calcResult = new(intensiveCalculationResult)
		*calcResult = intensiveCalculationResult{
			Code:   code,
			Result: arr,
		}
		intensiveCalculationResults = append(intensiveCalculationResults, calcResult)
		c.ShouldBindBodyWithJSON(&arr)
		go func() {
			var wg sync.WaitGroup
			wg.Add(1)
			sorting.SimpleSort(arr, fn, &wg)
			wg.Wait()
			calcResult.Result = arr
		}()
		c.JSON(http.StatusOK, code)
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

func intensiveCalculation(val *comparableElement[[]float64, float64]) {
	val.Cmp = doIntensiveCalculation(*val.Element)
}
func sortIntensiveCalculation(
	cmpArr []comparableElement[[]float64, float64],
	calcResult *intensiveCalculationResult,
) {
	var wg sync.WaitGroup
	wg.Add(1)
	sorting.SimpleSort(
		cmpArr,
		func(a0, a1 comparableElement[[]float64, float64]) bool {
			return a0.Cmp < a1.Cmp
		},
		&wg,
	)
	wg.Wait()
	var arr = make([]*[]float64, 0)
	for _, item := range cmpArr {
		arr = append(arr, item.Element)
	}
	calcResult.Result = arr
}

func sortIntensivelyCalculatedObjectForComparation(c *gin.Context) {
	var arr []*[]float64
	c.ShouldBindBodyWithJSON(&arr)

	var cmpArr = make([]comparableElement[[]float64, float64], 0)
	for _, item := range arr {
		cmpArr = append(cmpArr, comparableElement[[]float64, float64]{item, 0})
	}
	for index, item := range cmpArr {
		intensiveCalculation(&item)
		cmpArr[index] = item
	}

	var code = random.RandSeq(80)
	var calcResult = new(intensiveCalculationResult)
	*calcResult = intensiveCalculationResult{
		Code:   code,
		Result: arr,
	}
	intensiveCalculationResults = append(intensiveCalculationResults, calcResult)
	go sortIntensiveCalculation(cmpArr, calcResult)

	arr = make([]*[]float64, 0)
	for _, item := range cmpArr {
		arr = append(arr, item.Element)
	}

	c.JSON(http.StatusOK, code)
}

func findStoredResult(c *gin.Context) {
	code := c.Param("code")
	for _, v := range intensiveCalculationResults {
		if v.Code == code {
			c.JSON(http.StatusOK, v.Result)
			return
		}
	}
	c.JSON(http.StatusNoContent, nil)
}

func findAllStoredResult(c *gin.Context) {
	c.JSON(http.StatusOK, intensiveCalculationResults)
}

func prepareSort(route *gin.RouterGroup) {
	intensiveCalculationResults = make([]*intensiveCalculationResult, 0)
	route.GET("", findAllStoredResult)
	route.GET("/code/:code", findStoredResult)
	route.POST("", incrementalSort)
	route.POST("/reverse", decrementalSort)
	route.POST("/abs", absoluteIncrementalSort)
	route.POST("/abs-reverse", absoluteDecrementalSort)
	route.POST("/calculative/prioritize-memory", intensiveSort)
	route.POST("/calculative/prioritize-speed", sortIntensivelyCalculatedObjectForComparation)
}
