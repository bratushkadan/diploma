package pb

// go:generate protoc --go_out=. --go-grpc_out=.  --include_imports --descriptor_set_out=floral/auth/v1/descriptors.pb -I ../proto floral/auth/v1/user.proto
//go:generate protoc --go_out=. --go-grpc_out=. -I ../proto --grpc-gateway_out=logtostderr=true:../generated floral/auth/v1/user.proto
