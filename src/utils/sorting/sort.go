package sorting

import (
	"sync"

	"golang.org/x/exp/constraints"
)

func JoinSortedArrays[T constraints.Ordered](arr0, arr1 []T) (result []T) {
	result = make([]T, 0)
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

func SimpleSort[T constraints.Ordered](arr []T, wg *sync.WaitGroup) {
	defer wg.Done()
	var innerWg sync.WaitGroup
	if len(arr) < 2 {
		return
	}
	innerWg.Add(2)
	go SimpleSort(arr[:len(arr)/2], &innerWg)
	go SimpleSort(arr[len(arr)/2:], &innerWg)
	innerWg.Wait()
	for i, v := range JoinSortedArrays(arr[:len(arr)/2], arr[len(arr)/2:]) {
		arr[i] = v
	}
}
