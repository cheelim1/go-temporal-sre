run:
	@go run *.go
test:
	@gotest ./...

server:
	@temporal server 
