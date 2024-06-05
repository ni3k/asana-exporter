build:
	@go build -o bin/main
run: build
	@./bin/main
build-consumer:
	@go build consumer/main.go 
	@mv main bin/consumer
run-consumer: build-consumer
	@./bin/consumer
test:
	@go test -v ./...