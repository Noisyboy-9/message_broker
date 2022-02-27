FROM golang

# bootstrap os
RUN apt -y update

# install protoc
RUN apt install -y protobuf-compiler

# install make file
RUN apt install -y make
RUN apt install -y build-essential

# copy project files & and go inside project directory

WORKDIR ./message_broker

COPY go.mod .
RUN go mod tidy

COPY . .

ENV PROMETHEUS_SERVER_PORT=8000
ENV BROKER_SERVER_PORT=9000

EXPOSE $PROMETHEUS_SERVER_PORT
EXPOSE $BROKER_SERVER_PORT

# bootstrap go-protc
CMD go run ./api/server/main.go

