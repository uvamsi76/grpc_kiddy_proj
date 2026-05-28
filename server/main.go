// server/main.go
package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "github.com/uvamsi76/grpc_kiddy_proj/gen/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	port := "50051"

	// ── 1. Create TCP listener ───────────────────────────
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("failed to listen on port %s: %v", port, err)
	}

	// ── 2. Create gRPC server ────────────────────────────
	grpcServer := grpc.NewServer(
	// interceptors go here in Module 7
	)

	// ── 3. Register your service ─────────────────────────
	pb.RegisterUserServiceServer(grpcServer, NewUserServer())

	// ── 4. Enable reflection (lets grpcurl discover services) ──
	reflection.Register(grpcServer)

	// ── 5. Graceful shutdown ─────────────────────────────
	// Listen for SIGINT / SIGTERM (Ctrl+C or kill)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("Shutting down gRPC server...")
		grpcServer.GracefulStop() // finish in-flight RPCs, then stop
	}()

	// ── 6. Start serving ─────────────────────────────────
	log.Printf("gRPC server listening on :%s", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
