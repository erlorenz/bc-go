test:
	go test github.com/erlorenz/bc-go/bc

test/all:
	go test -race  ./...

test/all/v:
	go test -v -race  ./...

test/int:
	go test github.com/erlorenz/bc-go/internal/testv2

test/int/v:
	go test -v -race github.com/erlorenz/bc-go/internal/testv2


.PHONY: test, test/all, test/all/v, test/int, test/int/v

