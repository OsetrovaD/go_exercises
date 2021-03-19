package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// сюда писать код
const multiHashIterationAmount = 6

type sortable []string

type stepHolder struct {
	workers []job
	wg      *sync.WaitGroup
}

//I don't really need this. I just wanted to implement a pipeline
type pipeline interface {
	add(worker job) pipeline
	build()
}

func (values sortable) sortFunction() func(i, j int) bool {
	return func(i, j int) bool {
		return values[i] < values[j]
	}
}

func create(wg *sync.WaitGroup) pipeline {
	return &stepHolder{[]job{}, wg}
}

func (s *stepHolder) add(worker job) pipeline {
	s.workers = append(s.workers, worker)
	return s
}

func (s *stepHolder) build() {
	in := make(chan interface{}, MaxInputDataLen)
	for _, worker := range s.workers {
		s.wg.Add(1)
		out := make(chan interface{}, MaxInputDataLen)
		go func(worker job, in chan interface{}, out chan interface{}) {
			defer s.wg.Done()
			defer close(out)
			worker(in, out)
		}(worker, in, out)
		in = out
	}
}

func SingleHash(in, out chan interface{}) {
	singleHashWG := &sync.WaitGroup{}
	singleHashMutex := &sync.Mutex{}
	for data := range in {
		dataInt := data.(int)
		singleHashWG.Add(1)
		go calculateSingleHash(dataInt, out, singleHashMutex, singleHashWG)
	}
	singleHashWG.Wait()
}

func calculateSingleHash(data int, out chan interface{}, mutex *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()
	stringValue := strconv.Itoa(data)
	firstPartChannel := make(chan string)
	secondPartChannel := make(chan string)
	go calculateFirstSingleHashPart(stringValue, firstPartChannel)
	go calculateSecondSingleHashPart(stringValue, secondPartChannel, mutex)
	s := <-firstPartChannel + "~" + <-secondPartChannel
	close(firstPartChannel)
	close(secondPartChannel)
	fmt.Println(s)
	out <- s
}

func calculateFirstSingleHashPart(data string, resultChannel chan string) {
	resultChannel <- DataSignerCrc32(data)
}

func calculateSecondSingleHashPart(data string, resultChannel chan string, mutex *sync.Mutex) {
	mutex.Lock()
	md5 := DataSignerMd5(data)
	mutex.Unlock()
	resultChannel <- DataSignerCrc32(md5)
}

func MultiHash(in, out chan interface{}) {
	multiHashWG := &sync.WaitGroup{}
	for data := range in {
		value := data.(string)
		multiHashWG.Add(1)
		go calculateMultiHash(multiHashWG, value, out)
	}
	multiHashWG.Wait()
}

func calculateMultiHash(wg *sync.WaitGroup, data string, out chan interface{}) {
	defer wg.Done()
	values := make([]string, multiHashIterationAmount)
	innerWG := &sync.WaitGroup{}
	mutex := &sync.Mutex{}
	for i := 0; i < multiHashIterationAmount; i++ {
		innerWG.Add(1)
		go calculateMultiHashSingleValue(innerWG, i, data, &values, mutex)
	}
	innerWG.Wait()
	join := strings.Join(values, "")
	fmt.Println("value:", data, "multiHash:", join)
	out <- join
}

func calculateMultiHashSingleValue(wg *sync.WaitGroup, index int, data string, values *[]string, mutex *sync.Mutex) {
	defer wg.Done()
	crc32Value := DataSignerCrc32(strconv.Itoa(index) + data)
	mutex.Lock()
	(*values)[index] = crc32Value
	mutex.Unlock()
}

func CombineResults(in, out chan interface{}) {
	hashedValues := sortable([]string{})
	for data := range in {
		hashedValues = append(hashedValues, data.(string))
	}
	sort.Slice(hashedValues, hashedValues.sortFunction())
	join := strings.Join(hashedValues, "_")
	fmt.Println("result:", join)
	out <- join
}

func ExecutePipeline(jobs ...job) {
	wg := &sync.WaitGroup{}
	p := create(wg)
	for _, worker := range jobs {
		p.add(worker)
	}
	p.build()
	wg.Wait()
}
