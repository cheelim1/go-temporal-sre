#!/bin/bash
# Library of functions for payment processing
# Source this file in other scripts with: source "./func_collect_payment.sh"

# Function to get IP address
myip() {
	curl -s http://icanhazip.com | \
	cut -d"/" -f 1 | \
	grep -v 127\.0 | \
	grep -v \:\:1 | \
	awk '{$1=$1};1'
}

# Function 1: Fails 50% of the time
# Input: OrderID 
# Output: Either fails immediately or prints "Step1 <OrderID>" after 1-3s
process_step1() {
    local order_id="$1"
    
    # 50% chance of failure
    if (( RANDOM % 2 )); then
        echo "FAILED: Processing Step 1 for OrderID $order_id"
        return 1
    else
        # Random sleep between 1-3 seconds
        local sleep_time=$(( ( RANDOM % 3 ) + 1 ))
        sleep "$sleep_time"
        echo "Step1 $order_id"
        return 0
    fi
}

# Function 2: Fails 20% of the time (timeout) or prints gibberish 20% of the time
# Input: OrderID
# Output: "Step2 <OrderID>" after 0-2s, or random gibberish, or timeout error
process_step2() {
    local order_id="$1"
    
    # Generate a random number 0-99
    local rand=$((RANDOM % 100))
    
    # 20% chance of timeout failure
    if (( rand < 20 )); then
        # Timeout after 3-5 seconds
        local timeout=$(( ( RANDOM % 3 ) + 3 ))
        sleep "$timeout"
        echo "ERROR: Timeout occurred after ${timeout}s for OrderID $order_id"
        return 2
    
    # 20% chance of gibberish failure  
    elif (( rand < 40 )); then
        # Output random gibberish
        echo "$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1) $order_id ERROR!"
        return 3
    
    # 60% chance of success
    else
        # Random sleep between 0-2 seconds
        local sleep_time=$(( RANDOM % 3 ))
        sleep "$sleep_time"
        echo "Step2 $order_id"
        return 0
    fi
}