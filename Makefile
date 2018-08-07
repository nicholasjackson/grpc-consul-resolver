build:
	go build ./*.go

test_unit:
	go test -race ./...

test_functional:
	cd functional_tests && go test -v --godog.format=pretty --godog.random
