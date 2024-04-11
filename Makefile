default: compile

compile:
	go build -o build/secguro .

lint:
	golangci-lint run
