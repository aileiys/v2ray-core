// +build linux

package tcp

import (
	"net"
	"syscall"

	v2net "v2ray.com/core/common/net"
	"v2ray.com/core/transport/internet"
	"v2ray.com/core/transport/internet/internal"
)

const SO_ORIGINAL_DST = 80

func GetOriginalDestination(conn internet.Connection) (v2net.Destination, error) {
	fd, err := internal.GetSysFd(conn.(net.Conn))
	if err != nil {
		return v2net.Destination{}, newError("failed to get original destination").Base(err)
	}

	addr, err := syscall.GetsockoptIPv6Mreq(fd, syscall.IPPROTO_IP, SO_ORIGINAL_DST)
	if err != nil {
		return v2net.Destination{}, newError("failed to call getsockopt").Base(err)
	}
	ip := v2net.IPAddress(addr.Multiaddr[4:8])
	port := uint16(addr.Multiaddr[2])<<8 + uint16(addr.Multiaddr[3])
	return v2net.TCPDestination(ip, v2net.Port(port)), nil
}
