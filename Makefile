test:
	go test github.com/erlorenz/bc-go/bc

test-all:
	go test -race  ./...

test-int:
	go test github.com/erlorenz/bc-go/internal/testv2



.PHONY: test, test-all,  test-int

