default: compile

compile:
	(cd src/secguro && go build -o ../../build/secguro .)

lint:
	(cd src/secguro && golangci-lint run)
