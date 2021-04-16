package main

import (
	"flag"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"

	"github.com/frankonly/upchain/api"
	pb "github.com/frankonly/upchain/api/accumulator"
	"github.com/frankonly/upchain/data"
	"github.com/frankonly/upchain/log"
	"github.com/frankonly/upchain/storage"
)

var (
	tls      = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	certFile = flag.String("cert_file", "", "The TLS cert file")
	keyFile  = flag.String("key_file", "", "The TLS key file")
	dbDir    = flag.String("db_dir", "accumulator.db", "The upchain DB directory")
	port     = flag.Int("port", 10000, "The server port")
)

func main() {
	flag.Parse()
	logger := log.New()

	db, err := storage.NewLevelDB(data.Path(*dbDir))
	if err != nil {
		logger.Fatalf("failed to initialize db: %v", err)
	}

	merkle, err := storage.NewMerkleTreeStreaming(db)
	if err != nil {
		logger.Fatalf("failed to initialize merkle accumulator: %v", err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		logger.Fatalf("failed to listen: %v", err)
	}

	var opts []grpc.ServerOption
	if *tls {
		if *certFile == "" {
			*certFile = data.Path("x509/server_cert.pem")
		}
		if *keyFile == "" {
			*keyFile = data.Path("x509/server_key.pem")
		}
		creds, err := credentials.NewServerTLSFromFile(*certFile, *keyFile)
		if err != nil {
			logger.Fatalf("Failed to generate credentials %v", err)
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}

	grpcServer := grpc.NewServer(opts...)
	apiServer := api.NewServer(merkle, logger)
	pb.RegisterAccumulatorServer(grpcServer, apiServer)
	reflection.Register(grpcServer)

	logger.Infow("upchain starts serving", "port", *port)
	err = grpcServer.Serve(lis)
	logger.Errorw("upchain stops", "err", err)
}
