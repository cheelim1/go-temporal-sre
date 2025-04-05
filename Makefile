run:
	@go run *.go
test:
	@gotest ./...

server:
	@temporal server 

present:
	@present -notes

start-temporal:
	@echo "Start Temporal Server"
	@temporal server start-dev --http-port 9090 --ui-port 3001 --metrics-port 9091 --log-level info

start-kilcron:
	@echo "Start kilcron .."
	@cd cmd/kilcron && go run debug.go worker.go main.go

stop:
	@kill `pgrep temporal`

run-script:
	@echo "Demo non-Idempotent script. Do NOT run twice!!"
	@cd ./internal/superscript/scripts && ./traditional_payment_collection.sh

# SuperScript Demo Targets
.PHONY: superscript-demo-1 superscript-start superscript-demo-2 superscript-demo-3 superscript-demo-4 superscript-stop

superscript-setup:
	@echo "Running SuperScript Demo: Setup and Build"
	@cd ./internal/superscript/scripts && ./demo-1-setup.sh
	# @go build -o bin/superscript ./cmd/superscript/

superscript-start:
	@echo "Starting SuperScript Application"
	# chmod +x ./internal/superscript/scripts/demo-start.sh
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

# JIT Demo Targets
.PHONY: jit-demo-start jit-server jit-fe jit-demo-stop jit-deps

jit-deps:
	@echo "Checking and installing dependencies..."
	@which godotenv >/dev/null 2>&1 || (echo "Installing godotenv..." && go install github.com/joho/godotenv/cmd/godotenv@latest)
	@export PATH="$$(go env GOPATH)/bin:$$PATH"

##ensure start-temporal is running
start-jit-worker: jit-deps
	@echo "Starting Temporal JIT Worker ..."
	@cd demo/jit/demo-be && PATH="$$(go env GOPATH)/bin:$$PATH" godotenv -f .env.local go run cmd/worker/main.go

##ensure start-temporal is running
start-jit-server: jit-deps
	@echo "Starting Backend Server..."
	@cd demo/jit/demo-be && PATH="$$(go env GOPATH)/bin:$$PATH" godotenv -f .env.local go run cmd/server/main.go

start-jit-fe:
	@echo "Starting Streamlit Frontend..."
	@cd demo/jit/demo-fe && python3 -m streamlit run app.py

jit-demo-start:
	@echo "Starting JIT Demo (all components)..."
	@make start-jit-worker &
	@make start-jit-server &
	@sleep 5  # Give Temporal Server time to start
	@make start-jit-fe

jit-demo-stop:
	@echo "Stopping JIT Demo components..."
	@make stop  # This will stop the Temporal server
	@pkill -f "cmd/worker/main.go" || true
	@pkill -f "cmd/server/main.go" || true
	@pkill -f "streamlit run app.py" || true 