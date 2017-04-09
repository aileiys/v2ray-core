package websocket_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	v2net "v2ray.com/core/common/net"
	"v2ray.com/core/testing/assert"
	tlsgen "v2ray.com/core/testing/tls"
	"v2ray.com/core/transport/internet"
	v2tls "v2ray.com/core/transport/internet/tls"
	. "v2ray.com/core/transport/internet/websocket"
)

func Test_listenWSAndDial(t *testing.T) {
	assert := assert.On(t)
	conns := make(chan internet.Connection, 16)
	listen, err := ListenWS(internet.ContextWithTransportSettings(context.Background(), &Config{
		Path: "ws",
	}), v2net.DomainAddress("localhost"), 13146, conns)
	assert.Error(err).IsNil()
	go func() {
		for conn := range conns {
			go func(c internet.Connection) {
				defer c.Close()

				var b [1024]byte
				n, err := c.Read(b[:])
				//assert.Error(err).IsNil()
				if err != nil {
					return
				}
				assert.Bool(bytes.HasPrefix(b[:n], []byte("Test connection"))).IsTrue()

				_, err = c.Write([]byte("Response"))
				assert.Error(err).IsNil()
			}(conn)
		}
	}()

	ctx := internet.ContextWithTransportSettings(context.Background(), &Config{Path: "ws"})
	conn, err := Dial(ctx, v2net.TCPDestination(v2net.DomainAddress("localhost"), 13146))

	assert.Error(err).IsNil()
	_, err = conn.Write([]byte("Test connection 1"))
	assert.Error(err).IsNil()

	var b [1024]byte
	n, err := conn.Read(b[:])
	assert.Error(err).IsNil()
	assert.String(string(b[:n])).Equals("Response")

	assert.Error(conn.Close()).IsNil()
	<-time.After(time.Second * 5)
	conn, err = Dial(ctx, v2net.TCPDestination(v2net.DomainAddress("localhost"), 13146))
	assert.Error(err).IsNil()
	_, err = conn.Write([]byte("Test connection 2"))
	assert.Error(err).IsNil()
	n, err = conn.Read(b[:])
	assert.Error(err).IsNil()
	assert.String(string(b[:n])).Equals("Response")
	assert.Error(conn.Close()).IsNil()
	<-time.After(time.Second * 15)
	conn, err = Dial(ctx, v2net.TCPDestination(v2net.DomainAddress("localhost"), 13146))
	assert.Error(err).IsNil()
	_, err = conn.Write([]byte("Test connection 3"))
	assert.Error(err).IsNil()
	n, err = conn.Read(b[:])
	assert.Error(err).IsNil()
	assert.String(string(b[:n])).Equals("Response")
	assert.Error(conn.Close()).IsNil()

	assert.Error(listen.Close()).IsNil()

	close(conns)
}

func Test_listenWSAndDial_TLS(t *testing.T) {
	assert := assert.On(t)
	go func() {
		<-time.After(time.Second * 5)
		assert.Fail("Too slow")
	}()

	ctx := internet.ContextWithTransportSettings(context.Background(), &Config{
		Path: "wss",
	})
	ctx = internet.ContextWithSecuritySettings(ctx, &v2tls.Config{
		AllowInsecure: true,
		Certificate:   []*v2tls.Certificate{tlsgen.GenerateCertificateForTest()},
	})
	conns := make(chan internet.Connection, 16)
	listen, err := ListenWS(ctx, v2net.DomainAddress("localhost"), 13143, conns)
	assert.Error(err).IsNil()
	go func() {
		conn := <-conns
		conn.Close()
		listen.Close()
	}()

	conn, err := Dial(ctx, v2net.TCPDestination(v2net.DomainAddress("localhost"), 13143))
	assert.Error(err).IsNil()
	conn.Close()
}
