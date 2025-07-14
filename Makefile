run:
	@go run *.go
test:
	@gotest ./...

server:
	@temporal server 

present:
	@present -notes

#run beforehand
start-temporal:
	@echo "Start Temporal Server"
	@temporal server start-dev --http-port 9090 --ui-port 3001 --metrics-port 9091 --log-level info

stop:
	@kill `pgrep temporal`

# Centralized Worker Targets
.PHONY: worker start-worker kilcron-demo superscript-demo jit-demo jit-fe jit-fe-setup

# Start centralized worker with all features
start-worker:
	@echo "Starting Centralized Temporal Worker"
	@set -a; [ -f .env ] && source .env; set +a; go run cmd/worker/main.go

# Start centralized worker with specific features
start-worker-features:
	@echo "Starting Centralized Worker with specific features: $(FEATURES)"
	@set -a; [ -f .env ] && source .env; set +a; ENABLED_FEATURES=$(FEATURES) go run cmd/worker/main.go

# Kilcron demo using centralized worker
kilcron-demo:
	@echo "Starting Kilcron Demo (using centralized worker)"
	@set -a; [ -f .env ] && source .env; set +a; go run cmd/demos/kilcron/main.go

# SuperScript demo using centralized worker
superscript-demo:
	@echo "Starting SuperScript Demo (using centralized worker)"
	@set -a; [ -f .env ] && source .env; set +a; go run cmd/demos/superscript/main.go

# JIT demo using centralized worker
jit-demo:
	@echo "Starting JIT Access Demo (using centralized worker)"
	@set -a; [ -f .env ] && source .env; set +a; go run cmd/demos/jit/main.go

# JIT frontend setup - create virtual environment and install dependencies
jit-fe-setup:
	@echo "Setting up JIT Frontend virtual environment..."
	@cd demo/jit/demo-fe && python3 -m venv venv
	@echo "Installing dependencies..."
	@cd demo/jit/demo-fe && venv/bin/pip install -r requirements.txt
	@echo "Setup complete! You can now run 'make jit-fe'"

# JIT frontend (Streamlit app)
jit-fe:
	@echo "Starting JIT Access Frontend (Streamlit)"
	@echo "Note: If setup not done, run 'make jit-fe-setup' first"
	@cd demo/jit/demo-fe && venv/bin/streamlit run app.py

# Legacy support - redirect to new demos
start-kilcron: kilcron-demo
	@echo "Note: start-kilcron now uses the centralized worker"

# Build targets for the new structure
build-all: build-worker build-demos

build-worker:
	@echo "Building centralized worker..."
	@go build -o bin/centralized-worker cmd/worker/main.go

build-demos:
	@echo "Building demo applications..."
	@go build -o bin/kilcron-demo cmd/demos/kilcron/main.go
	@go build -o bin/superscript-demo cmd/demos/superscript/main.go
	@go build -o bin/jit-demo cmd/demos/jit/main.go

# Clean up build artifacts
clean:
	@echo "Cleaning up build artifacts..."
	@rm -f bin/centralized-worker bin/kilcron-demo bin/superscript-demo bin/jit-demo

run-script:
	@echo "Demo non-Idempotent script. Do NOT run twice!!"
	@cd ./internal/superscript/scripts && ./traditional_payment_collection.sh

# SuperScript Demo Targets
.PHONY: superscript-demo-1 superscript-start superscript-demo-2 superscript-demo-3 superscript-demo-4 superscript-stop

#run beforehand
superscript-setup:
	@echo "Running SuperScript Demo: Setup and Build"
	@cd ./internal/superscript/scripts && ./demo-1-setup.sh
	# @go build -o bin/superscript ./cmd/superscript/

#run beforehand
superscript-start:
	@echo "Starting SuperScript Application"
	# chmod +x ./internal/superscript/scripts/demo-start.sh
	#@./internal/superscript/scripts/demo-start.sh
	go run cmd/superscript/*.go

superscript-demo-2:
	@echo "Running SuperScript Demo 2: Traditional Non-Idempotent Script"
	@./internal/superscript/scripts/traditional_payment_collection.sh

superscript-demo-3:
	@echo "Running SuperScript Demo 3: Single Payment Idempotent Workflow"
	@./internal/superscript/scripts/demo-3-single-payment.sh

superscript-demo-4:
	@echo "Running SuperScript Demo 4: Orchestrator Workflow"
	@./internal/superscript/scripts/demo-4-orchestrator.sh

superscript-stop:
	@echo "Stopping SuperScript Application"
	@./internal/superscript/scripts/demo-stop.sh 

.PHONY: jit-demo-start jit-demo-stop

jit-demo-start: jit-demo
	@echo "Note: jit-demo-start now uses the centralized worker (jit-demo)"

jit-demo-stop:
	@echo "Stopping processes..."
	@pkill -f "cmd/demos/jit/main.go" || true 

# MongoDB Demo Targets
.PHONY: jit-mongo-demo

jit-mongo-demo:
	@echo "Connecting to MongoDB and running demo commands..."
	@cd demo/jit/demo-be && \
		MONGODB_PASSWORD=$$(grep MONGODB_PASSWORD .env.local | cut -d '=' -f2) && \
		mongosh "mongodb+srv://demo-user:$$MONGODB_PASSWORD@clustercl.mrszj.azure.mongodb.net/demo?authSource=admin" \
		--apiVersion 1 \
		--quiet \
		--eval 'print("Inserting document..."); \
			db.getCollection("demo-jit").insertOne({ name: "John Doe", age: 27 }); \
			print("\nCurrent documents in collection:"); \
			db.getCollection("demo-jit").find().pretty()' 