package main

import (
	"app/internal/kilcron"
	"context"
	"fmt"
	"go.temporal.io/sdk/client"
	"net/http"
)

const TQ = "kilcron-task-queue"
const orgID = "GopherPayNET"

// debugAccessHandler will start flaky cron ..
func debugAccessHandler(w http.ResponseWriter, r *http.Request) {
	payID := "GoBux"
	// Check if action is happening ... after done redirect back ..
	q := r.URL.Query()
	if q.Has("action") {
		switch q.Get("action") {
		case "flaky":
			payID += "Flaky"
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}

	wfr, err := c.ExecuteWorkflow(context.Background(), client.StartWorkflowOptions{
		ID:        orgID,
		TaskQueue: TQ,
	}, kilcron.PaymentWorkflow, payID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed to start workflow: %v", err)
		return
	}
	render := `
<html>
<body>

`
	render += "WF Started - ID: " + wfr.GetID() + " RunID: " + wfr.GetRunID()
	render += `
	<div>
	<p><a href="/demo/debug/">Run Normal</a></p>
	<p><a href="/demo/debug/?action=flaky">Run Flaky</a></p>
	</div>
</body>
</html>
`
	// DEBUG
	//fmt.Println(render)
	fmt.Fprintf(w, render)
	return
}
