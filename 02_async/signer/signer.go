package main

import (
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
)

func worker(function job, in, out chan interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	function(in, out)
	close(out)
}

func ExecutePipeline(funcs ...job) {
	wg := &sync.WaitGroup{}
	in := make(chan interface{})
	for _, function := range funcs {
		out := make(chan interface{})
		wg.Add(1)
		go worker(function, in, out, wg)
		in = out
	}
	wg.Wait()
}

func SingleHashWorker1(chanstr1 chan string, val string,) {
	chanstr1 <- DataSignerCrc32(val)
}

func SingleHashWorker2(chanstr2 chan string, str2 string) {
	chanstr2 <- DataSignerCrc32(str2)
}

func SingleHashWorker3( out chan interface{}, valStr, str2 string, wg *sync.WaitGroup) {
	defer wg.Done()
	chanstr1 := make (chan string)
	chanstr2 := make (chan string)
	go SingleHashWorker1(chanstr1, valStr)
	go SingleHashWorker2(chanstr2, str2)
	str1 := <- chanstr1
	str2 = <- chanstr2
	out <- str1 + "~" + str2
}

func SingleHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	for val := range in {
		valStr := strconv.Itoa(val.(int))
		str2 := DataSignerMd5(valStr)
		wg.Add(1)
		go SingleHashWorker3(out, valStr, str2, wg)
	}
	wg.Wait()
}

func MultiHashWorker(betweenResult map[int32](map[int32]string), indexOfInputData, iteration int32, valStr string, mu *sync.RWMutex, wg *sync.WaitGroup) {
	defer wg.Done()
	curStr := DataSignerCrc32(valStr)
	mu.Lock()
	betweenResult[indexOfInputData][iteration] = curStr
	mu.Unlock()
}

func MultiHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	mu := &sync.RWMutex{}
	betweenResult := make(map[int32](map[int32]string))
	var j int32 // индекс, формирующий результат
	for val := range in {
		valStr := val.(string)
		var i int32 // итерация мультихэша
		mu.Lock()
		betweenResult[j] = make(map[int32]string)
		mu.Unlock()
		for i < 6 {
			th := strconv.Itoa(int(i))
			th += valStr
			wg.Add(1)
			go MultiHashWorker(betweenResult, j, i, th, mu, wg)
			atomic.AddInt32(&i, 1)
		}
		atomic.AddInt32(&j, 1)
	}
	wg.Wait()
	var k int32
	for ; k < j; k++ {
		var str string
		var i int32
		for ; i < 6; i++ {
			str += betweenResult[k][i]
		}
		out <- str
	}
}

func CombineResults(in, out chan interface{}) {
	sliceOfResults := make([]string, len(in)) 
	var result string
	for buf := range in {
		bufStr := buf.(string)
		sliceOfResults = append(sliceOfResults, bufStr)
	}
	sort.Strings(sliceOfResults)
	for j, str := range sliceOfResults {
		if j != 0 {
			result += "_"
		}
		result += str
	}
	out <- result
}
