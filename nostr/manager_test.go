package nostr

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"testing"
	"time"

	"github.com/nbd-wtf/go-nostr"
	"github.com/stretchr/testify/require"

	"github.com/pion/logging"
	"github.com/pion/turn/v3"
)

func TestTurnParty1(t *testing.T) {
	username, password, err := turn.GenerateLongTermCredentials("openrelayprojectsecret", time.Second*30)
	require.NoError(t, err)

	turnServerAddr := "staticauth.openrelay.metered.ca:80"

	conn, err := net.Dial("tcp", turnServerAddr)
	require.NoError(t, err)

	cfg := &turn.ClientConfig{
		TURNServerAddr: turnServerAddr,
		Conn:           turn.NewSTUNConn(conn),
		Username:       username,
		Password:       password,
		LoggerFactory:  logging.NewDefaultLoggerFactory(),
	}

	client, err := turn.NewClient(cfg)
	require.NoError(t, err)

	defer client.Close()

	err = client.Listen()
	require.NoError(t, err)

	relayConn, err := client.Allocate()
	require.NoError(t, err)
	defer func() {
		if closeErr := relayConn.Close(); closeErr != nil {
			t.Logf("Failed to close connection: %s", closeErr)
			t.Fatal(closeErr)
		}
	}()

	// The relayConn's local address is actually the transport
	// address assigned on the TURN server.
	t.Logf("relayed-address=%s", relayConn.LocalAddr().String())

}

func TestTurnParty2(t *testing.T) {
	username, password, err := turn.GenerateLongTermCredentials("openrelayprojectsecret", time.Second*30)
	require.NoError(t, err)

	turnServerAddr := "staticauth.openrelay.metered.ca:80"

	conn, err := net.Dial("tcp", turnServerAddr)
	require.NoError(t, err)

	cfg := &turn.ClientConfig{
		TURNServerAddr: turnServerAddr,
		Conn:           turn.NewSTUNConn(conn),
		Username:       username,
		Password:       password,
		LoggerFactory:  logging.NewDefaultLoggerFactory(),
	}

	client, err := turn.NewClient(cfg)
	require.NoError(t, err)

	defer client.Close()

	err = client.Listen()
	require.NoError(t, err)

	allocation, err := client.AllocateTCP()
	require.NoError(t, err)
	defer func() {
		if closeErr := allocation.Close(); closeErr != nil {
			t.Logf("Failed to close connection: %s", closeErr)
			t.Fatal(closeErr)
		}
	}()

	t.Logf("relayed-address=%s", allocation.Addr())

	// Learn the peers relay address via signaling channel
	addrCh := make(chan string, 5)
	setupSignalingChannel(addrCh, true, allocation.Addr().String())
	peerAddrStr := <-addrCh

	t.Logf("Received peer address: %s", peerAddrStr)

	peerAddr, err := net.ResolveTCPAddr("tcp", peerAddrStr)
	if err != nil {
		log.Panicf("Failed to resolve peer address: %s", err)
	}

	conn2, err := allocation.DialTCP("tcp", nil, peerAddr)
	if err != nil {
		log.Panicf("Failed to dial: %s", err)
	}
	buf := make([]byte, 4096)

	if _, err = conn.Write([]byte("hello!")); err != nil {
		log.Panicf("Failed to write: %s", err)
	}

	_, err = conn2.Read(buf)
	if err != nil {
		log.Panicf("Failed to read from relay connection: %s", err)
	}

	if err := conn2.Close(); err != nil {
		log.Panicf("Failed to close: %s", err)
	}
}
func TestTurnParty3(t *testing.T) {
	username, password, err := turn.GenerateLongTermCredentials("openrelayprojectsecret", time.Second*30)
	require.NoError(t, err)

	turnServerAddr := "staticauth.openrelay.metered.ca:80"

	conn, err := net.Dial("tcp", turnServerAddr)
	require.NoError(t, err)

	cfg := &turn.ClientConfig{
		TURNServerAddr: turnServerAddr,
		Conn:           turn.NewSTUNConn(conn),
		Username:       username,
		Password:       password,
		LoggerFactory:  logging.NewDefaultLoggerFactory(),
	}

	client, err := turn.NewClient(cfg)
	require.NoError(t, err)

	defer client.Close()

	err = client.Listen()
	require.NoError(t, err)

	allocation, err := client.AllocateTCP()
	require.NoError(t, err)
	defer func() {
		if closeErr := allocation.Close(); closeErr != nil {
			t.Logf("Failed to close connection: %s", closeErr)
			t.Fatal(closeErr)
		}
	}()

	t.Logf("relayed-address=%s", allocation.Addr())

	// Learn the peers relay address via signaling channel
	addrCh := make(chan string, 5)
	setupSignalingChannel(addrCh, false, allocation.Addr().String())
	peerAddrStr := <-addrCh
	peerAddr, err := net.ResolveTCPAddr("tcp", peerAddrStr)
	if err != nil {
		log.Panicf("Failed to resolve peer address: %s", err)
	}

	t.Logf("Received peer address: %s", peerAddrStr)
	buf := make([]byte, 4096)

	if err := client.CreatePermission(peerAddr); err != nil {
		log.Panicf("Failed to create permission: %s", err)
	}

	conn2, err := allocation.AcceptTCP()
	if err != nil {
		log.Panicf("Failed to accept TCP connection: %s", err)
	}

	t.Logf("Accepted connection from: %s", conn.RemoteAddr())

	_, err = conn2.Read(buf)
	if err != nil {
		log.Panicf("Failed to read from relay conn: %s", err)
	}

	if _, err := conn2.Write([]byte("hello back!")); err != nil {
		log.Panicf("Failed to write: %s", err)
	}

	if err := conn2.Close(); err != nil {
		log.Panicf("Failed to close: %s", err)
	}
}

func setupSignalingChannel(addrCh chan string, signaling bool, relayAddr string) {
	addr := "127.0.0.1:5000"
	if signaling {
		go func() {
			listener, err := net.Listen("tcp", addr)
			if err != nil {
				log.Panicf("Failed to create signaling server: %s", err)
			}
			defer listener.Close() //nolint:errcheck,gosec
			for {
				conn, err := listener.Accept()
				if err != nil {
					log.Panicf("Failed to accept: %s", err)
				}

				go func() {
					var message string
					message, err = bufio.NewReader(conn).ReadString('\n')
					if err != nil {
						log.Panicf("Failed to read from relayAddr: %s", err)
					}
					addrCh <- message[:len(message)-1]
				}()

				if _, err = conn.Write([]byte(fmt.Sprintf("%s\n", relayAddr))); err != nil {
					log.Panicf("Failed to write relayAddr: %s", err)
				}
			}
		}()
	} else {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			log.Panicf("Error dialing: %s", err)
		}
		message, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			log.Panicf("Failed to read relayAddr: %s", err)
		}
		addrCh <- message[:len(message)-1]
		if _, err = conn.Write([]byte(fmt.Sprintf("%s\n", relayAddr))); err != nil {
			log.Panicf("Failed to write relayAddr: %s", err)
		}
	}
}
func TestNostr(t *testing.T) {
	ctxb := context.Background()
	relay, err := nostr.RelayConnect(ctxb, "ws://localhost:7000")
	require.NoError(t, err)

	ctxt, cancel := context.WithTimeout(ctxb, time.Second*3)
	defer cancel()
	var filters []nostr.Filter
	since := nostr.Timestamp(int64(0))
	now := nostr.Now()
	filters = []nostr.Filter{
		{
			Kinds: []int{KindMintedMon},
			Since: &since,
			Until: &now,
		},
	}
	sub, err := relay.Subscribe(ctxt, filters)
	require.NoError(t, err)

forloop:
	for {
		select {
		case event := <-sub.Events:
			t.Log(event)
			// Do something with the event
		case <-sub.ClosedReason:
			t.Log("Closed")
			break forloop
		case <-sub.EndOfStoredEvents:
			t.Log("End of stored events")
			break forloop
		case <-ctxt.Done():
			break forloop
		}
	}

	sk := nostr.GeneratePrivateKey()
	pub, _ := nostr.GetPublicKey(sk)

	ev := nostr.Event{
		PubKey:    pub,
		CreatedAt: nostr.Now(),
		Kind:      KindMintedMon,
		Tags:      nil,
		Content:   "Hello World!",
	}
	ev.Sign(sk)

	err = relay.Publish(ctxb, ev)
	require.NoError(t, err)
}
