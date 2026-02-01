package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/seetadev/touchcalc-example-libp2p/api"
	"github.com/seetadev/touchcalc-example-libp2p/constants"
	"github.com/seetadev/touchcalc-example-libp2p/relay"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h, err := relay.NewRelayHost(ctx)
	if err != nil {
		log.Fatalf("Failed to create host: %v", err)
	}

	psManager, err := relay.NewPubSubManager(ctx, h)
	if err != nil {
		log.Fatalf("Failed to create PubSub manager: %v", err)
	}

	_, err = psManager.JoinTopic(constants.DiscoveryTopic)
	if err != nil {
		log.Fatalf("Failed to join discovery topic: %v", err)
	}

	_, err = psManager.JoinTopic(constants.SpreadsheetTopic)
	if err != nil {
		log.Fatalf("Failed to join spreadsheet topic: %v", err)
	}

	go api.StartAPIServer(h, constants.HTTPAPIPort)

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		for {
			select {
			case <-ticker.C:
				peers := h.Network().Peers()
				if len(peers) > 0 {
					fmt.Println("\n--- [DEBUG] Known Peer Addresses ---")
					for _, p := range peers {
						addrs := h.Peerstore().Addrs(p)
						fmt.Printf("Peer %s: %v\n", p.String()[:8], addrs)
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	fmt.Println("--- Relay Node Started ---")
	fmt.Printf("Peer ID: %s\n", h.ID().String())
	fmt.Println("P2P Multiaddresses:")
	for _, addr := range h.Addrs() {
		fmt.Printf("  %s/p2p/%s\n", addr, h.ID())
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	fmt.Println("Shutting down...")
	h.Close()
}
