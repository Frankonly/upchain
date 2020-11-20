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
)
