package api

import (
	"context"
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	pb "upchain/api/accumulator"
	"upchain/storage"
)

type APIServer struct {
	pb.UnimplementedAccumulatorServer
	accumulator storage.MerkleAccumulator
}

func NewServer(accumulator storage.MerkleAccumulator) *APIServer {
	return &APIServer{accumulator: accumulator}
}

func (s APIServer) Append(_ context.Context, hash *pb.Hash) (*pb.ID, error) {
	// TODO: check length of hash
	id, err := s.accumulator.Append(hash.Hash)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to append new hash")
	}

	return &pb.ID{Id: id}, nil
}

func (s APIServer) Get(_ context.Context, id *pb.ID) (*pb.Hash, error) {
	hash, err := s.accumulator.Get(id.Id)
	switch {
	case errors.Is(err, storage.ErrOutOfRange):
		return nil, status.Error(codes.OutOfRange, err.Error())
	case errors.Is(err, storage.ErrNotFound):
		return nil, status.Error(codes.NotFound, err.Error())
	case err != nil:
		return nil, status.Error(codes.Internal, err.Error())
	default:
		return &pb.Hash{Hash: hash}, err
	}
}

func (s APIServer) GetProofByID(_ context.Context, id *pb.ID) (*pb.GetProofReply, error) {
	path, err := s.accumulator.GetProof(id.Id)
	switch {
	case errors.Is(err, storage.ErrOutOfRange):
		return nil, status.Error(codes.OutOfRange, err.Error())
	case errors.Is(err, storage.ErrNotFound):
		return nil, status.Error(codes.NotFound, err.Error())
	case err != nil:
		return nil, status.Error(codes.Internal, err.Error())
	case len(path) == 0:
		return nil, status.Error(codes.Internal, "failed to get proof")
	default:
		reply := &pb.GetProofReply{}
		reply.TargetHash = path[0]
		reply.Digest = path[len(path)-1]
		if len(path) > 1 {
			reply.Path = path[1 : len(path)-1]
		}

		return reply, nil
	}
}
