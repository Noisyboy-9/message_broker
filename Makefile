gen:
	protoc --go_out=./api/pb/ --go_opt=paths=source_relative --go-grpc_out=./api/pb/ --go-grpc_opt=paths=source_relative ./api/proto/*.proto
clean:
	rm -rf ./api/pb
	mkdir ./api/pb
