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