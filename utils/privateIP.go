package utils

import (
	"net"
)

var tenNet *net.IPNet

func init() {
	_, tenNet, _ = net.ParseCIDR("10.0.0.0/8")
}

func Is10PrivateIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}
	return tenNet.Contains(ip)
}
