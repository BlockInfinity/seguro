default: compile

compile:
	(cd src && go build -o ../build/secguro .)

lint:
	(cd src && golangci-lint run)
