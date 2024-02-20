test:
	go test "github.com/erlorenz/bc-go/bc"

test/all:
	go test -v -race  ./...

test/int:
	go test -v  "github.com/erlorenz/bc-go/integration"


.PHONY: test, test/all, test/int

