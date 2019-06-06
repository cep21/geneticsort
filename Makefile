PROFILE_RUN ?= .
BENCH_RUN ?= .

build:
	go build ./...

run:
	go run main.go

# Run unit tests
test:
	env "GORACE=halt_on_error=1" go test -v -race ./...

# Format the code
fix:
	find . -iname '*.go' -not -path '*/vendor/*' -print0 | xargs -0 gofmt -s -w
	find . -iname '*.go' -not -path '*/vendor/*' -print0 | xargs -0 goimports -w

# Run benchmark examples
bench:
	go test -v -benchmem -run=^$$ -bench=$(BENCH_RUN) ./...

# Get some memory profiles
profile_memory:
	go test -benchmem -run=^$$ -bench=$(PROFILE_RUN) -memprofile=mem.out ./internal/arraysort
	go tool pprof --alloc_space arraysort.test mem.out
	rm -f mem.out arraysort.test

# Get some cpu profiles
profile_cpu:
	go test -run=^$$ -bench=$(PROFILE_RUN) -benchtime=15s -cpuprofile=cpu.out ./internal/arraysort
	go tool pprof arraysort.test cpu.out
	rm -f cpu.out arraysort.test

# Get some cpu profiles
profile_blocking:
	go test -run=^$$ -bench=$(PROFILE_RUN) -benchtime=15s -blockprofile=block.out ./internal/arraysort
	go tool pprof arraysort.test block.out
	rm -f block.out arraysort.test

# Lint the code
lint:
	golangci-lint run

# ci installs dep by direct version.  Users install with 'go get'
setup_ci:
	GO111MODULE=on go get github.com/golangci/golangci-lint/cmd/golangci-lint
