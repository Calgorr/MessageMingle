FROM golang:1.21.1-alpine

WORKDIR /home/app

# copy project files & and go inside project directory
COPY . .
RUN go mod tidy
CMD go run ./api/proto/server/main.go
