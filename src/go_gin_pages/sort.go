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

var validSortTypes []string = make([]string, 0)

var sortMutex sync.Mutex

type intensiveCalculationResult struct {
	Code        string
	SortType    string
	StartedAt   types.ISO8601Date
	CompletedAt types.ISO8601Date
	TimeTaken   types.ISO8601Duration
	Result      any
}

type intensiveCalculationMeta struct {
	SortType         string
	AverageTimeTaken types.ISO8601Duration
	MinTimeTaken     types.ISO8601Duration
	MaxTimeTaken     types.ISO8601Duration
	SampleSize       uint64
}

var intensiveCalculationResults []*intensiveCalculationResult

type comparableElement[T any, ord constraints.Ordered] struct {
	Element *T
	Cmp     ord
}

func registerSortType(sortType string) {
	validSortTypes = append(validSortTypes, sortType)
}

func makeSortFunction[T any](fn func(a0 T, a1 T) bool, sortType string) func(c *gin.Context) {
	registerSortType(sortType)
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

		sortMutex.Lock()
		intensiveCalculationResults = append(intensiveCalculationResults, calcResult)
		sortMutex.Unlock()

		c.ShouldBindBodyWithJSON(&arr)
		go func() {
			var wg sync.WaitGroup
			wg.Add(1)
			sorting.SimpleSort(arr, fn, &wg)
			wg.Wait()
			now := time.Now().UTC()

			sortMutex.Lock()
			calcResult.CompletedAt = now.Format(time.RFC3339)
			calcResult.TimeTaken = (&iso8601duration.Duration{Duration: time.Time.Sub(now, startedAt)}).String()
			calcResult.Result = arr
			sortMutex.Unlock()
		}()
		c.JSON(http.StatusCreated, code)
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
		return math.Abs(a0) > math.Abs(a1)
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

	sortMutex.Lock()
	calcResult.CompletedAt = now.Format(time.RFC3339)
	calcResult.TimeTaken = (&iso8601duration.Duration{Duration: time.Time.Sub(now, startedAt)}).String()
	calcResult.Result = arr
	sortMutex.Unlock()
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
	sortMutex.Lock()
	intensiveCalculationResults = append(intensiveCalculationResults, calcResult)
	sortMutex.Unlock()

	go sortIntensiveCalculation(cmpArr, calcResult, startedAt)

	arr = make([]*[]float64, 0)
	for _, item := range cmpArr {
		arr = append(arr, item.Element)
	}

	c.JSON(http.StatusCreated, code)
}

func fetchSortedMeta(meta *[]intensiveCalculationMeta, sortType string, wg *sync.WaitGroup) {
	defer wg.Done()

	var metaSample = intensiveCalculationMeta{
		SortType:   sortType,
		SampleSize: 0,
	}
	totalTimeTaken := 0 * time.Nanosecond
	var minTimeTaken time.Duration = math.MaxInt64
	var maxTimeTaken time.Duration = math.MinInt64

	sortMutex.Lock()
	for _, val := range intensiveCalculationResults {
		if val.SortType == sortType {
			timeTaken, err := parseISO8601Duration(val.TimeTaken, 0*time.Nanosecond)
			if err != nil {
				break
			}

			metaSample.SampleSize += 1
			if timeTaken < minTimeTaken {
				minTimeTaken = timeTaken
			}
			if timeTaken > maxTimeTaken {
				maxTimeTaken = timeTaken
			}
			totalTimeTaken += timeTaken
		}
	}
	sortMutex.Unlock()

	if metaSample.SampleSize > 0 {
		metaSample.MinTimeTaken = (&iso8601duration.Duration{Duration: minTimeTaken}).String()
		metaSample.MaxTimeTaken = (&iso8601duration.Duration{Duration: maxTimeTaken}).String()
		metaSample.AverageTimeTaken = (&iso8601duration.Duration{Duration: totalTimeTaken / time.Duration(metaSample.SampleSize)}).String()
		*meta = append(*meta, metaSample)
	}
}

func findStoredResult(c *gin.Context) {
	code := c.Param("code")
	var result *intensiveCalculationResult

	sortMutex.Lock()
	for _, v := range intensiveCalculationResults {
		if v.Code == code {
			result = v
			break
		}
	}
	sortMutex.Unlock()

	if result != nil {
		c.JSON(http.StatusOK, result)
		return
	}
	c.Status(http.StatusNoContent)
}

func findStoredResultMeta(c *gin.Context) {
	var storedResultMetaInformation = new([]intensiveCalculationMeta)
	*storedResultMetaInformation = make([]intensiveCalculationMeta, 0)
	var wg = new(sync.WaitGroup)

	for _, sortType := range validSortTypes {
		wg.Add(1)
		go fetchSortedMeta(storedResultMetaInformation, sortType, wg)
	}
	wg.Wait()
	c.JSON(http.StatusOK, storedResultMetaInformation)
}

func findAllStoredResult(c *gin.Context) {
	sortMutex.Lock()
	results := make([]*intensiveCalculationResult, len(intensiveCalculationResults))
	copy(results, intensiveCalculationResults)
	sortMutex.Unlock()
	c.JSON(http.StatusOK, results)
}

func deleteAllStoredResults(c *gin.Context) {
	sortMutex.Lock()
	intensiveCalculationResults = make([]*intensiveCalculationResult, 0)
	sortMutex.Unlock()
	c.JSON(http.StatusAccepted, nil)
}

func prepareSort(route *gin.RouterGroup) {
	sortMutex.Lock()
	intensiveCalculationResults = make([]*intensiveCalculationResult, 0)
	sortMutex.Unlock()

	route.GET("", findAllStoredResult)
	route.GET("/meta", findStoredResultMeta)
	route.GET("/code/:code", findStoredResult)
	route.POST("/increase", incrementalSort)
	route.POST("/decrease", decrementalSort)
	route.POST("/increase-abs", absoluteIncrementalSort)
	route.POST("/decrease-abs", absoluteDecrementalSort)
	route.POST("/calculative/intensive", intensiveSort)
	route.DELETE("/delete-all", deleteAllStoredResults)

	registerSortType("calculative/calculate-once")
	route.POST("/calculative/calculate-once", sortIntensivelyCalculatedObjectForComparation)
}
