package utils

import (
	"errors"
	"net"
)

func IpToInt(ip net.IP) (int32, error) {
	if len(ip) != net.IPv4len {
		return 0, errors.New("ipv4 must be provided")
	}
	result := int32(0)
	for i := range ip {
		result = (result * int32(1<<8)) + int32(ip[i])
	}
	// slog.Infof("source ip %v, result %v\n", ip.String(), result)
	return result, nil
}
