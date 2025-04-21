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

#skip
start-kilcron:
	@echo "Start kilcron .."
	@cd cmd/demos/kilcron && go run main.go

stop:
	@kill `pgrep temporal`

run-script:
	@echo "Demo non-Idempotent script. Do NOT run twice!!"
	@cd ./internal/features/superscript/scripts && ./traditional_payment_collection.sh

# SuperScript Demo Targets
.PHONY: superscript-demo-1 superscript-start superscript-demo-2 superscript-demo-3 superscript-demo-4 superscript-stop

#run beforehand
superscript-setup:
	@echo "Running SuperScript Demo: Setup and Build"
	@cd ./internal/features/superscript/scripts && ./demo-1-setup.sh

#run beforehand
superscript-start:
	@echo "Starting SuperScript Application"
	go run cmd/demos/superscript/*.go

superscript-demo-2:
	@echo "Running SuperScript Demo 2: Traditional Non-Idempotent Script"
	@./internal/features/superscript/scripts/traditional_payment_collection.sh

superscript-demo-3:
	@echo "Running SuperScript Demo 3: Single Payment Idempotent Workflow"
	@./internal/features/superscript/scripts/demo-3-single-payment.sh

superscript-demo-4:
	@echo "Running SuperScript Demo 4: Orchestrator Workflow"
	@./internal/features/superscript/scripts/demo-4-orchestrator.sh

superscript-stop:
	@echo "Stopping SuperScript Application"
	@./internal/features/superscript/scripts/demo-stop.sh 

# JIT Demo Targets
.PHONY: jit-demo-start jit-server jit-fe jit-demo-stop jit-deps

jit-deps:
	@echo "Checking and installing dependencies..."
	@which godotenv >/dev/null 2>&1 || (echo "Installing godotenv..." && go install github.com/joho/godotenv/cmd/godotenv@latest)
	@export PATH="$$(go env GOPATH)/bin:$$PATH"

##ensure start-temporal is running
start-jit-worker: jit-deps
	@echo "Starting Temporal JIT Worker ..."
	@cd internal/demos/jit && PATH="$$(go env GOPATH)/bin:$$PATH" godotenv -f .env.local go run main.go

##ensure start-temporal is running
start-jit-server: jit-deps
	@echo "Starting Backend Server..."
	@cd internal/demos/jit && PATH="$$(go env GOPATH)/bin:$$PATH" godotenv -f .env.local go run server.go

start-jit-fe:
	@echo "Starting Streamlit Frontend..."
	@cd internal/demos/jit && python3 -m streamlit run app.py

jit-demo-start:
	@echo "Starting JIT Demo (all components)..."
	@make start-jit-worker &
	@make start-jit-server &
	@sleep 5  # Give Temporal Server time to start
	@make start-jit-fe

jit-demo-stop:
	@echo "Stopping JIT Demo components..."
	@make stop  # This will stop the Temporal server
	@pkill -f "internal/demos/jit/main.go" || true
	@pkill -f "internal/demos/jit/server.go" || true
	@pkill -f "streamlit run app.py" || true 

# MongoDB Demo Targets
.PHONY: jit-mongo-demo

jit-mongo-demo:
	@echo "Connecting to MongoDB and running demo commands..."
	@cd internal/demos/jit && \
		MONGODB_PASSWORD=$$(grep MONGODB_PASSWORD .env.local | cut -d '=' -f2) && \
		mongosh "mongodb+srv://demo-user:$$MONGODB_PASSWORD@clustercl.mrszj.azure.mongodb.net/demo?authSource=admin" \
		--apiVersion 1 \
		--quiet \
		--eval 'print("Inserting document..."); \
			db.getCollection("demo-jit").insertOne({ name: "John Doe", age: 27 }); \
			print("\nCurrent documents in collection:"); \
			db.getCollection("demo-jit").find().pretty()' 