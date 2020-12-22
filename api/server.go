package api

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "upchain/api/accumulator"
	"upchain/storage"
)

type Server struct {
	pb.UnimplementedAccumulatorServer
	accumulator storage.MerkleAccumulator
}

func NewServer(accumulator storage.MerkleAccumulator) *Server {
	return &Server{accumulator: accumulator}
}

func (s Server) Append(_ context.Context, hash *pb.Hash) (*pb.ID, error) {
	// TODO: check length of hash
	id, err := s.accumulator.Append(hash.Hash)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to append new hash")
	}

	return &pb.ID{Id: id}, nil
}

func (s Server) Get(_ context.Context, id *pb.ID) (*pb.Hash, error) {
	hash, err := s.accumulator.Get(id.Id)
	switch {
	case errors.Is(err, storage.ErrOutOfRange):
		return nil, status.Error(codes.OutOfRange, err.Error())
	case errors.Is(err, storage.ErrNotFound):
		return nil, status.Error(codes.NotFound, err.Error())
	case err != nil:
		return nil, status.Error(codes.Internal, err.Error())
	default:
		return &pb.Hash{Hash: hash}, nil
	}
}

func (s Server) Search(_ context.Context, hash *pb.Hash) (*pb.ID, error) {
	id, err := s.accumulator.Search(hash.Hash)
	switch {
	case errors.Is(err, storage.ErrNotFound):
		return nil, status.Error(codes.NotFound, err.Error())
	case err != nil:
		return nil, status.Error(codes.Internal, err.Error())
	default:
		return &pb.ID{Id: id}, nil
	}
}

func (s Server) GetDigest(context.Context, *pb.Empty) (*pb.Hash, error) {
	digest, err := s.accumulator.Digest()
	switch {
	case errors.Is(err, storage.ErrEmpty):
		return nil, status.Error(codes.Unavailable, err.Error())
	case err != nil:
		return nil, status.Error(codes.Internal, err.Error())
	default:
		return &pb.Hash{Hash: digest}, nil
	}
}

func (s Server) GetProofByID(_ context.Context, id *pb.ID) (*pb.HashProof, error) {
	return s.getProofByID(id.Id, nil)
}

func (s Server) GetProofByHash(_ context.Context, hash *pb.Hash) (*pb.HashProof, error) {
	id, err := s.accumulator.Search(hash.Hash)
	if errors.Is(err, storage.ErrNotFound) {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return s.getProofByID(id, nil)
}

func (s Server) GetOldProofByID(_ context.Context, in *pb.GetOldProofByIDRequest) (*pb.HashProof, error) {
	return s.getProofByID(in.Id, in.Digest)
}

func (s Server) GetOldProofByHash(_ context.Context, in *pb.GetOldProofByHashRequest) (*pb.HashProof, error) {
	id, err := s.accumulator.Search(in.Hash)
	if errors.Is(err, storage.ErrNotFound) {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return s.getProofByID(id, in.Digest)
}

func (s Server) getProofByID(id uint64, digest []byte) (*pb.HashProof, error) {
	path, err := s.accumulator.GetProof(id, digest)
	switch {
	case errors.Is(err, storage.ErrOutOfRange):
		return nil, status.Error(codes.OutOfRange, err.Error())
	case errors.Is(err, storage.ErrNotFound):
		return nil, status.Error(codes.NotFound, err.Error())
	case err != nil:
		return nil, status.Error(codes.Internal, err.Error())
	case len(path) == 0:
		return nil, status.Error(codes.Internal, "failed to generate proof")
	default:
		proof := &pb.HashProof{}
		proof.Hash = path[0]
		proof.Digest = path[len(path)-1]
		if len(path) > 1 {
			proof.Path = path[1 : len(path)-1]
		}

		return proof, nil
	}
}
