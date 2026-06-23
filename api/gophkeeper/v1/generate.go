package gophkeeperv1

//go:generate protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative api/gophkeeper/v1/auth.proto api/gophkeeper/v1/secret.proto
