package go_gin_pages

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

func joinSortedArrays(arr0, arr1 []int) (result []int) {
	result = make([]int, 0)
	for len(arr0) > 0 && len(arr1) > 0 {
		if arr0[0] < arr1[0] {
			result = append(result, arr0[0])
			arr0 = arr0[1:]
		} else {
			result = append(result, arr1[0])
			arr1 = arr1[1:]
		}
	}
	result = append(result, arr0...)
	result = append(result, arr1...)
	return
}

func simpleSort(arr []int, wg *sync.WaitGroup) {
	defer wg.Done()
	var innerWg sync.WaitGroup
	if len(arr) < 2 {
		return
	}
	innerWg.Add(2)
	go simpleSort(arr[:len(arr)/2], &innerWg)
	go simpleSort(arr[len(arr)/2:], &innerWg)
	innerWg.Wait()
	for i, v := range joinSortedArrays(arr[:len(arr)/2], arr[len(arr)/2:]) {
		arr[i] = v
	}
}

func receiveSort(c *gin.Context) {
	var wg sync.WaitGroup
	var arr []int
	c.ShouldBindBodyWithJSON(&arr)
	wg.Add(1)
	simpleSort(arr, &wg)
	wg.Wait()
	c.JSON(http.StatusOK, arr)
}

func prepareSort(route *gin.RouterGroup) {
	route.POST("", receiveSort)
}
