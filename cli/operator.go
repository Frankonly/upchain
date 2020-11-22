package cli

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	pb "upchain/api/accumulator"
)

var (
	getCmd = &cobra.Command{
		Use:   "get ID",
		Short: "Get hash from upchain server by transaction id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid id %s: %w", args[0], err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
			defer cancel()

			hash, err := Client().Get(ctx, &pb.ID{Id: id})
			if err == nil {
				fmt.Println(hex.EncodeToString(hash.Hash))
			}

			return err
		},
	}

	appendCmd = &cobra.Command{
		Use:   "append HASH",
		Short: "Append hash to upchain server",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			hash, err := hex.DecodeString(args[0])
			if err != nil {
				return fmt.Errorf("invalid hash in hex: %w", err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
			defer cancel()

			id, err := Client().Append(ctx, &pb.Hash{Hash: hash})
			if err == nil {
				fmt.Println("transaction ID:", id.Id)
			}

			return err
		},
	}

	digestCmd = &cobra.Command{
		Use:   "digest",
		Short: "Get digest of merkle accumulator from upchain server",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
			defer cancel()

			hash, err := Client().GetDigest(ctx, &pb.Empty{})
			if err == nil {
				fmt.Println(hex.EncodeToString(hash.Hash))
			}

			return err
		},
	}

	proofCmd = &cobra.Command{
		Use:   "proof ID",
		Short: "Get hash proof of certain transaction from upchain server",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid id %s: %w", args[0], err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
			defer cancel()

			hashProof, err := Client().GetProofByID(ctx, &pb.ID{Id: id})
			if err == nil {
				fmt.Println("Hash:", hex.EncodeToString(hashProof.Hash))
				fmt.Println("Digest:", hex.EncodeToString(hashProof.Digest))

				path := make([]string, 0, len(hashProof.Path))
				for _, hash := range hashProof.Path {
					path = append(path, hex.EncodeToString(hash))
				}
				fmt.Println("HashPath:", path)
			}

			return err
		},
	}
)
