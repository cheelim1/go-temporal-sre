# SCENARIO

## Demo Idempotency with Temporal

NOTE: All unit test for functions stored in activities.go will be in activities_test.go

### Unit Test Case #1 - no-idempotency

- The TestNonIdempotency case with a OrderID - ORD-12345  will be started; as a httptest case 
that takes in the OrderID as part of its path; with HTTP POST
- The handler for this pattern will call a function DeductFee stored in activities.go; deducting USD10 for each call; after a delay between 200ms and 2s
- The account will start with USD100; with a httptest that will return the amount in account when it is called
- At about 100ms; the same call is made to teh httptest endpoint
- Block until both calls return a result
- Assert at the end that both are OK with HTTP 200; record down the tie each take and the overall time needed
- Assert that the account will now have USD80 since 2 calls will deduct UUSD10 each time

