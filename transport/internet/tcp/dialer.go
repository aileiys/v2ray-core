package tcp

import (
	"context"
	"crypto/tls"

	"v2ray.com/core/app/log"
	"v2ray.com/core/common"
	v2net "v2ray.com/core/common/net"
	"v2ray.com/core/transport/internet"
	v2tls "v2ray.com/core/transport/internet/tls"
)

func Dial(ctx context.Context, dest v2net.Destination) (internet.Connection, error) {
	log.Trace(newError("dailing TCP to ", dest))
	src := internet.DialerSourceFromContext(ctx)

	tcpSettings := internet.TransportSettingsFromContext(ctx).(*Config)

	conn, err := internet.DialSystem(ctx, src, dest)
	if err != nil {
		return nil, err
	}
	if securitySettings := internet.SecuritySettingsFromContext(ctx); securitySettings != nil {
		tlsConfig, ok := securitySettings.(*v2tls.Config)
		if ok {
			config := tlsConfig.GetTLSConfig()
			if dest.Address.Family().IsDomain() {
				config.ServerName = dest.Address.Domain()
			}
			conn = tls.Client(conn, config)
		}
	}
	if tcpSettings.HeaderSettings != nil {
		headerConfig, err := tcpSettings.HeaderSettings.GetInstance()
		if err != nil {
			return nil, newError("failed to get header settings").Base(err).AtError()
		}
		auth, err := internet.CreateConnectionAuthenticator(headerConfig)
		if err != nil {
			return nil, newError("failed to create header authenticator").Base(err).AtError()
		}
		conn = auth.Client(conn)
	}
	return internet.Connection(conn), nil
}

func init() {
	common.Must(internet.RegisterTransportDialer(internet.TransportProtocol_TCP, Dial))
}
