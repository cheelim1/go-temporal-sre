#!/bin/bash
##
## Traditional Payment Collection Script
## This script processes a batch of OrderIDs using the single_payment_collection.sh script

# Always use strict mode for reliable error handling
# URL: http://redsymbol.net/articles/unofficial-bash-strict-mode/
set -euo pipefail
IFS=$'\n\t'

# Define color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Source the functions library
SOURCE_DIR="$(dirname "${BASH_SOURCE[0]}")"
if ! source "$SOURCE_DIR/func_collect_payment.sh"; then
    echo -e "${RED}ERROR: Cannot source functions library${NC}" >&2
    exit 1
fi

# List of OrderIDs to process (from our previous random test)
ORDER_IDS=(
    7307
    5493
    7387
    2614
    5999
    3078
    8577
    5479
    6606
    8448
)

# Summary counters
TOTAL_COUNT=0
SUCCESS_COUNT=0
FAIL_COUNT=0

# We use set +e to prevent the loop from exiting if the script fails
set +e
myIP=$(myip)
set -e
echo -e "\n${YELLOW}Script Run from IP: ${myIP} ${NC}"

echo "======================================================"
echo -e "${YELLOW}Starting batch processing of ${#ORDER_IDS[@]} OrderIDs${NC}"
echo "======================================================"

# Process each OrderID in the list
for order_id in "${ORDER_IDS[@]}"; do
    TOTAL_COUNT=$((TOTAL_COUNT + 1))
    echo -e "\n${YELLOW}Processing OrderID: $order_id (${TOTAL_COUNT}/${#ORDER_IDS[@]})${NC}"
    
    # Record start time
    start_time=$(date +%s)
    
    # Call the single payment collection script and capture output
    # We use set +e to prevent the loop from exiting if the script fails
    set +e
    output=$($SOURCE_DIR/single_payment_collection.sh "$order_id" 2>&1)
    exit_code=$?
    set -e
    
    # Record end time and calculate duration
    end_time=$(date +%s)
    duration=$((end_time - start_time))
    
    # Display result based on exit code
    if [[ $exit_code -eq 0 ]]; then
        SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
        echo -e "${GREEN}SUCCESS: OrderID $order_id processed successfully in ${duration}s${NC}"
        echo "$output" | sed 's/^/  /'
    else
        FAIL_COUNT=$((FAIL_COUNT + 1))
        echo -e "${RED}ERROR: OrderID $order_id failed with exit code $exit_code in ${duration}s${NC}"
        echo "$output" | sed 's/^/  /'
    fi
    
    echo "------------------------------------------------------"
done

# Print summary
echo -e "\n${YELLOW}Batch Processing Summary:${NC}"
echo "======================================================"
echo -e "Total OrderIDs processed: ${TOTAL_COUNT}"
echo -e "${GREEN}Successful: ${SUCCESS_COUNT}${NC}"
echo -e "${RED}Failed: ${FAIL_COUNT}${NC}"
echo -e "Success rate: $(( (SUCCESS_COUNT * 100) / TOTAL_COUNT ))%"
echo "======================================================"

# Always exit with success since we expect some failures
exit 0

