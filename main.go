package main

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

func GenerateRandomSlice(n, min, max int) []int {

	rand.Seed(time.Now().UnixNano()) // Seed the random number generator
	slice := make([]int, n)          // Create a slice of length n
	for i := range slice {
		slice[i] = rand.Intn(max-min) + min // Generate a random number in [min, max)
	}
	return slice
}

func raceConditionSum(x []int, sum *int, wg *sync.WaitGroup) {

	for _, v := range x {
		*sum += v
	}
	wg.Done()
}

func mutexSum(x []int, sum *int, mu *sync.Mutex, wg *sync.WaitGroup) {

	result := 0
	for _, v := range x {
		result += v
	}
	mu.Lock()
	*sum += result
	mu.Unlock()
	wg.Done()
}

func atomicSum(x []int, sum *int64, wg *sync.WaitGroup) {

	for _, v := range x {
		atomic.AddInt64(sum, int64(v)) // maximum pressure on shared resource for demo purposes
	}
	wg.Done()
}

func channelSum(x []int, ch chan int) {

	result := 0
	for _, v := range x {
		result += v
	}
	ch <- result
}

func serialSum(x []int) int {

	result := 0
	for _, v := range x {
		result += v
	}
	return result
}

func main() {

	var wg sync.WaitGroup
	data := GenerateRandomSlice(50000000, 0, 10)
	nThreads := 8
	blockSize := len(data) / nThreads

	// serialSum
	fmt.Println("serialSum: ", serialSum(data))

	//raceConditionSum
	wg.Add(nThreads)
	raceConditionResult := 0
	for i := 0; i < nThreads; i++ {
		blockSize := len(data) / nThreads
		if i != nThreads-1 {
			go raceConditionSum(data[blockSize*i:blockSize*(i+1)], &raceConditionResult, &wg)
		} else {
			go raceConditionSum(data[blockSize*i:], &raceConditionResult, &wg)
		}
	}
	wg.Wait()
	fmt.Println("raceConditionSum: ", raceConditionResult)

	// mutexSum
	var mu sync.Mutex
	wg.Add(nThreads)
	mutexResult := 0 // shared
	for i := 0; i < nThreads; i++ {
		if i != nThreads-1 {
			go mutexSum(data[blockSize*i:blockSize*(i+1)], &mutexResult, &mu, &wg)
		} else {
			go mutexSum(data[blockSize*i:], &mutexResult, &mu, &wg)
		}
	}
	wg.Wait()
	fmt.Println("mutexSum: ", mutexResult)

	//atomicSum
	wg.Add(nThreads)
	atomicResult := int64(0)
	for i := 0; i < nThreads; i++ {
		if i != nThreads-1 {
			go atomicSum(data[blockSize*i:blockSize*(i+1)], &atomicResult, &wg)
		} else {
			go atomicSum(data[blockSize*i:], &atomicResult, &wg)
		}
	}
	wg.Wait()
	fmt.Println("atomicSum: ", atomicResult)

	//channelSum
	ch := make(chan int, nThreads)
	channelResult := 0
	for i := 0; i < nThreads; i++ {
		if i != nThreads-1 {
			go channelSum(data[blockSize*i:blockSize*(i+1)], ch)
		} else {
			go channelSum(data[blockSize*i:], ch)
		}
	}
	//collect results from channels
	for i := 0; i < nThreads; i++ { // blocking until nThreads results has been received
		channelResult += <-ch
	}
	close(ch)
	fmt.Println("channelSum: ", channelResult)

}
