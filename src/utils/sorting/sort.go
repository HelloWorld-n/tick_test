package sorting

import (
	"sync"
)

func JoinSortedArrays[T any](
	arr0, arr1 []T,
	cmp func(T, T) bool,
) (result []T) {
	result = make([]T, 0)
	for len(arr0) > 0 && len(arr1) > 0 {
		if cmp(arr0[0], arr1[0]) {
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

func SimpleSort[T any](
	arr []T,
	cmp func(T, T) bool,
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	var innerWg sync.WaitGroup
	if len(arr) < 2 {
		return
	}
	innerWg.Add(2)
	go SimpleSort(arr[:len(arr)/2], cmp, &innerWg)
	go SimpleSort(arr[len(arr)/2:], cmp, &innerWg)
	innerWg.Wait()
	for i, v := range JoinSortedArrays(arr[:len(arr)/2], arr[len(arr)/2:], cmp) {
		arr[i] = v
	}
}
