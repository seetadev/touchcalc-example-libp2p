package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/libp2p/go-libp2p/core/host"
)

type AddressResponse struct {
	Websocket    []string `json:"websocket"`
	WebRTCDirect []string `json:"webrtcDirect"`
	TCP          []string `json:"tcp"`
	All          []string `json:"all"`
}

func StartAPIServer(h host.Host, port string) {
	http.HandleFunc("/api/addresses", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("[API] Address request from %s\n", r.RemoteAddr)

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		var resp AddressResponse
		peerID := h.ID().String()

		for _, addr := range h.Addrs() {
			fullAddr := fmt.Sprintf("%s/p2p/%s", addr.String(), peerID)
			resp.All = append(resp.All, fullAddr)

			// Categorize addresses based on transport protocols
			if strings.Contains(fullAddr, "/ws") {
				resp.Websocket = append(resp.Websocket, fullAddr)
			} else if strings.Contains(fullAddr, "/webrtc-direct") {
				resp.WebRTCDirect = append(resp.WebRTCDirect, fullAddr)
			} else if strings.Contains(fullAddr, "/tcp") && !strings.Contains(fullAddr, "/ws") {
				resp.TCP = append(resp.TCP, fullAddr)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	fmt.Printf("HTTP API listening on: http://0.0.0.0:%s\n", port)
	if err := http.ListenAndServe("0.0.0.0:"+port, nil); err != nil {
		fmt.Printf("HTTP server failed: %v\n", err)
	}
}
