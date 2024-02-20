test:
	go test "github.com/erlorenz/bcgo/bc"

test/all:
	go test -v -race  ./...

test/int:
	go test -v  "github.com/erlorenz/bcgo/integration"


.PHONY: test, test/all, test/int

