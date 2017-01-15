# lambda-cpu-cost
Test harness that invokes cpu-intensive Lambda functions at various memory levels and compares performance and cost.
## What is it?
This is a test harness that invokes a CPU-bound Lambda function concurrently so that execution time and cost can be measured. The maximum concurrency is configurable so that the function can be executed many times in parallel without causing AWS to throttle it.
### Sieve of Eratosthenes - Python
The Labmda function implements the Sieve of Eratosthenes algorithm for calculating prime numbers.  All prime numbers less than or equal to n are calculated, where n is configurable.  The Sieve consumes memory and the function will loop a configurable number of times to accumulate CPU usage within a constrained memory footprint.  The execution time is mesaured by the function and output to the test harness.

Multiple variants of the function are supported so that the performance of various Lambda memory settings can be compared with each other. The smallest Lambda memory setting is 128MB.  By default, four functions are defined: 128MB, 256MB, 512MB and 1024MB.  Every time the memory size doubles, Lambda provides double the CPU, making it straight forward to compare the execution time for calculating prime numbers across various memory settings.
### Test Harness - Golang
The test harness is written in Go and uses channels to control concurrency.  A counting semaphore is employed to limit the parallelism to a maximum concurrency limit.  A sync.WaitGroup counter is used to ensure all goroutines finish before the program exits.

The same function code, instantiated as multiple Lambda functions each with a distinct level of memory, are executed the same number of times.  Since each Lambda level of memory corresponds to a distinct amount of CPU, we can measure the execution time of calculating primes for multiple levels of CPU.  And since Lambda charges per GB-second of execution time, we can show the cost of calculating primes at various Lambda memory configurations.  We can answer the question: if we double the CPU, will that cut the time in half while still costing the same?
## Learnings
## How to set it up
### Lambda functions
1. In the AWS Console, visit the Lambda section and click the **Create a Lambda function** button.
1. Under **Select blueprint** choose **Blank Function**.
1. Under **Configure triggers**, click the dashed rounded square and choose **API Gateway**.
  1. Choose a name for your new API, such as **Eratosthenes**.
  1. Leave the **Deployment stage** as **prod**.
  1. Set **Security** to **Open**.  This will enable your API to be invoked via HTTP without security credentials.
  1. Click **Next**.
1. Under **Configure function**:
  1. Name the function: eratosthenes-128
    * The "128" means 128mb of memory.
  1. For the **Runtime** choose **Python 2.7**.
### Test harness
## How to run it
