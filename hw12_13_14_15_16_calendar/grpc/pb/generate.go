package pb

//go:generate protoc -I=../../api --go_out=. --go-grpc_out=. --grpc-gateway_out=. ../../api/event/EventService.proto
