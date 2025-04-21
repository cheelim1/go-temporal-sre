package kilcron

import (
	"fmt"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"strconv"
	"time"
)

func PaymentWorkflow(ctx workflow.Context, paymentID string) error {
	fmt.Println("INSIDE ===> PaymentWorkflow")
	// Options for retry ..
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 60 * time.Second, // Adjust as necessary
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Millisecond * 500,
			BackoffCoefficient: 2.0,
			MaximumAttempts:    10,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// DEBUG
	//workflow.ExecuteActivity(ctx, MakePayment, paymentID+"-42").Get(ctx, nil)
	var futures []workflow.Future
	for i := 0; i < 10; i++ {
		activityName := "MakePayment-" + paymentID + "-" + strconv.Itoa(i)
		fmt.Println("INSIDE ===> Activity", activityName)
		//workflow.GoNamed(ctx, activityName, func(ctx workflow.Context) {
		//	future := workflow.ExecuteActivity(ctx, MakePayment, paymentID+"-"+string(i))
		//	futures = append(futures, future)
		//})
		future := workflow.ExecuteActivity(ctx, MakePayment, activityName)
		futures = append(futures, future)
	}

	// Now collect it back ..
	for _, future := range futures {
		if err := future.Get(ctx, nil); err != nil {
			return err
		}
	}
	return nil
}
