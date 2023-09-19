FROM golang:1.21.1-alpine

WORKDIR /home/app

# copy project files & and go inside project directory
COPY go.mod .
COPY go.sum .
RUN go mod download -x

COPY . .

CMD go run ./api/proto/server/main.go
