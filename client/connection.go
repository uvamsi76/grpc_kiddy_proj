package main

import (
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

func newConnection(addr string) *grpc.ClientConn {
	conn, err := grpc.NewClient(
		addr,

		// ── Transport ──────────────────────────────────────
		// No TLS for now (we add this in Module 8)
		grpc.WithTransportCredentials(insecure.NewCredentials()),

		// ── Keep-alive ─────────────────────────────────────
		// Prevents idle connections from being silently dropped
		// by firewalls or load balancers
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second, // send a ping every 10s if idle
			Timeout:             5 * time.Second,  // wait 5s for ping response
			PermitWithoutStream: true,             // ping even with no active RPCs
		}),

		// ── Backoff / Retry on connect ──────────────────────
		// If server isn't up yet, retry with exponential backoff
		grpc.WithBlock(), // wait until connected (useful in dev)
	)
	if err != nil {
		log.Fatalf("failed to connect to %s: %v", addr, err)
	}

	return conn
}
