FROM golang

# bootstrap os
RUN apt -y update


WORKDIR /home/app

# copy project files & and go inside project directory
COPY go.mod .
RUN go mod tidy

COPY . .

CMD go run ./api/proto/server/main.go
