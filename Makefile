run:
	@go run *.go
test:
	@gotest ./...

server:
	@temporal server 

present:
	@present -notes

start-kilcron:
	@echo "Start kilcron .."
	@cd cmd/kilcron && go run debug.go worker.go main.go

start-temporal:
	@echo "Start Temporal Server"
	@temporal server start-dev --http-port 9090 --ui-port 3001 --metrics-port 9091 --log-level info

stop:
	@kill `pgrep temporal`

run-script:
	@echo "Demo non-Idempotent script. Do NOT run twice!!"
	@cd ./internal/superscript/scripts && ./traditional_payment_collection.sh

# SuperScript Demo Targets
.PHONY: superscript-demo-1 superscript-start superscript-demo-2 superscript-demo-3 superscript-demo-4 superscript-stop

superscript-demo-1:
	@echo "Running SuperScript Demo 1: Setup and Build"
	@./internal/superscript/scripts/demo-1-setup.sh

superscript-start:
	@echo "Starting SuperScript Application"
	@chmod +x ./internal/superscript/scripts/demo-start.sh
	@./internal/superscript/scripts/demo-start.sh

superscript-demo-2:
	@echo "Running SuperScript Demo 2: Traditional Non-Idempotent Script"
	@./internal/superscript/scripts/demo-2-traditional.sh

superscript-demo-3:
	@echo "Running SuperScript Demo 3: Single Payment Idempotent Workflow"
	@./internal/superscript/scripts/demo-3-single-payment.sh

superscript-demo-4:
	@echo "Running SuperScript Demo 4: Orchestrator Workflow"
	@./internal/superscript/scripts/demo-4-orchestrator.sh

superscript-stop:
	@echo "Stopping SuperScript Application"
	@./internal/superscript/scripts/demo-stop.sh 
