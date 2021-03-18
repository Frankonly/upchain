package cli

import (
	"crypto/tls"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	pb "github.com/frankonly/upchain/api/accumulator"
)

var apiClient pb.AccumulatorClient

// Client news or returns a accumulator client
func Client() pb.AccumulatorClient {
	if apiClient == nil {
		var err error
		var conn *grpc.ClientConn

		if secureConn {
			conn, err = grpc.Dial(endpoint, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
			if err != nil {
				log.Fatalf("failed to establish connect(TLS) with %s: %v", endpoint, err)
			}
		} else {
			conn, err = grpc.Dial(endpoint, grpc.WithInsecure())
			if err != nil {
				log.Fatalf("failed to establish insurece connect with %s: %v", endpoint, err)
			}
		}

		apiClient = pb.NewAccumulatorClient(conn)
	}

	return apiClient
}
