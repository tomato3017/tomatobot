dep:
	go mod download
	go mod tidy

generate:
	go generate ./...

build: dep generate
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux go build -o ./bin/ .

image:
	docker build . -q --tag tomatobot:latest

lint:
	docker run --rm -v $(CURDIR):/app -w /app golangci/golangci-lint:v1.59.1 golangci-lint run -v

test:
	go test -race -v ./...