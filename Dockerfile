# syntax=docker/dockerfile:1

FROM golang:1.22.2

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod ./
RUN go mod download

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/reference/dockerfile/#copy
COPY *.go ./
COPY Makefile ./

## Build
#RUN make build
#
## Run
#CMD ["/tomatobot"]