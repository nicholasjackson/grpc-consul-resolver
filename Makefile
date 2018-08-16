build:
	go build ./*.go

test_unit:
	go test -v -race .

goconvey:
	goconvey -excludedDirs "vendor,functional_tests"

test_functional:
	cd functional_tests && go test -v --godog.format=pretty --godog.random #--godog.tags=@connect
