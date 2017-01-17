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
### API Gateway Timeout
While Lambda offers a 5 minute timeout for functions, API Gateway only offers 30 seconds.  So, any use of Lambda by the API Gateway is limited to functions running 30 seconds or less.
### Variation of Lambda function execution times
The same CPU bound code, run with the same Lambda memory/cpu configuration, runs different amounts of time when run again and again.  For example, calculating all prime numbers <=1M in the 128mb Lambda setting, run ten times sequentially (not parallel), took the following number of seconds:
17.9 17.1 14.8 16.6 16.1 16.5 16.5 15.1 15.1 15.4

This raises an interesting point.  Lambda costs different amounts to perform the same work at different times.  An interesting side effect of serverless computing.
### The effective cost of different memory levels is similar
Each step-up in Lambda memory carries with it a corresponding step-up in CPU power.  The results below show the effective cost for performing the same workload (calculating all prime numbers <=1M) is roughly the same for each memory level.  If your workload requires significant CPU and is CPU bound, it appears to be worthwhile to choose a higher memory configuration.  Even if you don't need the memory, the additional CPU that comes with it will get the job done faster for the same price.  Your lambda function will be much faster without additional cost.  This exacerbates the benefits of serverless computing because you get the same serverless benefit, pay the same amount, and get your work done in a fraction of the time.
```
Stats for each Lambda function by Lambda memory allocation:
 128mb 11.722965sec(avg) $0.024628(total) to calculate 1000 times all prime numbers <=1000000
 256mb 6.678945sec(avg) $0.028035(total) to calculate 1000 times all prime numbers <=1000000
 512mb 3.194954sec(avg) $0.026830(total) to calculate 1000 times all prime numbers <=1000000
 1024mb 1.465984sec(avg) $0.024638(total) to calculate 1000 times all prime numbers <=1000000
Total cost of this test run: $0.104130
```
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
### Calculate even higher prime numbers
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
All we need to do is set the number of executions (-exec) to a high number so that Lambda has plenty of function calls to make.  We'll leave the concurrency limit to the default of 80 to ensure we don't get throttled by Lambda and to leave some breathing room.  We'll leave "n", our maximum prime number (-max), to the default of 1M to ensure the eratosthenes-128 function doesn't time out at all (even though it quite often will calculate all primes in about 12 seconds, I have seen it timeout (30 second timeout) from time to time.  This test should generally not produce any errors, although it is not unheard of for eratosthenes-128 to timeout a couple times when ran with a "-execs" value of 1000.
<br>`go run main.go -execs 1000`

Output (note that eratosthenes-128 errored out twice -- these were timeouts):
```
Number of lambda executions returning errors: 2
Stats for each Lambda function by Lambda memory allocation:
 128mb 21.118205sec(avg) $0.044117(total) to calculate 998 times all prime numbers <=1000000
 256mb 10.268854sec(avg) $0.042995(total) to calculate 1000 times all prime numbers <=1000000
 512mb 3.164584sec(avg) $0.026577(total) to calculate 1000 times all prime numbers <=1000000
 1024mb 1.433012sec(avg) $0.024088(total) to calculate 1000 times all prime numbers <=1000000
Total cost of this test run: $0.137777
```
## Logging
### Lambda logging
You can see the Lambda function logging directly by testing it in the AWS Console.  On the Lambda section of the console, choose the desired function and click the **Test** button.  Be sure the test event is configured as specified above.  Both the function output and the logging will appear in the console.
### API Gateway logging
When invoked via API Gateway, the logging for both Lambda and API Gateway are configured in the API Gateway section of the AWS Console. Choose the Eratosthenes API in the API Gateway then in the left nav click **Stages**, then in the Stages pane click **prod**.  Then in the Settings tab, under "CloudWatch Settings", check the box **Enable CloudWatch Logs** and click **Save Changes**.  Once you make this settings change you may need to wait a few minutes before the logs start showing up.

It is not necessary to check the checkbox **Log full requests/responses data** in order to see the error messages.  You can also set the log level to ERROR.

Turning on **Log full requests/responses data** results in too much verbose logging.  What I did was make a short test case that will force an error, then turn on **Log full requests/responses data** long enough to capture useful information.

To view your logs, visit the CloudWatch section of the AWS Console and choose **Logs** in the left nav.  Look for the Eratosthenes log group.
### CloudWatch Logging example: timeout
This is what appears in the CloudWatch log when the function is terminated due to an API Gateway timeout.  This line appears regardless of whether **Log full requests/responses data** is turned on.
```
Execution failed due to a timeout error
```
The API Gateway returns **status code 504** when the error is due to an API Gateway timeout.
### CloudWatch logging example: excessive memory consumption
This is what appears in the CloudWatch log when the function encounters an error due to too much memory consumption.  The last line appears regardless of whether **Log full requests/responses data** is turned on, but the rest of it appears only when that is turned on.
```
Endpoint response body before transformations:
{
    "errorMessage": "RequestId: d27362eb-dc3a-11e6-ad1e-d97a8717a019 Process exited before completing request"
}
Execution failed due to configuration error: Malformed Lambda proxy response
```
What's happening here is that Lambda is returning json output that the API Gateway doesn't understand.  The API Gateway expects to receive json of the following format, and if the format differs it simply says "Malformed Lambda proxy response" and returns a 502.
```
{
"statusCode": 200, 
"headers": {"Content-Type": "application/json"},
"body": "body_text_goes_here"
}
```
This is what the output from the Eratosthenes Lambda function typically looks like:
```
{
"statusCode": 200,
"headers": {"Content-Type": "application/json"},
"body": "{\"durationSeconds\": 8.345243, \"max\": 1000000, \"loops\": 1}"
}
```
The API Gateway returns **status code 502** when the error is due excessive memory consumption within the Lambda function
