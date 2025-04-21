# README

## Objective

To show common batch issues and how Temporal can play a part in simplifying things.

## Idempotency

First start by showing common activities (3rd party API, DB access, random actions) are usually not idempotent.  This

```bash

$ gotest -v ./...  -timeout 30s -run ^TestNonIdempotency$
=== RUN   TestNonIdempotency
    activities_test.go:90: Account balance correctly reduced to 80.00 after two non-idempotent calls
    activities_test.go:94: First request took: 1.1224115s
    activities_test.go:95: Second request took: 1.308017916s
    activities_test.go:96: Overall operation took: 1.409267167s
--- PASS: TestNonIdempotency (1.41s)
PASS
ok       app/internal/batch      1.664s

```

## Making scripts idempotent

Show how by just wrapping the script in a call; gets idempotency for free.  
This is done by on the behavior of the WorkflowID; which should represent a business

## Making scripts invincible!

Now that the activities are idempotent.  The whole workflow should now be ported over easily; just copy the loop over.  The workflow ID chosen here could be the process name so that if it took longer than expected; another process will not overlap.  Demo this too ..

It can be demo-ed by randomly stopping the workers and server.  
Remember to use the persistent version! Use the short cut; start-temporal-db
