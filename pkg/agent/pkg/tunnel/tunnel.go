package tunnel

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/client"
	"github.com/multiformats/go-multiaddr"
	"github.com/shellhub-io/shellhub/pkg/models"
)

type Tunnel struct {
	peer    host.Host
	Handler func(network.Stream)
}

const (
	SSHProtocol = "/ssh/1.0.0"
)

type Data struct {
	Namespace string `json:"namespace"`
	UID       string `json:"uid"`
}

func NewTunnel(privFile string, auth *models.DeviceAuthResponse) *Tunnel {
	pemData, err := os.ReadFile(privFile)
	if err != nil {
		fmt.Println("Error reading PEM file:", err)
		panic(err)
	}

	block, _ := pem.Decode(pemData)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		fmt.Println("Failed to decode PEM block or incorrect type")

		panic(err)
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		fmt.Println("Error parsing RSA private key:", err)
		panic(err)
	}

	privKey, err := crypto.UnmarshalRsaPrivateKey(x509.MarshalPKCS1PrivateKey(privateKey))
	if err != nil {
		fmt.Println("Error unmarshaling RSA private key:", err)
		panic(err)
	}

	h, err := libp2p.New(
		libp2p.NoListenAddrs,
		libp2p.Identity(privKey),
		libp2p.EnableRelay(),
	)
	if err != nil {
		panic(err)
	}

	t := &Tunnel{
		peer: h,
	}

	h.SetStreamHandler(SSHProtocol, func(s network.Stream) {
		fmt.Println("New stream from:", s.Conn().RemotePeer())

		t.Handler(s)
	})

	time.Sleep(3 * time.Second)

	fmt.Println("------------------------------")
	fmt.Println("agent id: ", h.ID())
	fmt.Println("------------------------------")

	ma, err := multiaddr.NewMultiaddr(auth.ServerAddress)
	if err != nil {
		log.Println("multiaddr.NewMultiaddr err:", err)

		panic(err)
	}

	relay1info, err := peer.AddrInfoFromP2pAddr(ma)
	if err != nil {
		log.Println("peer.AddrInfoFromP2pAddr err:", err)

		panic(err)
	}

	ctx := context.TODO()

	_, err = client.Reserve(context.Background(), h, *relay1info)
	if err != nil {
		log.Printf("unreachable2 failed to receive a relay reservation from relay1. %v", err)
		panic(err)
	}

	if err := h.Connect(ctx, *relay1info); err != nil {
		log.Println("h.Connect err:", err)

		panic(err)
	}

	s, err := h.NewStream(ctx, relay1info.ID, "/register/1.0.0")
	if err != nil {
		log.Println("h.NewStream err:", err)
		panic(err)
	}

	data := Data{
		Namespace: auth.Namespace,
		UID:       auth.UID,
	}

	d, err := json.Marshal(data)
	if err != nil {
		log.Println("json.Marshal err:", err)
		panic(err)
	}

	s.Write(d)

	// relayaddr, err := multiaddr.NewMultiaddr("/p2p/" + "" + "/p2p-circuit/p2p/" + "")
	// if err != nil {
	// 	return nil
	// }

	// for _, a := range peer.Addrs() {
	// 	fmt.Printf(" - %s/p2p/%s\n", a.String(), peer.ID())
	// }

	return t
}

// Listen to reverse listener.
func (t *Tunnel) Listen() error {
	// ctx := context.TODO()

	// ma, err := multiaddr.NewMultiaddr(addrStr)
	// if err != nil {
	// 	return err
	// }

	// addr, err := peer.AddrInfoFromP2pAddr(ma)
	// if err != nil {
	// 	return err
	// }

	// TODO: Warp connect's error into a custom error.
	// t.peer.Connect(ctx, *addr)

	time.Sleep(10 * time.Hour)

	return nil
}

// Close closes the tunnel.
func (t *Tunnel) Close() error {
	return nil
}
