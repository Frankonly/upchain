package api

import (
	"context"
	"encoding/hex"
	"errors"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/frankonly/upchain/api/accumulator"
	"github.com/frankonly/upchain/log"
	"github.com/frankonly/upchain/storage"
)

// API names
const (
	apiAppend            = "Append"
	apiGet               = "Get"
	apiSearch            = "Search"
	apiGetDigest         = "GetDigest"
	apiGetProofByID      = "GetProofByID"
	apiGetProofByHash    = "GetProofByHash"
	apiGetOldProofByID   = "GetOldProofByID"
	apiGetOldProofByHash = "GetOldProofByHash"
)

// Server implements API server
type Server struct {
	pb.UnimplementedAccumulatorServer
	accumulator storage.MerkleAccumulator
	logger      *zap.SugaredLogger
}

// NewServer returns a new API server
func NewServer(accumulator storage.MerkleAccumulator, logger *zap.SugaredLogger) *Server {
	return &Server{accumulator: accumulator, logger: logger}
}

// Append appends new hash to accumulator
func (s Server) Append(_ context.Context, hash *pb.Hash) (*pb.ID, error) {
	hashLog := hex.EncodeToString(hash.Hash)
	s.infoRequest(apiAppend, "Hash", hashLog)

	// TODO: check length of hash
	id, err := s.accumulator.Append(hash.Hash)
	if err != nil {
		s.infoError(apiAppend, "hash", hashLog, "Error", "failed to append new hash")
		return nil, status.Error(codes.Internal, "failed to append new hash")
	}
	s.infoResponse(apiAppend, "hash", hashLog, "ID", id)
	return &pb.ID{Id: id}, nil
}

// Get gets certain hash by id from accumulator
func (s Server) Get(_ context.Context, id *pb.ID) (*pb.Hash, error) {
	s.infoRequest(apiGet, "ID", id.Id)

	hash, err := s.accumulator.Get(id.Id)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrOutOfRange):
			err = status.Error(codes.OutOfRange, err.Error())
		case errors.Is(err, storage.ErrNotFound):
			err = status.Error(codes.NotFound, err.Error())
		default:
			err = status.Error(codes.Internal, err.Error())
		}

		s.infoError(apiGet, "id", id.Id, "Error", err)
		return nil, err
	}

	s.infoResponse(apiGet, "id", id.Id, "Hash", hex.EncodeToString(hash))
	return &pb.Hash{Hash: hash}, nil
}

// Search searches accumulator and returns id of oldest node related to input hash
func (s Server) Search(_ context.Context, hash *pb.Hash) (*pb.ID, error) {
	hashLog := hex.EncodeToString(hash.Hash)
	s.infoRequest(apiSearch, "Hash", hashLog)

	id, err := s.accumulator.Search(hash.Hash)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrNotFound):
			err = status.Error(codes.NotFound, err.Error())
		default:
			err = status.Error(codes.Internal, err.Error())
		}

		s.infoError(apiSearch, "hash", hashLog, "Error", err)
		return nil, err
	}

	s.infoResponse(apiSearch, "hash", hashLog, "ID", id)
	return &pb.ID{Id: id}, nil
}

// GetDigest requests latest digest from accumulator
func (s Server) GetDigest(context.Context, *pb.Empty) (*pb.Hash, error) {
	s.infoRequest(apiGetDigest)

	digest, err := s.accumulator.Digest()
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrEmpty):
			err = status.Error(codes.Unavailable, err.Error())
		default:
			err = status.Error(codes.Internal, err.Error())
		}

		s.infoError(apiGetDigest, "Error", err)
		return nil, err
	}

	s.infoResponse(apiGetDigest, "Digest", hex.EncodeToString(digest))
	return &pb.Hash{Hash: digest}, nil
}

// GetProofByID requests hash proof of certain node to latest digest by id
func (s Server) GetProofByID(_ context.Context, id *pb.ID) (*pb.HashProof, error) {
	s.infoRequest(apiGetProofByID, "ID", id.Id)

	p, err := s.getProofByID(id.Id, nil)
	if err != nil {
		s.infoError(apiGetProofByID, "id", id.Id, "Error", err)
		return nil, err
	}

	s.infoResponse(apiGetProofByID, "id", id.Id, "HashProof", log.HashProofLog(p))
	return p, nil
}

// GetProofByHash requests hash proof of certain node to latest digest by hash
func (s Server) GetProofByHash(_ context.Context, hash *pb.Hash) (*pb.HashProof, error) {
	hashLog := hex.EncodeToString(hash.Hash)
	s.infoRequest(apiGetProofByHash, "Hash", hashLog)

	id, err := s.accumulator.Search(hash.Hash)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrNotFound):
			err = status.Error(codes.NotFound, err.Error())
		default:
			err = status.Error(codes.Internal, err.Error())
		}

		s.infoError(apiGetProofByHash, "hash", hashLog, "Error", err)
		return nil, err
	}

	p, err := s.getProofByID(id, nil)
	if err != nil {
		s.infoError(apiGetProofByHash, "hash", hashLog, "Error", err)
		return nil, err
	}

	s.infoResponse(apiGetProofByHash, "hash", hashLog, "HashProof", log.HashProofLog(p))
	return p, nil
}

// GetOldProofByID requests hash proof of certain node to a past digest by id
func (s Server) GetOldProofByID(_ context.Context, in *pb.GetOldProofByIDRequest) (*pb.HashProof, error) {
	digestLog := hex.EncodeToString(in.Digest)
	s.infoRequest(apiGetOldProofByID, "ID", in.Id, "Digest", digestLog)

	p, err := s.getProofByID(in.Id, in.Digest)
	if err != nil {
		s.infoError(apiGetOldProofByID, "id", in.Id, "digest", digestLog, "Error", err)
		return nil, err
	}

	s.infoResponse(apiGetOldProofByID, "id", in.Id, "digest", digestLog, "HashProof", log.HashProofLog(p))
	return p, nil
}

// GetOldProofByHash requests hash proof of certain node to a past digest by hash
func (s Server) GetOldProofByHash(_ context.Context, in *pb.GetOldProofByHashRequest) (*pb.HashProof, error) {
	hashLog := hex.EncodeToString(in.Hash)
	digestLog := hex.EncodeToString(in.Digest)
	s.infoRequest(apiGetOldProofByHash, "Hash", hashLog, "Digest", digestLog)

	id, err := s.accumulator.Search(in.Hash)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrNotFound):
			err = status.Error(codes.NotFound, err.Error())
		default:
			err = status.Error(codes.Internal, err.Error())
		}

		s.infoError(apiGetOldProofByHash, "hash", hashLog, "digest", digestLog, "Error", err)
		return nil, err
	}

	p, err := s.getProofByID(id, in.Digest)
	if err != nil {
		s.infoError(apiGetOldProofByHash, "hash", hashLog, "digest", digestLog, "Error", err)
		return nil, err
	}

	s.infoResponse(apiGetOldProofByHash, "hash", hashLog, "digest", digestLog, "HashProof", log.HashProofLog(p))
	return p, nil
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

func (s Server) infoRequest(keysAndValues ...interface{}) {
	keysAndValues = append([]interface{}{"API"}, keysAndValues...)
	s.logger.Infow("grpc request", keysAndValues...)
}

func (s Server) infoResponse(keysAndValues ...interface{}) {
	keysAndValues = append([]interface{}{"API"}, keysAndValues...)
	s.logger.Infow("grpc response", keysAndValues...)
}

func (s Server) infoError(keysAndValues ...interface{}) {
	keysAndValues = append([]interface{}{"API"}, keysAndValues...)
	s.logger.Infow("grpc error", keysAndValues...)
}
