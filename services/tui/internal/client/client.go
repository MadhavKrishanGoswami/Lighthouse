// Package client provides functionality to connect to the gRPC server
package client

import (
	"log"
	"time"

	tui "github.com/MadhavKrishanGoswami/Lighthouse/services/common/genproto/tui"
	"github.com/MadhavKrishanGoswami/Lighthouse/services/tui/internal/config"

	"google.golang.org/grpc"

	"google.golang.org/grpc/credentials/insecure"
)

// StartClient waits for the server to be ready using the new grpc.NewClient API.

func StartClient(cfg config.Config) (tui.TUIServiceClient, *grpc.ClientConn, error) {
	addr := cfg.OrchestratorAddr

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	var conn *grpc.ClientConn

	var err error

	deadline := time.Now().Add(10 * time.Second)
	for {
		conn, err = grpc.NewClient(addr, opts...)
		if err == nil {
			break
		}
		if time.Now().After(deadline) {
			return nil, nil, err
		}
		log.Printf("Waiting for gRPC server at %s... retrying in 2s (err=%v)", addr, err)
		time.Sleep(2 * time.Second)
	}

	client := tui.NewTUIServiceClient(conn)

	log.Printf("gRPC client connected to orchestrator at %s", addr)

	return client, conn, nil
}
