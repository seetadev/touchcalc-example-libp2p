package relay

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	rcmgr "github.com/libp2p/go-libp2p/p2p/host/resource-manager"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
	webrtc "github.com/libp2p/go-libp2p/p2p/transport/webrtc"
	"github.com/libp2p/go-libp2p/p2p/transport/websocket"

	"github.com/seetadev/touchcalc-example-libp2p/constants"
)

func loadOrCreateIdentity(keyPath string) (crypto.PrivKey, error) {
	keyData, err := os.ReadFile(keyPath)
	if err == nil {
		keyBytes, err := base64.StdEncoding.DecodeString(string(keyData))
		if err != nil {
			return nil, fmt.Errorf("failed to decode key: %w", err)
		}
		priv, err := crypto.UnmarshalPrivateKey(keyBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal key: %w", err)
		}
		fmt.Printf("[Identity] Loaded existing key from %s\n", keyPath)
		return priv, nil
	}

	priv, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}

	keyBytes, err := crypto.MarshalPrivateKey(priv)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal key: %w", err)
	}

	if err := os.MkdirAll("keys", 0755); err != nil {
		return nil, fmt.Errorf("failed to create keys directory: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString(keyBytes)
	if err := os.WriteFile(keyPath, []byte(encoded), 0600); err != nil {
		return nil, fmt.Errorf("failed to save key: %w", err)
	}

	fmt.Printf("[Identity] Generated new key and saved to %s\n", keyPath)
	return priv, nil
}

func NewRelayHost(ctx context.Context) (host.Host, error) {
	privKey, err := loadOrCreateIdentity(constants.DefaultKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load identity: %w", err)
	}

	rm, _ := rcmgr.NewResourceManager(rcmgr.NewFixedLimiter(rcmgr.InfiniteLimits))

	addrs := []string{
		"/ip4/0.0.0.0/tcp/9091",
		"/ip4/0.0.0.0/tcp/9092/ws",
		"/ip4/0.0.0.0/udp/9093/webrtc-direct",
	}

	h, err := libp2p.New(
		libp2p.Identity(privKey),
		libp2p.ListenAddrStrings(addrs...),
		// libp2p.Transport(tcp.NewTCPTransport),
		libp2p.Transport(websocket.New),
		libp2p.Transport(webrtc.New),
		libp2p.ResourceManager(rm),
		// libp2p.ForceReachabilityPublic(),
		libp2p.EnableNATService(),
		libp2p.EnableRelayService(),
		libp2p.EnableHolePunching(),
	)
	if err != nil {
		return nil, err
	}

	go monitorEvents(ctx, h)

	_, err = relay.New(h)
	return h, err
}

func monitorEvents(ctx context.Context, h host.Host) {
	sub, err := h.EventBus().Subscribe([]interface{}{
		new(event.EvtPeerIdentificationCompleted),
		new(event.EvtPeerConnectednessChanged),
		new(event.EvtLocalReachabilityChanged),
	})
	if err != nil {
		fmt.Printf("Failed to subscribe to events: %v\n", err)
		return
	}
	defer sub.Close()

	for {
		select {
		case evt := <-sub.Out():
			switch e := evt.(type) {
			case event.EvtLocalReachabilityChanged:
				fmt.Printf("[Reachability] Status changed to: %s\n", e.Reachability.String())

			case event.EvtPeerIdentificationCompleted:
				fmt.Printf("[Identify] Completed for peer %s\n", e.Peer.String()[:8])

			case event.EvtPeerConnectednessChanged:
				peerIdShort := e.Peer.String()[:8]
				switch e.Connectedness {
				case network.Connected:
					fmt.Printf("[Network] New connection from %s\n", peerIdShort)
				case network.NotConnected:
					fmt.Printf("[Network] Disconnected from %s\n", peerIdShort)
				}
			}
		case <-ctx.Done():
			return
		}
	}
}
