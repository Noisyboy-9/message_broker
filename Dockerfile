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
COPY ./.gitignore ./message_broker
COPY ./go.mod ./message_broker
COPY ./main.go ./main.go
COPY Makefile ./message_broker
COPY ./README.md ./message_broker

WORKDIR ./message_broker

# RUN go mod tidy
RUN go mod tidy


EXPOSE 8080/udp
EXPOSE 8080/tcp

# setup go-protc
CMD go run ./api/server/main.go

