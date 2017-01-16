# lambda-cpu-cost
Test harness that invokes cpu-intensive Lambda functions at various memory levels and compares performance and cost.
## What is it?
This is a test harness that invokes a CPU-bound Lambda function concurrently so that execution time and cost can be measured. The maximum concurrency is configurable so that the function can be executed many times in parallel without causing AWS to throttle it.
### Sieve of Eratosthenes - Python
The Labmda function implements the Sieve of Eratosthenes algorithm for calculating prime numbers.  All prime numbers less than or equal to n are calculated, where n is configurable.  The Sieve consumes memory and the function will loop a configurable number of times to accumulate CPU usage within a constrained memory footprint.  The execution time is measured by the function and output to the test harness.

Multiple variants of the function are supported so that the performance of various Lambda memory settings can be compared with each other. The smallest Lambda memory setting is 128MB.  By default, four functions are defined: 128MB, 256MB, 512MB and 1024MB.  Every time the memory size doubles, Lambda provides double the CPU, making it straight forward to compare the execution time for calculating prime numbers across various memory settings.
### Test Harness - Golang
The test harness is written in Go and uses a channel to control concurrency.  A counting semaphore is employed to limit the parallelism to a maximum concurrency limit.  A sync.WaitGroup counter is used to ensure all goroutines finish before the program exits.

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
1. Under **Lambda function code** for **Code entry type** choose **Edit code inline**.  Copy and paste the contents of the file eratosthenes_lambda.py (in the lambda-cpu-cost repo) into the text box.
1. Under **Lambda function handler and role**:
  1. For **Role** choose **Create new role from template(s)**.
  1. For **Role name** enter a name such as **lambdaExecutionRole**.
1. Under **Advanced settings**:
  1. For **Memory (MB)** choose **128**.
  1. For **Timeout** choose **30** sec.
    1. *Hint: API Gateway imposes a 30 second timeout so we may as well make the lambda timeout match that*
1. Choose **Next**
1. Under **Review** choose **Create function**.
1. You should see a "Congratulations!" confirmation message.  On that page you should see the full URL to your new API endpoint. <br>Example: `https://abcdefghij.execute-api.us-west-2.amazonaws.com/prod/eratosthenes-128`
<br>*where "abcdefghij" is a token unique to your API*
1. Click the **Actions** and choose **Configure test event**.
1. Under **Input test event** enter the following code:
<br>`{
  "queryStringParameters": {
      "max": 1000000,
      "loops": 1
  }
}`
1. Click **Save and test**.
1. Your function should now execute (it will take several seconds) then show the following output:
<br>`{`
<br>`"body": "{\"durationSeconds\": 5.48261809349, \"max\": 1000000, \"loops\": 1}",`
<br>`"headers": {"Content-Type": "application/json"},`
<br>`"statusCode": 200`
<br>`}`
1. It should also show the following logs:
<br>`START RequestId: 23f39528-db63-11e6-a488-013190970ce0 Version: $LATEST`
<br>`looping 1 time(s)`
<br>`Highest 3 primes: 999983, 999979, 999961`
<br>`END RequestId: 23f39528-db63-11e6-a488-013190970ce0`
<br>`REPORT RequestId: 23f39528-db63-11e6-a488-013190970ce0	Duration: 5484.17 ms	Billed Duration: 5500 ms 	Memory Size: 128 MB	Max Memory Used: 65 MB`
1. Your eratosthenes-128 function is now working!
1. Repeat the above steps to create three additional functions:
  - eratosthenes-256
  - eratosthenes-512
  - eratosthenes-1024
  - Everything will be the same except for these changes:
    - Don't create a new API name but rather select the API you created earlier named "Eratosthenes" (or whatever you named it).
    - Don't create a new role but instead select the role you created earlier named "lambdaExecutionRole" (or whatever you named it).
    - Select the memory setting that correponds to your function (256, 512, or 1024 MB).

You have now prepared all four required functions! Please proceed to setup the test harness.

### Test harness
Edit the config file (config.json) so that the function URLs correspond to the API Gateway endpoints you created above.  It should be enough to simply replace the string "abcdefghij" with the corresponding token in your endpoints, if you named them as suggested above.
```
{
    "functions": {
        "128": "https://abcdefghij.execute-api.us-west-2.amazonaws.com/prod/eratosthenes-128",
        "256": "https://abcdefghij.execute-api.us-west-2.amazonaws.com/prod/eratosthenes-256",
        "512": "https://abcdefghij.execute-api.us-west-2.amazonaws.com/prod/eratosthenes-512",
        "1024": "https://abcdefghij.execute-api.us-west-2.amazonaws.com/prod/eratosthenes-1024"
    }
}
```
## How to run it
Usage:
<br>`cd src/github.com/jconning/lambda-cpu-cost`
<br>`go run main.go`

Command line parameters:
- **-conc** *integer*: the maxiumum number of Lambda functions to run concurrently (default 80)
- **-execs** *integer*: the number of times to execute each Lambda function (default 20)
- **-loops** *integer*: the number of times to repeat the calculation of primes, without consuming additional memory (default 1)
- **-max** *integer*: this is n, and all primes will be calculated up to and including n

### Run with defaults
Running with the defaults (no command line parameters) will invoke each of the four Lambda functions 20 times, and each function will execute one loop of calculating all prime numbers up to 1M.  
<br>`go run main.go`

Sample output:
```
Number of lambda executions returning errors: 0
Stats for each Lambda function by Lambda memory allocation:
 128mb 8.592724sec(avg) $0.000362(total) to calculate 20 times all prime numbers <=1000000
 256mb 4.369844sec(avg) $0.000368(total) to calculate 20 times all prime numbers <=1000000
 512mb 2.385856sec(avg) $0.000402(total) to calculate 20 times all prime numbers <=1000000
 1024mb 1.233483sec(avg) $0.000415(total) to calculate 20 times all prime numbers <=1000000
Total cost of this test run: $0.001547
```
There should be no execution errors.  Function eratosthenes-128 ran twice as long as eratosthenes-256, which makes sense since the 256mb Lambda memory setting provides double the CPU power as the 128mb setting.
### Calculate higher prime numbers
We can increase n so that we calculate higher prime numbers, while still invoking the same number of Labmda function calls.  This will consume more time since the CPU has to work more to calculate the primes.  We set a limit of 1.5M which will result in a Sieve that still fits within the 128mb memory footprint.
<br>`go run main.go -max 1500000`
### Calculate even more prime numbers
Let's try a prime number limit of 2.5M, which will cause the Sieve of Eratosthenes to exceed the 128mb limit.  This will result in status code 502 for the eratosthenes-128 function, which is due to excess memory consumption in the Lambda function.
<br>`go run main.go -max 2500000`
### Loop twice
Let's loop the prime number calculation twice.  This will cause the Lambda function to repeat the calculation of primes without consuming more memory.  It will take twice as long to run as when we ran with the defaults.  There should be no errors from Lambda.
<br>`go run main.go -loops 2`
### Loop four times
Now let's loop four times.  This will cause eratosthenes-128 to take longer than 30 seconds, which is the maximum time allowed by the API Gateway.  You should see errors from the eratosthenes-128 function and potentially some from the eratosthenes-256 function.
<br>`go run main.go -loops 4`
### Excessive concurrency
Let's run each function 50 times and increase the concurrency to an excessively high number (9999).  This will cause Lambda to throttle the function and you will see some errors returned, assuming your Lambda concurrency cap is set to 100, which is the default for new AWS accounts.  You can increase the concurrency (-execs) a lot more if you want to see interesting errors like too many sockets open on the client.
<br>`go run main.go -execs 50 -conc 9999`
### Proper long running test
We'll set n to 1.5M so the function has plenty of primes to calculate (and will keep the number of loops to the default of one).  We'll set the number of executions (per function) to 1000 so we put a lot of executions through Lambda.  We'll keep the concurrency limit to the default of 80 so we stay within the Lambda cap of 100 with some breathing room.  This test should take about 7 minutes to run.  To run it longer, simply increase the executions (-execs).  This test shouldn't produce any errors.
<br>`go run main.go -max 1500000 -execs 1000`
