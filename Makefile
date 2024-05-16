default: compile

compile:
	go build -o build/secguro .

compile-dev:
	go build -tags=dev -o build/secguro .

lint:
	golangci-lint run
