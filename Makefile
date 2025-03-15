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

