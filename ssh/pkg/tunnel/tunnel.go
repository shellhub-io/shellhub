package tunnel

import (
	"bufio"
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
	"github.com/multiformats/go-multiaddr"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	log "github.com/sirupsen/logrus"
)

var (
	ErrWebEndpointForbidden      = errors.New("web endpoint not found")
	ErrDeviceTunnelDial          = errors.New("failed to connect to device")
	ErrDeviceTunnelWriteRequest  = errors.New("failed to send data to the device")
	ErrDeviceTunnelReadResponse  = errors.New("failed to write the response back to the client")
	ErrDeviceTunnelHijackRequest = errors.New("failed to capture the request")
	ErrDeviceTunnelParsePath     = errors.New("failed to parse the path")
	ErrDeviceTunnelConnect       = errors.New("failed to connect to the port on device")
)

type Message struct {
	Message string `json:"message"`
}

func NewMessageFromError(err error) Message {
	return Message{
		Message: err.Error(),
	}
}

type Config struct {
	// Tunnels defines if tunnel's feature is enabled.
	Tunnels bool
	// TunnelsDomain define the domain of tunnels feature when it's enabled.
	TunnelsDomain string
	// RedisURI is the redis URI connection.
	RedisURI string
}

func (c Config) Validate() error {
	if c.Tunnels && c.TunnelsDomain == "" {
		return errors.New("tunnels feature is enabled, but tunnel's domain is empty")
	}

	if c.RedisURI == "" {
		return errors.New("redis uri is empty")
	}

	return nil
}

type filter struct{}

// AllowConnect implements relay.ACLFilter.
func (f *filter) AllowConnect(src peer.ID, srcAddr multiaddr.Multiaddr, dest peer.ID) bool {
	// allow only the relay to connect
	return false
}

// AllowReserve implements relay.ACLFilter.
func (f *filter) AllowReserve(p peer.ID, a multiaddr.Multiaddr) bool {
	// allow only if device accepted in a namespace
	return true
}

type Foo struct {
	relay host.Host
	M     *map[string]peer.ID
}

func NewFoo() *Foo {
	m := make(map[string]peer.ID, 0)

	pemData, err := os.ReadFile("/tmp/private_key.pem")
	if err != nil {
		fmt.Println("Error reading PEM file:", err)
		panic(err)
	}

	block, _ := pem.Decode(pemData)
	if block == nil || block.Type != "PRIVATE KEY" {
		log.Fatalf("failed to decode PEM block containing private key")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		log.Fatalf("unable to parse PKCS#8 private key: %v", err)
	}

	// Assert it is an RSA key
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		log.Fatalf("not an RSA private key")
	}

	privKey, err := crypto.UnmarshalRsaPrivateKey(x509.MarshalPKCS1PrivateKey(rsaKey))
	if err != nil {
		fmt.Println("Error unmarshaling RSA private key:", err)
		panic(err)
	}

	listenAddr, _ := multiaddr.NewMultiaddr("/ip4/0.0.0.0/tcp/9000/ws")
	relay1, err := libp2p.New(
		libp2p.Identity(privKey),
		libp2p.ListenAddrs(listenAddr),
	)
	if err != nil {
		log.Printf("Failed to create relay1: %v", err)

		return nil
	}

	// Configure the host to offer the circuit relay service.
	// Any host that is directly dialable in the network (or on the internet)
	// can offer a circuit relay service, this isn't just the job of
	// "dedicated" relay services.
	// In circuit relay v2 (which we're using here!) it is rate limited so that
	// any node can offer this service safely
	_, err = relay.New(relay1, relay.WithACL(&filter{}))
	if err != nil {
		log.Printf("Failed to instantiate the relay: %v", err)
	}

	relay1info := peer.AddrInfo{
		ID:    relay1.ID(),
		Addrs: relay1.Addrs(),
	}

	relay1.SetStreamHandler("/register/1.0.0", func(s network.Stream) {
		fmt.Println("Got a new stream!")

		for _, p := range relay1.Network().Peers() {
			fmt.Println("Connected peer:", p)
		}

		var buffer [512]byte

		d, err := s.Read(buffer[:])
		if err != nil {
			fmt.Println("Error reading from buffer:", err)
		}

		data := Data{}
		if err := json.Unmarshal(buffer[:d], &data); err != nil {
			fmt.Println("Error unmarshaling JSON:", err)
		}

		mapKey := fmt.Sprintf("%s/%s", data.Namespace, data.UID)
		fmt.Println("Map key:", mapKey)
		mapValue := s.Conn().RemotePeer()
		fmt.Println("Map value (peer ID):", mapValue)

		m[mapKey] = mapValue

		fmt.Printf("Updated connections map: %v\n", m)

		fmt.Printf("Received data: Namespace=%s, UID=%s\n", data.Namespace, data.UID)

		s.Close()
	})

	for _, addr := range relay1info.Addrs {
		fmt.Printf("Relay address: %s/p2p/%s\n", addr, relay1info.ID)
	}

	return &Foo{
		relay: relay1,
		M:     &m,
	}
}

func (f *Foo) Dial(ctx context.Context, key string) (net.Conn, error) {
	fmt.Println("Dialing to key:", key)

	ctx, _ = context.WithTimeout(context.Background(), 30*time.Hour)

	fmt.Println("Instance:", f)

	m := *f.M
	fmt.Println("Current connections map:", m)

	id := m[key]

	fmt.Println("Dialing to ID:", id)
	fmt.Println("Peer store", f.relay.Peerstore().Peers())

	// /ip4/<relay-ip>/tcp/<relay-port>/p2p/<relay-peer-id>/p2p-circuit/p2p/<destination-peer-id>
	maddrStr := fmt.Sprintf("%s/p2p/%s/p2p-circuit/p2p/%s", f.relay.Addrs()[0], f.relay.ID(), id)
	maddr, err := multiaddr.NewMultiaddr(maddrStr)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Multiaddr:", maddr.String())

	pi, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		log.Fatal(err)
	}

	err = f.relay.Connect(ctx, *pi)
	fmt.Println("Connect to peer", err)

	s, err := f.relay.NewStream(ctx, id, "/ssh/1.0.0")
	if err != nil {
		fmt.Println("Dialing to ID:", id)
		log.Println("h.NewStream err:", err)

		return nil, err
	}

	/*
		h, _ := libp2p.New(
			libp2p.NoListenAddrs,
			libp2p.EnableRelay(),
		)

		ma, err := multiaddr.NewMultiaddr("/ip4/172.18.0.6/tcp/9000/ws/p2p/QmXVv62vivDGCTwtA56Y4s3cYujnWvX4wDW2LvAQr3HAMJ")
		if err != nil {
			log.Println("multiaddr.NewMultiaddr err:", err)

			panic(err)
		}

		relay1info, err := peer.AddrInfoFromP2pAddr(ma)
		if err != nil {
			log.Println("peer.AddrInfoFromP2pAddr err:", err)

			panic(err)
		}

		err = h.Connect(ctx, *relay1info)
		if err != nil {
			log.Println("h.Connect err:", err)

			return nil, err
		}

		// /ip4/<relay-ip>/tcp/<relay-port>/p2p/<relay-peer-id>/p2p-circuit/p2p/<destination-peer-id>
		// relayaddr, err := multiaddr.NewMultiaddr("/p2p/" + f.relay.ID + "/p2p-circuit/p2p/" + id)
		// relayaddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/p2p/%s/p2p-circuit/p2p/%s", f.relay.ID(), id))

		// fmt.Println("Relay address:", f.relay.Addrs())

		// pi, err := peer.AddrInfoFromP2pAddr(maddr)
		// if err != nil {
		// 	log.Fatal(err)
		// }

		fmt.Println("First relay address: ", f.relay.Addrs()[1])

		relayAddrStr := fmt.Sprintf("%s/p2p/%s/p2p-circuit/p2p/%s", f.relay.Addrs()[1], f.relay.ID(), id)

		relayAddr, err := multiaddr.NewMultiaddr(relayAddrStr)
		if err != nil {
			log.Println("multiaddr.NewMultiaddr err:", err)

			return nil, err
		}

		relayInfo, err := peer.AddrInfoFromP2pAddr(relayAddr)
		if err != nil {
			log.Println("peer.AddrInfoFromP2pAddr err:", err)

			return nil, err
		}

		fmt.Println("Relay info:", relayInfo.String())

		err = h.Connect(ctx, *relayInfo)
		fmt.Println("Connect to peer", err)

		fmt.Println(id)
		fmt.Println(id)
		fmt.Println(id)
		fmt.Println(id)

		s, err := h.NewStream(network.WithAllowLimitedConn(ctx, "/ssh/1.0.0"), id, "/ssh/1.0.0")
		if err != nil {
			log.Println("h.NewStream err:", err)

			return nil, err
		}*/

	return NewStreamConn(s), nil
}

type Tunnel struct {
	Tunnel *Foo
	API    internalclient.Client
	router *echo.Echo
}

type Data struct {
	Namespace string `json:"namespace"`
	UID       string `json:"uid"`
}

func NewTunnel(t *Foo, connection string, dial string, config Config) (*Tunnel, error) {
	api, err := internalclient.NewClient(internalclient.WithAsynqWorker(config.RedisURI))
	if err != nil {
		return nil, err
	}

	fmt.Println("NewTunnel called with Foo instance:", t)

	tunnel := &Tunnel{
		Tunnel: t,
		API:    api,
		// m:      &m,
	}

	// tunnel.Tunnel.Listen()

	// tunnel.h = relay1

	// select {}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	tunnel.router = echo.New()

	// tunnel.Tunnel.ConnectionHandler = func(request *http.Request) (string, error) {
	// 	tenant := request.Header.Get("X-Tenant-ID")
	// 	uid := request.Header.Get("X-Device-UID")

	// 	// WARN:
	// 	// In versions before 0.15, the agent's authentication may not provide the "X-Tenant-ID" header.
	// 	// This can cause issues with establishing sessions and tracking online devices. To solve this,
	// 	// we retrieve the tenant ID by querying the API. Maybe this can be removed in a future release.
	// 	if tenant == "" {
	// 		device, err := tunnel.API.GetDevice(uid)
	// 		if err != nil {
	// 			log.WithError(err).
	// 				WithField("uid", uid).
	// 				Error("unable to retrieve device's tenant id")

	// 			return "", err
	// 		}

	// 		tenant = device.TenantID
	// 	}

	// 	return tenant + ":" + uid, nil
	// }
	// tunnel.Tunnel.CloseHandler = func(key string) {
	// 	parts := strings.Split(key, ":")
	// 	if len(parts) != 2 {
	// 		log.Error("failed to parse key at close handler")

	// 		return
	// 	}

	// 	tenant := parts[0]
	// 	uid := parts[1]

	// 	if err := tunnel.API.DevicesOffline(uid); err != nil {
	// 		log.WithError(err).
	// 			WithFields(log.Fields{
	// 				"uid":       uid,
	// 				"tenant_id": tenant,
	// 			}).
	// 			Error("failed to set device offline")
	// 	}
	// }
	// tunnel.Tunnel.KeepAliveHandler = func(key string) {
	// 	parts := strings.Split(key, ":")
	// 	if len(parts) != 2 {
	// 		log.Error("failed to parse key at keep alive handler")

	// 		return
	// 	}

	// 	tenant := parts[0]
	// 	uid := parts[1]

	// 	if err := tunnel.API.DevicesHeartbeat(uid); err != nil {
	// 		log.WithError(err).
	// 			WithFields(log.Fields{
	// 				"uid":       uid,
	// 				"tenant_id": tenant,
	// 			}).
	// 			Error("failed to send heartbeat signal")
	// 	}
	// }

	// tunnel.router = tunnel.Tunnel.Router().(*echo.Echo)

	// `/sessions/:uid/close` is the endpoint that is called by the agent to inform the SSH's server that the session is
	// closed.
	tunnel.router.POST("/api/sessions/:uid/close", func(c echo.Context) error {
		var data struct {
			UID    string `param:"uid"`
			Device string `json:"device"`
		}

		if err := c.Bind(&data); err != nil {
			return err
		}

		ctx := c.Request().Context()

		tenant := c.Request().Header.Get("X-Tenant-ID")

		conn, err := tunnel.Dial(ctx, fmt.Sprintf("%s:%s", tenant, data.Device))
		if err != nil {
			log.WithError(err).Error("could not found the connection to this device")

			return err
		}

		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/ssh/close/%s", data.UID), nil)
		if err != nil {
			log.WithError(err).Error("failed to create a the request for the device")

			return err
		}

		if err := req.Write(conn); err != nil {
			log.WithError(err).Error("failed to perform the HTTP request to the device to close the session")

			return err
		}

		return c.NoContent(http.StatusOK)
	})

	if config.Tunnels {
		// The `/http/proxy` endpoint is invoked by the NGINX gateway when a tunnel URL is accessed. It processes the
		// `X-Address` and `X-Path` headers, which specify the tunnel's address and the target path on the server, returning
		// an error related to the connection to device or what was returned from the server inside the tunnel.
		tunnel.router.Any("/http/proxy", func(c echo.Context) error {
			requestID := c.Request().Header.Get("X-Request-ID")

			address := c.Request().Header.Get("X-Address")
			log.WithFields(log.Fields{
				"request-id": requestID,
				"address":    address,
			}).Debug("address value")

			path := c.Request().Header.Get("X-Path")
			log.WithFields(log.Fields{
				"request-id": requestID,
				"address":    address,
			}).Debug("path")

			endpoint, err := tunnel.API.LookupWebEndpoints(address)
			if err != nil {
				log.WithError(err).Error("failed to get the web endpoint")

				return c.JSON(http.StatusForbidden, NewMessageFromError(ErrWebEndpointForbidden))
			}

			logger := log.WithFields(log.Fields{
				"request-id": requestID,
				"namespace":  endpoint.Namespace,
				"device":     endpoint.Device,
			})

			in, err := tunnel.Dial(c.Request().Context(), fmt.Sprintf("%s:%s", "default", "123456"))
			if err != nil {
				logger.WithError(err).Error("failed to dial to device")

				return c.JSON(http.StatusForbidden, NewMessageFromError(ErrDeviceTunnelDial))
			}

			defer in.Close()

			logger.Trace("new web endpoint connection initialized")
			defer logger.Trace("web endpoint connection doned")

			// NOTE: Connects to the HTTP proxy before doing the actual request. In this case, we are connecting to all
			// hosts on the agent because we aren't specifying any host, on the port specified. The proxy route accepts
			// connections for any port, but this route should only connect to the HTTP server.
			req, _ := http.NewRequest(http.MethodConnect, fmt.Sprintf("/http/proxy/%s:%d", endpoint.Host, endpoint.Port), nil)

			if err := req.Write(in); err != nil {
				logger.WithError(err).Error("failed to write the request to the agent")

				return c.JSON(http.StatusInternalServerError, NewMessageFromError(ErrDeviceTunnelWriteRequest))
			}

			if resp, err := http.ReadResponse(bufio.NewReader(in), req); err != nil || resp.StatusCode != http.StatusOK {
				logger.WithError(err).Error("failed to connect to HTTP port on device")

				return c.JSON(http.StatusInternalServerError, NewMessageFromError(ErrDeviceTunnelConnect))
			}

			req = c.Request()
			req.Host = strings.Join([]string{address, config.TunnelsDomain}, ".")
			req.URL, err = url.Parse(path)
			if err != nil {
				logger.WithError(err).Error("failed to parse the path")

				return c.JSON(http.StatusInternalServerError, NewMessageFromError(ErrDeviceTunnelReadResponse))
			}

			if err := req.Write(in); err != nil {
				logger.WithError(err).Error("failed to write the request to the agent")

				return c.JSON(http.StatusInternalServerError, NewMessageFromError(ErrDeviceTunnelWriteRequest))
			}

			ctr := http.NewResponseController(c.Response())
			out, _, err := ctr.Hijack()
			if err != nil {
				logger.WithError(err).Error("failed to hijact the http request")

				return c.JSON(http.StatusInternalServerError, NewMessageFromError(ErrDeviceTunnelHijackRequest))
			}

			defer out.Close()

			// Bidirectional copy between the client and the device.
			var wg sync.WaitGroup
			wg.Add(2)

			done := sync.OnceFunc(func() {
				defer in.Close()
				defer out.Close()

				logger.Trace("close called on in and out connections")
			})

			go func() {
				defer done()
				defer wg.Done()

				if _, err := io.Copy(in, out); err != nil {
					logger.WithError(err).Debug("in and out done returned a error")
				}

				logger.Trace("in and out done")
			}()

			go func() {
				defer done()
				defer wg.Done()

				if _, err := io.Copy(out, in); err != nil {
					logger.WithError(err).Debug("out and in done returned a error")
				}

				logger.Trace("out and in done")
			}()

			wg.Wait()

			logger.Debug("http proxy is done")

			return nil
		})
	}

	tunnel.router.GET("/healthcheck", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	return tunnel, nil
}

func (t *Tunnel) GetRouter() *echo.Echo {
	return t.router
}

// StreamConn wraps a libp2p stream to implement net.Conn minimally.
type StreamConn struct {
	stream network.Stream
}

func NewStreamConn(s network.Stream) net.Conn { return &StreamConn{stream: s} }

func (c *StreamConn) Read(b []byte) (int, error) {
	log.Printf("StreamConn Read: requesting %d bytes\n", len(b))

	return c.stream.Read(b)
}

func (c *StreamConn) Write(b []byte) (int, error) {
	log.Printf("StreamConn Write: sending %d bytes\n", len(b))

	return c.stream.Write(b)
}
func (c *StreamConn) Close() error { return c.stream.Close() }

func (c *StreamConn) LocalAddr() net.Addr  { return dummyAddr("libp2p-local") }
func (c *StreamConn) RemoteAddr() net.Addr { return dummyAddr("libp2p-remote") }

func (c *StreamConn) SetDeadline(t time.Time) error      { return nil }
func (c *StreamConn) SetReadDeadline(t time.Time) error  { return nil }
func (c *StreamConn) SetWriteDeadline(t time.Time) error { return nil }

// dummyAddr is a tiny net.Addr implementation for compatibility.
type dummyAddr string

func (d dummyAddr) Network() string { return string(d) }
func (d dummyAddr) String() string  { return string(d) }

// Dial trys to get a connetion to a device specifying a key, what is a combination of tenant and device's UID.
func (t *Tunnel) Dial(ctx context.Context, key string) (net.Conn, error) {
	// // return t.Tunnel.Dial(ctx, key)
	// m := *t.m
	// fmt.Println("Current connections map:", m)

	// id := m["default/123456"]
	// fmt.Println("Dialing to ID:", id)
	// fmt.Println("Dialing to ID:", id)
	// fmt.Println("Dialing to ID:", id)
	// fmt.Println("Dialing to ID:", id)
	// fmt.Println("Dialing to ID:", id)
	// fmt.Println("Dialing to ID:", id)
	// fmt.Println("Dialing to ID:", id)

	// s, err := t.h.NewStream(ctx, peer.ID(id), "/ssh/1.0.0")
	// if err != nil {
	// 	log.Println("h.NewStream err:", err)

	// 	return nil, err
	// }

	// return NewStreamConn(s), nil
	return nil, nil
}

// NetworkNotifiee implements the network.Notifiee interface
type NetworkNotifiee struct{}

// Connected implements network.Notifiee.
func (n *NetworkNotifiee) Connected(net network.Network, c network.Conn) {
	fmt.Println("new connection:", net)
}

// Disconnected implements network.Notifiee.
func (n *NetworkNotifiee) Disconnected(net network.Network, c network.Conn) {
	fmt.Println("disconnected:", net)
}

// Listen implements network.Notifiee.
func (n *NetworkNotifiee) Listen(net network.Network, ma multiaddr.Multiaddr) {
	fmt.Println("listening:", ma.String())
}

// ListenClose implements network.Notifiee.
func (n *NetworkNotifiee) ListenClose(net network.Network, ma multiaddr.Multiaddr) {
	fmt.Println("close listening:", ma.String())
}
