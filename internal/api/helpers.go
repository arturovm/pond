package api

import (
	"encoding/hex"
	"encoding/json"
	"net"
	"net/http"
	"net/netip"
)

func parseClientIP(r *http.Request) netip.Addr {
	var ip netip.Addr
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		ip, _ = netip.ParseAddr(host)
	}
	return ip
}

func writeTokenResponse(w http.ResponseWriter, token []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(struct {
		SessionToken string `json:"session_token"`
	}{SessionToken: hex.EncodeToString(token)})
}
