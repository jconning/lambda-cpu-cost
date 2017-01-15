package main

// Author: Jim Conning, Jan 2017

import (
	"fmt"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"sync"
	"errors"
	"sort"
	"flag"
)

var maxPrime int
var numExecutions int
var numLoops int
var maxConcurrency int
func init() {
	flag.IntVar(&maxPrime, "max", 1000000, "maximum number to search for primes (<=2M to not cause out of memory in the lowest Lambda memory setting)")
	flag.IntVar(&numExecutions, "execs", 20, "number of times to execute the Lambda function")
	flag.IntVar(&numLoops, "loops", 1, "number of times to repeat the search for primes (without consuming additional memory)")
	flag.IntVar(&maxConcurrency, "conc", 100, "limit of concurrently running Lambda functions")
	flag.Parse()
}

type execution struct {
	DurationSeconds float64
	memory int
}

var lambdaFunctions = map[int]string{
	128:"https://652zahuut4.execute-api.us-west-2.amazonaws.com/prod/eratosthenes-128",
	256:"https://652zahuut4.execute-api.us-west-2.amazonaws.com/prod/eratosthenes-256",
	512:"https://652zahuut4.execute-api.us-west-2.amazonaws.com/prod/eratosthenes-512",
	1024:"https://652zahuut4.execute-api.us-west-2.amazonaws.com/prod/eratosthenes-1024",
}

// AWS Lambda pricing in USD as of Jan 2017
var costPerRequest float64 = 0.0000002
var costPerGbSeconds float64 = 0.00001667 

func triggerLambda(url string, mem int, max int, loops int) (execution, error) {
	var e execution
	e.memory = mem

	resp, err := http.Get(fmt.Sprintf("%s?max=%d&loops=%d", url, max, loops));
	if err != nil {
		return e, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return e, errors.New(fmt.Sprintf("status code: %d", resp.StatusCode))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return e, err
	}
	err = json.Unmarshal(body, &e)

	return e, nil
}

func main() {
	var wg sync.WaitGroup
	var lambdaErrors int
	var tokens = make(chan struct{}, maxConcurrency) // counting semaphore used to enforce a concurrency limit on calls to Lambda
	executions := make(chan execution)

	fmt.Printf("Triggering %d Lambda functions %d times each all in parallel...\n", len(lambdaFunctions), numExecutions)
	for mem, url := range lambdaFunctions {
		for f := 0; f < numExecutions; f++ {
			wg.Add(1)
			go func(u string, m int) {
				defer wg.Done()
				tokens <- struct{}{} // acquire a token
				e, err := triggerLambda(u, m, maxPrime, numLoops)
				<-tokens // release the token
				if err != nil {
					fmt.Println(err)
					lambdaErrors++
				}
				executions <- e
			}(url, mem)
		}
	}

	// Wait for all goroutines to finish their work
	go func() {
		wg.Wait()
		close(executions)
	}()

	var totalDurations map[int]float64 = make(map[int]float64)
	var executionCounts map[int]int = make(map[int]int)

	// Pull all execution results from the channel
	for e := range executions {
		if e.DurationSeconds > 0 { // only count executions that didn't error
			totalDurations[e.memory] += e.DurationSeconds
			executionCounts[e.memory]++
		}
	}

	// Sort the various lambda function memory sizes for pretty printing
	var memories []int
	for mem, _ := range lambdaFunctions {
		memories = append(memories, mem)
	}
	sort.Ints(memories)

	fmt.Printf("Number of lambda executions returning errors: %d\n", lambdaErrors)
	fmt.Println("Stats for each Lambda function by Lambda memory allocation:")
	var totalCost float64
	for _, mem := range memories {
		cost := float64(executionCounts[mem]) * costPerRequest + 
			(float64(mem)/float64(1024)) * float64(totalDurations[mem]) * costPerGbSeconds // convert duration to GB-seconds
		totalCost += cost
		fmt.Printf("  %dmb %fsec(avg) $%f(total) to calculate %d times all prime numbers <=%d\n", 
			mem, totalDurations[mem]/float64(executionCounts[mem]), cost, executionCounts[mem], maxPrime)
	}
	fmt.Printf("Total cost of this test run: %f\n", totalCost);
}

