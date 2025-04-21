package superscript

import (
	"context"
	gocmp "github.com/google/go-cmp/cmp"
	"go.temporal.io/sdk/log"
	"log/slog"
	"testing"
)

func TestActivities_RunPaymentCollectionScript(t *testing.T) {
	// Create a basic logger - we'll use a simple adapter for standard logger

	type fields struct {
		ScriptBasePath string
		Logger         log.Logger
	}
	type args struct {
		ctx     context.Context
		orderID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *PaymentResult
		wantErr bool
	}{
		{"case #0", fields{
			ScriptBasePath: "./",
			Logger:         slog.Default(),
		}, args{
			ctx:     context.Background(),
			orderID: "ORD-1234", // Will fail script which expects int
		}, &PaymentResult{
			OrderID:      "ORD-1234",
			Success:      false,
			Output:       "ERROR: OrderID must be a number\nCleaning up resources...\nERROR: Script terminated with exit code: 2\n",
			ErrorMessage: "Script failed with exit code: 2",
			ExitCode:     2,
		}, true},
		{"case #1", fields{
			ScriptBasePath: "./",
			Logger:         slog.Default(),
		}, args{
			ctx:     context.Background(),
			orderID: "4242",
		}, &PaymentResult{
			OrderID:      "4242",
			Success:      true,
			Output:       "Starting payment processing for OrderID: 4242\nStarting processing step 1...\nStep 1 completed successfully: Step1 4242\nPayment processing completed successfully for OrderID: 4242\nCleaning up resources...\n",
			ErrorMessage: "",
			ExitCode:     0,
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Activities{
				ScriptBasePath: tt.fields.ScriptBasePath,
				Logger:         tt.fields.Logger,
			}
			got, err := a.RunPaymentCollectionScript(tt.args.ctx, tt.args.orderID)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunPaymentCollectionScript() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// DEBUG
			//spew.Dump(got)
			gotSample := &PaymentResult{
				OrderID:      got.OrderID,
				Success:      got.Success,
				Output:       got.Output,
				ErrorMessage: got.ErrorMessage,
				ExitCode:     got.ExitCode,
			}
			// Check subset
			if !gocmp.Equal(gotSample, tt.want) {
				t.Errorf("RunPaymentCollectionScript() Diff: %s", gocmp.Diff(gotSample, tt.want))
			}
			// Full check will never match as it has more info
			//if !gocmp.Equal(got, tt.want) {
			//	t.Errorf("RunPaymentCollectionScript() Diff: %s", gocmp.Diff(tt.want, got))
			//}
		})
	}
}
