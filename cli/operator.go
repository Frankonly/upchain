package cli

import (
	"context"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	pb "upchain/api/accumulator"
	"upchain/crypto"
)

var (
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
				fmt.Println("Transaction ID:", id.Id)
			}

			return err
		},
	}

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

	searchCmd = &cobra.Command{
		Use:   "search HASH",
		Short: "Get id from upchain server by transaction hash",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			hash, err := hex.DecodeString(args[0])
			if err != nil {
				return fmt.Errorf("invalid hash in hex: %w", err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
			defer cancel()

			id, err := Client().Search(ctx, &pb.Hash{Hash: hash})
			if err == nil {
				fmt.Println(id.Id)
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
		Use:   "proof (HASH|ID)",
		Short: "Get hash proof of certain transaction from upchain server",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var hashProof *pb.HashProof

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
			defer cancel()

			hash, err := hex.DecodeString(args[0])
			if err != nil {
				id, err2 := strconv.ParseUint(args[0], 10, 64)
				if err2 != nil {
					return fmt.Errorf("invalid input %s, need uint64 or hex string", args[0])
				}

				hashProof, err = Client().GetProofByID(ctx, &pb.ID{Id: id})
			} else {
				hashProof, err = Client().GetProofByHash(ctx, &pb.Hash{Hash: hash})
			}

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

	registerCmd = &cobra.Command{
		Use:   "register [FILE]",
		Short: "Register a file to upchain server by saving its hash for proof",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fileByte, err := ioutil.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("invalid file path %s: %w", args[0], err)
			}

			hash := crypto.Hash(fileByte)
			fmt.Println("Hash:", hex.EncodeToString(hash))

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
			defer cancel()

			id, err := Client().Append(ctx, &pb.Hash{Hash: hash})
			if err == nil {
				fmt.Println("Transaction ID:", id.Id)
			}

			return err
		},
	}
)
