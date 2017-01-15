from __future__ import print_function
from timeit import default_timer as timer

import json
import datetime

print('Loading function')  


def eratosthenes(n):
    sieve = [ True for i in range(n+1) ]
    def markOff(pv):
        for i in range(pv+pv, n+1, pv):
            sieve[i] = False
    markOff(2)
    for i in range(3, n+1):
        if sieve[i]:
            markOff(i)
    return [ i for i in range(1, n+1) if sieve[i] ]

def lambda_handler(event, context):
    start = timer()
    #print("Received event: " + json.dumps(event, indent=2))

    maxPrime = int(event['queryStringParameters']['max'])
    numLoops = int(event['queryStringParameters']['loops'])

    print("looping " + str(numLoops) + " time(s)")
    for loop in range (0, numLoops):
        primes = eratosthenes(maxPrime)
        print("Highest 3 primes: " + str(primes.pop()) + ", " + str(primes.pop()) + ", " + str(primes.pop()))

    durationSeconds = timer() - start
    return {"statusCode": 200, \
        "headers": {"Content-Type": "application/json"}, \
        "body": "{\"durationSeconds\": " + str(durationSeconds) + \
        ", \"max\": " + str(maxPrime) + ", \"loops\": " + str(numLoops) + "}"}

