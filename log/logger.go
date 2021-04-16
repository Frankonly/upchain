package log

import (
	"encoding/hex"

	"go.uber.org/zap"

	pb "github.com/frankonly/upchain/api/accumulator"
)

var logger *zap.SugaredLogger

// New returns the same logger all the time
func New() *zap.SugaredLogger {
	if logger != nil {
		return logger
	}

	cfg := zap.NewDevelopmentConfig()
	cfg.Encoding = "json"
	cfg.OutputPaths = []string{"stdout", "log/upchain.log"}
	cfg.ErrorOutputPaths = []string{"stderr", "log/upchain.log"}
	base, err := cfg.Build()
	if err != nil {
		panic(err)
	}

	logger = base.Sugar()
	return logger
}

type HashProof struct {
	Hash   string
	Digest string
	Path   []string
}

func HashProofLog(p *pb.HashProof) HashProof {
	proof := HashProof{
		Hash:   hex.EncodeToString(p.Hash),
		Digest: hex.EncodeToString(p.Digest),
		Path:   make([]string, 0, len(p.Path)),
	}
	for _, hash := range p.Path {
		proof.Path = append(proof.Path, hex.EncodeToString(hash))
	}
	return proof
}
