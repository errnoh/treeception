package main

import (
	"sync"

//        "runtime"
)

// 'table' is just a named type for slice of string pointers.
type table []*string

// Since this version of mergesort is not in-place sort, we're using another array as helper array.
var temp table

// Sort slice of strings by doing a merge sort on their pointers instead of actual strings.
func MergeSort(t []string) []*string {
	newtable := make(table, len(t))
	temp = make(table, len(t))
	for i := 0; i < len(t); i++ {
		newtable[i] = &t[i]
	}
	if *serial {
		newtable.serialsort()
	} else {
		newtable.sort(nil, 0)
	}
	return newtable
}

// Sort a slice of string by separating it to two.
//
// No need to start a new thread for all of them, just split to new threads enough times to
// keep all cores busy and then trust that each one takes almost the same time to finish as the others.
//
// wg.Wait() waits until it has value 0.
// wg.Add(x) adds x to wg value.
// wg.Done() subtracts 1 from wg value.
func (t table) sort(ownwg *sync.WaitGroup, depth uint) {
	if len(t) < 20 {
		t.insert_sort(0, len(t))
		return
	}

	middle := len(t) / 2
	if depth < targetdepth {
		wg := new(sync.WaitGroup)
		wg.Add(2)
		go t[:middle].sort(wg, depth+1)
		go t[middle:].sort(wg, depth+1)
		wg.Wait()
	} else {
		t[:middle].sort(nil, depth+1)
		t[middle:].sort(nil, depth+1)
	}
	t.merge(middle)
	if ownwg != nil {
		ownwg.Done()
	}

	/*
		        // For debugging purposes.
			if depth < 6 {
		                v := runtime.NumGoroutine()
		                println("threads:", v, "depth:", depth)
			}
	*/
}

// non-paraller version of mergesort
func (t table) serialsort() {
	if len(t) < 20 {
		t.insert_sort(0, len(t))
		return
	}

	middle := len(t) / 2
	t[:middle].serialsort()
	t[middle:].serialsort()
	t.merge(middle)
}

// Merge two parts by proceeding from left to right while checking which one has a bigger value.
// Basic mergesort.
func (t table) merge(mid int) {
	if mid == 0 || mid == len(t) {
		return
	}

	slicepos := len(temp) - cap(t)
	tempslice := temp[slicepos : slicepos+len(t)]
	if len(t) == 2 {
		if t.compare(0, 1) < 0 {
			tempslice[0], tempslice[1] = t[1], t[0]
			return
		}
	}

	left, right, current := 0, mid, 0
	var comp int

	for {
		if left == mid {
			if right == len(t) {
				break
			}
			tempslice[current] = t[right]
			current++
			right++
			continue
		} else if right == len(t) {
			tempslice[current] = t[left]
			current++
			left++
			continue
		}

		comp = t.compare(left, right)

		if comp == 0 {
			tempslice[current] = t[left]
			tempslice[current+1] = t[right]
			current += 2
			left++
			right++
		} else if comp < 0 {
			tempslice[current] = t[left]
			left++
			current++
		} else {
			tempslice[current] = t[right]
			right++
			current++
		}
	}
	copy(t, tempslice)
}

// Compare underlying values of two pointers
func (t table) compare(i, j int) int {
	if *t[i] < *t[j] {
		return -1
	}
	if *t[i] > *t[j] {
		return 1
	}
	return 0
}

// Insertion sort.
func (t table) insert_sort(from, to int) {
	if to > from+1 {
		for i := from + 1; i < to; i++ {
			for j := i; j > from; j-- {
				if t.compare(j, j-1) < 0 {
					t[j], t[j-1] = t[j-1], t[j]
				} else {
					break
				}
			}
		}
	}
}
