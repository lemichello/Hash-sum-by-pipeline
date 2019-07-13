package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

func ExecutePipeline(jobs ...job) {
	var prevCh chan interface{}

	start := time.Now()
	wg := &sync.WaitGroup{}

	for _, j := range jobs {
		wg.Add(1)

		outCh := make(chan interface{}, 1)

		go worker(j, prevCh, outCh, wg)

		prevCh = outCh
	}

	wg.Wait()

	fmt.Println("Total time :", time.Since(start))
}

func worker(job job, in, out chan interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(out)

	job(in, out)
}

func SingleHash(in, out chan interface{}) {
	mx := &sync.Mutex{}
	wg := &sync.WaitGroup{}

	for data := range in {
		wg.Add(1)

		go singleHashWorker(data, out, mx, wg)
	}

	wg.Wait()
}

func singleHashWorker(inData interface{}, out chan interface{}, mx *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()

	fmt.Println(inData, "SingleHash data", inData)

	data := strconv.Itoa(inData.(int))

	// Avoiding overheat by mutexes.
	mx.Lock()
	md5 := DataSignerMd5(data)
	mx.Unlock()

	fmt.Println(data, "SingleHash md5(data)", md5)

	crc32Md5Ch := make(chan string)

	go calculateCrc32Async(md5, crc32Md5Ch)

	crc32Res := DataSignerCrc32(data)
	crc32Md5Res := <-crc32Md5Ch

	fmt.Println(data, "SingleHash crc32(md5(data))", crc32Md5Res)
	fmt.Println(data, "SingleHash crc32(data)", crc32Res)
	fmt.Println(data, "SingleHash result", crc32Res+"~"+crc32Md5Res)

	out <- crc32Res + "~" + crc32Md5Res
}

func calculateCrc32Async(data string, out chan string) {
	out <- DataSignerCrc32(data)
}

func MultiHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}

	for data := range in {
		wg.Add(1)
		go multiHashWorker(data, out, wg)
	}

	wg.Wait()
}

func multiHashWorker(data interface{}, out chan interface{}, outerWg *sync.WaitGroup) {
	defer outerWg.Done()

	hashResults := make([]string, 6)

	wg := &sync.WaitGroup{}
	mx := &sync.Mutex{}

	for i := 0; i < 6; i++ {
		wg.Add(1)

		go func(index int, wg *sync.WaitGroup, mx *sync.Mutex) {
			defer wg.Done()

			calculated := DataSignerCrc32(strconv.Itoa(index) + data.(string))

			mx.Lock()

			hashResults[index] = calculated
			fmt.Println(data, "MultiHash: crc32(th+step1))", index, calculated)

			mx.Unlock()

		}(i, wg, mx)
	}

	wg.Wait()

	multiHashResult := strings.Join(hashResults, "")

	fmt.Println(data, "MultiHash result:", multiHashResult)

	out <- multiHashResult
}

func CombineResults(in, out chan interface{}) {
	var res []string

	for val := range in {
		res = append(res, val.(string))
	}

	sort.Strings(res)

	combinedHash := strings.Join(res, "_")

	fmt.Println("Combine results\n", combinedHash)

	out <- combinedHash
}

func main() {
	inputData := []int{1, 2, 3, 3, 5, 6}

	hashSignJobs := []job{
		job(func(in, out chan interface{}) {
			for _, fibNum := range inputData {
				out <- fibNum
			}
		}),
		job(SingleHash),
		job(MultiHash),
		job(CombineResults),
		job(func(in, out chan interface{}) {
			// Custom functions.
		}),
	}

	ExecutePipeline(hashSignJobs...)
}
