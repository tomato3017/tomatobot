build:
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux go build -o ./bin/ ./cmd/tomatobot

image:
	docker build . -q --tag tomatobot:latest