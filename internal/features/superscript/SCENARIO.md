# SCENARIO

## Objective

- Show how to wrap a non-idempotent script to become idempotent

- Show how to easily port over to Temporal workflow

## Directory structure

cmd/superscript |
    -- main.go <-- main Run loop here; for demo run concurrent http + worker
    -- http.go <-- HTTP handler here
    -- worker.go <-- Worker code here; including any activities setup

internal/superscript |
    -- SCENARIO.md <-- this file; high level objective and expected output
    -- superscript.go <-- structs + shared models here 
    -- activities.go <-- Activities code here 
    -- workflow.go <-- Workflow code here
    -- scripts <-- All non-idempotent scripts here

## Standards

  - Use Golang v1.24.1 standard libs + idioms as much as possible
  - Review example of script wrapper library; use its example as much as possible when executing the bash script - https://github.com/bitfield/script
  - Use idiomatic Temporal patterns as much as possible
  - Make code easily readable; be exact and clear; do not be clever

## Make single script to be Idempotent

- As learned previously; wrap the script for a single payment collection 
  ./scripts/single_payment_collection.sh inside an Activity to be idempotent with OrderID as WorkflowID

- Wrap the script using the library https://github.com/bitfield/script; use its capabilities to execute
  command and return the exit code as success or failure

- With Temporal; execution will be idempotent even with multiple calls

- Will return previous successful runs if run multiple time; without triggering script again

- Can trigger one sample OrderID - ORD-DEMO-123 to show how the script wrapping will work

## Port over Orchestration 

- Parent Workflow which acts as the Orchestrator will have a fixed WorkflowID based on the Date

- The loop calling the single payment collection should be ported over with great fidelity; 
  the script is at ./scripts/traditional_payment_collection.sh

- The Orchestrator output should be ported over; pay attention to the output message, content and color

- Each OrderID being a new child workflow; with the OrderID being the WorkflowID of the previously 
  described wrapped workflow

- After all the child workflows are completed; can close off the parent workflow


## Demo

- Output instructions of setup into a README.md file in same folder as this file.  Have the step-by-step to demo the following below items

- Will first run the script ./scripts/traditional_payment_collection.sh a few times to demo that 
  non-idempotency will create undesired effects

- Will next show how to easily wrap the non-idempotent code in a Temporal Workflow + Activities 
  using the script library - https://github.com/bitfield/script

- Will show the workflow for a single collection will retry as per needed until it succeeds 

- Show multiple call to the same sample OrderID will always return the same result

- Finally will run the full ported Orchestrated Workflow; multiple times and show that it will 
  complete successfully without having duplicate calls to the Activity 

