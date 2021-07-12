package main

import (
	"fmt"
	"time"
)

//simple functions i square
func pow_2(i int) string {
	time.Sleep(time.Millisecond * 400)
	return fmt.Sprintf("%d square: %d", i, i*i)
}

//simple functions i cube
func pow_3(i int) string {
	time.Sleep(time.Millisecond * 600)
	return fmt.Sprintf("%d cube: %d", i, i*i*i)
}

//Future struct
type Future struct {
	resultChannel    chan string //channel containing results
	stopChannel      chan bool   //channel for sending stop signal
	channelCancelled bool        //channel to check if execution successfully cancelled
	channelStopped   bool        //channel stopped or not stopped
}

//execute function in go routine and send result to future's resultchannel
func gofunc(f func(int) string, input []int, fut *Future) {
	go func() {
		for _, value := range input {
			select {
			case fut.channelStopped = <-fut.stopChannel:
				fut.channelCancelled = fut.channelStopped
				return
			case fut.resultChannel <- f(value):
			}
		}
		fut.channelStopped = true
	}()
}

//get result from future object (timeout in millisecond)
func (fut *Future) result(timeout_ms int) string {
	timeout := time.After(time.Millisecond * time.Duration(timeout_ms))
	for {
		if fut.running() {
			select {
			case s := <-fut.resultChannel:
				return s
			case <-timeout:
				panic("Timeout Error")
			}

		} else {
			//goroutines asleep error if execution stopped
			return <-fut.resultChannel
		}
	}
}

//stop the go routine if not stopped
//return false if already stopped
func (fut *Future) cancel() bool {
	if fut.done() {
		return false
	}
	fut.stopChannel <- true
	return true
}

//check if execution stopped
func (fut *Future) cancelled() bool {
	return fut.channelCancelled
}

//check if executing
func (fut *Future) running() bool {
	return !fut.channelStopped
}

//check if cancelled or finished running
func (fut *Future) done() bool {
	return fut.channelStopped
}

//executor for submitting function and input
//also contains futures created by submission
type Executor struct {
	Futures []*Future
}

//executor submit function, executes the function on input, returns the future instance
func (exec *Executor) submit(f func(int) string, input []int) *Future {
	fut := Future{make(chan string), make(chan bool), false, false}
	gofunc(f, input, &fut)
	exec.Futures = append(exec.Futures, &fut)
	return &fut
}

func main() {
	inp_list1 := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	inp_list2 := []int{14, 15, 16, 17, 18, 19, 20}
	exec := Executor{}
	//starting two go routines from executor
	future1 := exec.submit(pow_3, inp_list1)
	future2 := exec.submit(pow_2, inp_list2)
	//checking running status of future1 and future2
	fmt.Printf("check running future1: %t, future2: %t", future1.running(), future2.running())
	for i := 0; i < 10; i++ {
		if future1.running() {
			//get result if future1 is running
			fmt.Println(future1.result(700))
		}
		if future2.running() {
			//get result if future2 is running
			fmt.Println(exec.Futures[1].result(700))
		}
		if i == 5 {
			fmt.Println("cancelling future1:", future1.cancel())
			fmt.Println("checkcancelled future1:", future1.cancelled())
		}
		if i == 7 {
			fmt.Println("cancelling future2:", future2.cancel())
			fmt.Println("checkcancelled future2:", future2.cancelled())
		}
	}
	fmt.Printf("check running future1: %t, future2: %t\n", future1.running(), future2.running())
	fmt.Printf("check cancelled future1: %t, future2: %t\n", future1.cancelled(), future2.cancelled())
	fmt.Printf("check done future1: %t, future2: %t\n", future1.done(), future2.done())
}
