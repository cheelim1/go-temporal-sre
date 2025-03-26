package data_enrichment

import (
	"go.temporal.io/sdk/workflow"
	"reflect"
	"testing"
)

func TestDataEnrichmentWorkflow(t *testing.T) {
	type args struct {
		ctx       workflow.Context
		customers []Customer
	}
	tests := []struct {
		name    string
		args    args
		want    []EnrichedCustomer
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DataEnrichmentWorkflow(tt.args.ctx, tt.args.customers)
			if (err != nil) != tt.wantErr {
				t.Errorf("DataEnrichmentWorkflow() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DataEnrichmentWorkflow() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnrichSingleCustomerWorkflow(t *testing.T) {
	type args struct {
		ctx      workflow.Context
		customer Customer
	}
	tests := []struct {
		name    string
		args    args
		want    EnrichedCustomer
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EnrichSingleCustomerWorkflow(tt.args.ctx, tt.args.customer)
			if (err != nil) != tt.wantErr {
				t.Errorf("EnrichSingleCustomerWorkflow() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EnrichSingleCustomerWorkflow() got = %v, want %v", got, tt.want)
			}
		})
	}
}
