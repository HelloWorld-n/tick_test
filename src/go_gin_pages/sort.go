package go_gin_pages

import (
	"math"
	"net/http"
	"sync"
	"time"

	"tick_test/types"
	"tick_test/utils/random"
	"tick_test/utils/sorting"

	"github.com/gin-gonic/gin"
	"github.com/kodergarten/iso8601duration"
	"golang.org/x/exp/constraints"
)

type intensiveCalculationResult struct {
	Code        string
	SortType    string
	StartedAt   types.ISO8601Date
	CompletedAt types.ISO8601Date
	TimeTaken   types.ISO8601Duration
	Result      any
}

var intensiveCalculationResults []*intensiveCalculationResult

type comparableElement[T any, ord constraints.Ordered] struct {
	Element *T
	Cmp     ord
}

func makeSortFunction[T any](fn func(a0 T, a1 T) bool, sortType string) func(c *gin.Context) {
	return func(c *gin.Context) {
		var arr []T
		var code = random.RandSeq(80)
		startedAt := time.Now().UTC()
		var calcResult = new(intensiveCalculationResult)
		*calcResult = intensiveCalculationResult{
			Code:      code,
			SortType:  sortType,
			StartedAt: startedAt.Format(time.RFC3339),
			Result:    arr,
		}
		intensiveCalculationResults = append(intensiveCalculationResults, calcResult)
		c.ShouldBindBodyWithJSON(&arr)
		go func() {
			var wg sync.WaitGroup
			wg.Add(1)
			sorting.SimpleSort(arr, fn, &wg)
			wg.Wait()
			now := time.Now().UTC()
			calcResult.CompletedAt = now.Format(time.RFC3339)
			calcResult.TimeTaken = (&iso8601duration.Duration{Duration: time.Time.Sub(now, startedAt)}).String()
			calcResult.Result = arr
		}()
		c.JSON(http.StatusOK, code)
	}
}

var incrementalSort = makeSortFunction(
	func(a0 float64, a1 float64) bool {
		return a0 < a1
	},
	"increase",
)

var decrementalSort = makeSortFunction(
	func(a0 float64, a1 float64) bool {
		return a0 > a1
	},
	"decrease",
)

var absoluteIncrementalSort = makeSortFunction(
	func(a0 float64, a1 float64) bool {
		return math.Abs(a0) < math.Abs(a1)
	},
	"increase-abs",
)

var absoluteDecrementalSort = makeSortFunction(
	func(a0 float64, a1 float64) bool {
		return math.Abs(a0) < math.Abs(a1)
	},
	"decrease-abs",
)

var intensiveSort = makeSortFunction(
	func(a0 []float64, a1 []float64) bool {
		return doIntensiveCalculation(a0) < doIntensiveCalculation(a1)
	},
	"calculative/intensive",
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
	startedAt time.Time,
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
	now := time.Now().UTC()
	calcResult.CompletedAt = now.Format(time.RFC3339)
	calcResult.TimeTaken = (&iso8601duration.Duration{Duration: time.Time.Sub(now, startedAt)}).String()
	calcResult.Result = arr
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
	startedAt := time.Now().UTC()
	var calcResult = new(intensiveCalculationResult)
	*calcResult = intensiveCalculationResult{
		Code:      code,
		SortType:  "calculative/calculate-once",
		StartedAt: startedAt.Format(time.RFC3339),
		Result:    arr,
	}
	intensiveCalculationResults = append(intensiveCalculationResults, calcResult)
	go sortIntensiveCalculation(cmpArr, calcResult, startedAt)

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
			c.JSON(http.StatusOK, v)
			return
		}
	}
	c.Status(http.StatusNoContent)
}

func findAllStoredResult(c *gin.Context) {
	c.JSON(http.StatusOK, intensiveCalculationResults)
}

func prepareSort(route *gin.RouterGroup) {
	intensiveCalculationResults = make([]*intensiveCalculationResult, 0)
	route.GET("", findAllStoredResult)
	route.GET("/code/:code", findStoredResult)
	route.POST("/increase", incrementalSort)
	route.POST("/decrease", decrementalSort)
	route.POST("/increase-abs", absoluteIncrementalSort)
	route.POST("/decrease-abs", absoluteDecrementalSort)
	route.POST("/calculative/intensive", intensiveSort)
	route.POST("/calculative/calculate-once", sortIntensivelyCalculatedObjectForComparation)
}
