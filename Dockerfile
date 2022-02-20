FROM golang

# setup os
RUN apt -y update

# install protoc
RUN apt install -y protobuf-compiler

# install make file
RUN apt install -y make
RUN apt install -y build-essential

# copy project files & and go inside project directory
COPY ./api ./message_broker/api/
COPY ./internal ./message_broker/internal
COPY ./pkg ./message_broker/pkg
COPY ./prometheus ./message_broker/prometheus
COPY ./.gitignore ./message_broker
COPY ./go.mod ./message_broker
COPY ./main.go ./main.go
COPY Makefile ./message_broker
COPY ./README.md ./message_broker

WORKDIR ./message_broker

ENV PROMETHEUS_SERVER_PORT=8000
ENV BROKER_SERVER_PORT=9000

EXPOSE $PROMETHEUS_SERVER_PORT
EXPOSE $BROKER_SERVER_PORT

RUN go mod tidy
# setup go-protc
CMD go run ./api/server/main.go

