package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type netbirdError struct {
	cause  error
	output string
}

func (e *netbirdError) Error() string {
	return fmt.Sprintf("%v: %s", e.cause, strings.TrimSpace(e.output))
}

func runNetbird(args ...string) error {
	cmd := exec.Command("netbird", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return &netbirdError{cause: err, output: string(out)}
	}
	return nil
}

func isAuthError(err error) bool {
	if err == nil {
		return false
	}
	ne, ok := err.(*netbirdError)
	if !ok {
		return false
	}
	lower := strings.ToLower(ne.output)
	return strings.Contains(lower, "unauthenticated") ||
		strings.Contains(lower, "login required") ||
		strings.Contains(lower, "authentication") ||
		strings.Contains(lower, "not logged")
}

type netbirdStatusOutput struct {
	LocalPeerState struct {
		IP   string `json:"IP"`
		FQDN string `json:"fqdn"`
	} `json:"localPeerState"`
}

type netbirdStatus struct {
	IP       string
	FQDN     string
	TunIface string
}

func getNetbirdStatus() (*netbirdStatus, error) {
	out, err := exec.Command("netbird", "status", "--json").Output()
	if err != nil {
		return nil, err
	}
	var s netbirdStatusOutput
	if err := json.Unmarshal(out, &s); err != nil {
		return nil, err
	}
	ip := strings.SplitN(s.LocalPeerState.IP, "/", 2)[0]
	return &netbirdStatus{
		IP:       ip,
		FQDN:     s.LocalPeerState.FQDN,
		TunIface: "wt0",
	}, nil
}
