default: compile

compile:
	(cd src/secguro && go build -o ../../build/secguro .)
