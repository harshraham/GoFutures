package main

import (
	"math/rand"
	"sync"
	"testing"
	"time"
)

//multiple threads/goroutines are trying to call the functions you have written on the same futures object.
//multiple goroutines trying to cancel the same futures object and checking if it has been cancelled.

//testing simple functions

func TestFunctions(t *testing.T) {
	var tests = []struct {
		input    int
		function func(int) string
		expected string
	}{{4, pow_2, "4 square: 16"},
		{5, pow_2, "5 square: 25"},
		{6, pow_3, "6 cube: 216"},
		{3, pow_3, "3 cube: 27"}}
	for _, test := range tests {
		if output := test.function(test.input); output != test.expected {
			t.Errorf("Test Failed: %d inputted, %s expected, received %s", test.input, test.expected, output)
		} else {
			t.Log(output)
		}
	}
}

//testing results syncronous execution
func TestFuturesSync(t *testing.T) {
	var tests = []struct {
		input    []int
		function func(int) string
	}{{[]int{1, 2, 3, 4}, pow_2}, {[]int{4, 5, 6, 7}, pow_3}}

	exec := Executor{}
	var futures []*Future
	for i, test := range tests {
		futures = append(futures, exec.submit(test.function, test.input))
		for futures[i].running() {
			res, _ := futures[i].result(700)
			t.Log(res)
		}
	}
}

//testing results Asynchronous
func TestFuturesAsync(t *testing.T) {
	var tests = []struct {
		input    []int
		function func(int) string
	}{{[]int{1, 2, 3, 4}, pow_2},
		{[]int{11, 12, 13, 14}, pow_3},
		{[]int{21, 22, 23, 24}, pow_2}}
	exec := Executor{}
	//var futures []*Future
	var wg sync.WaitGroup
	for _, test := range tests {
		//futures = append(futures, exec.submit(test.function, test.input))
		wg.Add(1)
		go func(test_future *Future) {
			for test_future.running() {
				res, _ := test_future.result(1000)
				t.Log(res)
			}
			wg.Done()
		}(exec.submit(test.function, test.input))
	}
	wg.Wait()
}

//Asynchronous results, cancelling in same go routine
func TestFuturesFunctionsAsync(t *testing.T) {
	var tests = []struct {
		input    []int
		function func(int) string
	}{{[]int{1, 2, 3, 4}, pow_2},
		{[]int{11, 12, 13, 14}, pow_3},
		{[]int{21, 22, 23, 24}, pow_2}}
	exec := Executor{}
	var wg sync.WaitGroup
	var futures []*Future
	for i, test := range tests {
		futures = append(futures, exec.submit(test.function, test.input))
		wg.Add(1)
		go func(test_future *Future, index int, wg *sync.WaitGroup) {
			for callcount := 0; callcount < 5; callcount++ {
				//cancel future execution
				if [2]bool{true, false}[rand.Intn(2)] {
					t.Logf("CANCELLING FUTURE#%d\n", index)
					if !test_future.cancel() {
						t.Logf("FUTURE#%d ALREADY STOPPED!!\n", index)
					}
				}
				//Check if future is running
				if [2]bool{true, false}[rand.Intn(2)] {
					if test_future.running() {
						t.Logf("FUTURE#%d RUNNING...\n", index)
					} else {
						t.Logf("FUTURE#%d STOPPED RUNNING\n", index)
					}
				}
				//fetch result
				res, err := test_future.result(700)
				if err != nil {
					t.Logf("FUTURE#%d-> %v\n", index, err)
				} else {
					t.Logf("FUTURE#%d-> %q\n", index, res)
				}
			}
			wg.Done()
		}(futures[i], i, &wg)
	}
	wg.Wait()
}

//Asynchronous results, cancelling using separate go routine
func TestFuturesCancelAsync(t *testing.T) {
	var tests = []struct {
		input    []int
		function func(int) string
	}{{[]int{1, 2, 3, 4}, pow_2},
		{[]int{11, 12, 13, 14}, pow_3},
		{[]int{21, 22, 23, 24}, pow_2}}
	exec := Executor{}
	var wg sync.WaitGroup
	var futures []*Future
	wg.Add(6)
	for i, test := range tests {
		futures = append(futures, exec.submit(test.function, test.input))

		go func(test_future *Future, index int, wg *sync.WaitGroup) {
			defer wg.Done()
			for callcount := 0; callcount < 5; callcount++ {
				//fetch result
				res, err := test_future.result(1000)
				if err != nil {
					t.Logf("FUTURE#%d-> %v\n", index, err)
				} else {
					t.Logf("FUTURE#%d-> %q\n", index, res)
				}
			}
		}(futures[i], i, &wg)
	}
	//starting goroutine for cancelling and checking status after 1 sec
	time.Sleep(time.Millisecond * 1000)

	for i := range tests {
		go func(test_future *Future, index int, wg *sync.WaitGroup) {
			defer wg.Done()
			for callcount := 0; callcount < 2; callcount++ {
				//check running status
				if test_future.running() {
					t.Logf("FUTURE#%d STATUS = RUNNING\n", index)
				} else {
					t.Logf("FUTURE#%d STATUS = STOPPED\n", index)
				}
				//cancel execution, report if already stopped
				t.Logf("CANCELLING FUTURE#%d = %s\n", index, map[bool]string{true: "CANCELLED", false: "ALREADY STOPPED"}[test_future.cancel()])
			}
		}(futures[i], i, &wg)
	}
	wg.Wait()
}
