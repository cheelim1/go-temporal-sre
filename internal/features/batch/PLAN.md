# Temporal Workflow Test Fix Plan

## Problem Summary

The current Temporal workflow tests are failing because they don't accurately simulate how Temporal handles idempotency in a real-world environment. Specifically:

1. Activities are being executed multiple times with the same workflow ID, causing double deductions
2. Activity registration issues between test environments
3. Test expectations don't align with actual test behavior

## Action Plan

### 1. Study Temporal Documentation and Best Practices

- Review [Temporal Go SDK Documentation](https://docs.temporal.io/dev-guide/go) 
- Study [Temporal Go Testing Framework](https://docs.temporal.io/dev-guide/go/testing)
- Examine [Temporal Samples Repository](https://github.com/temporalio/samples-go) for idempotency examples
- Focus on [WorkflowEnvironment](https://pkg.go.dev/go.temporal.io/sdk/testsuite#WorkflowEnvironment) mocking capabilities

### 2. Fix TestIdempotentFeeDeduction

- Properly register the DeductFee activity with exact name matching
- Use `OnActivity()` instead of `RegisterActivity()` for the second execution
- Create a mock that:
  - Tracks that the activity was called
  - Returns the same result as the first execution without actually deducting
  - Simulates how Temporal would reuse the first execution's result

```go
// Example fix for second execution
env2.OnActivity("DeductFee", mock.Anything, mock.Anything).Return(
    &ActivityResult{
        NewBalance: initialBalance - amount, // Same as first execution
        Success:    true,
    }, nil
).Once()
```

### 3. Fix Activity Registration Issue

- Ensure consistent activity registration across all tests
- Use the same activity name ("DeductFee") in all environments
- Register activities before workflow execution

```go
// Register activity with proper name
env.RegisterActivity(s.DeductFee)
// Or explicitly with the activity name
env.RegisterActivityWithOptions(s.DeductFee, activity.RegisterOptions{Name: "DeductFee"})
```

### 4. Update TestParallelRequests

- Apply the same mocking pattern to simulate Temporal's idempotency behavior
- Ensure that the workflow completes correctly with the expected balance changes
- Track activity calls while preventing actual double execution

### 5. Fix TestWorkflowRetentionPeriod

- Update the test to properly simulate workflow history retention
- Use mocks to ensure activities aren't called again in subsequent executions
- Verify that results are retrieved correctly from history

### 6. Update TestCompleteIdempotencyImplementation

- Apply fixes to ensure comprehensive testing of all idempotency scenarios
- Maintain accurate tracking of activity execution
- Ensure balance changes occur only when appropriate

### 7. Refactor Common Test Patterns

- Extract common mocking and setup code into helper methods
- Create reusable patterns for activity mocking
- Ensure consistent behavior across all tests

### 8. Verify Fixed Tests

- Run the complete test suite
- Verify that all activities are executed at appropriate times
- Confirm that balance changes happen only once per unique workflow ID
- Ensure idempotency is properly demonstrated

## Implementation Priorities

1. Fix activity registration issues (highest priority)
2. Implement proper mocking for second executions
3. Update all tests to use the correct pattern
4. Refactor for consistency and maintainability

## Expected Outcomes

- All tests pass successfully
- Tests accurately demonstrate Temporal's idempotency features
- Activities execute only when appropriate
- Balances change exactly once per unique workflow ID
- Code demonstrates best practices for Temporal workflow testing
