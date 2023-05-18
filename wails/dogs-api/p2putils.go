package main

import (
	"context"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/net/swarm"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/client"
	"github.com/multiformats/go-multiaddr"
	ma "github.com/multiformats/go-multiaddr"
)

func (c *P2Papp) listCons() {
	c.fmtPrintln("Conns open:")
	for _, v := range c.Host.Network().Conns() {
		c.fmtPrintln(v)
	}
}

// func to close all streams of a given protocol
func (c *P2Papp) closeStreams(protocol string) {
	for _, v := range c.Host.Network().Conns() {
		//list all streams for the connection
		for _, s := range v.GetStreams() {
			if string(s.Protocol()) == protocol {
				s.Reset()
			}
		}

	}
}

// func to list all streams open
func (c *P2Papp) listStreams() {
	c.fmtPrintln("Streams open:")
	for _, v := range c.Host.Network().Conns() {
		//list all streams for the connection
		for _, s := range v.GetStreams() {
			c.fmtPrintln(s.ID(), " ", s.Protocol())
		}

	}
}

func (c *P2Papp) setPeersTRansport(ctx context.Context, rendezvous string, preferQUIC bool) bool {
	//if anyone is succesfully changed, return true
	ret := false
	//get all peers connected using rendezvous

	Peers, _ := c.Get(rendezvous, true)

	for _, v := range Peers {
		aux := c.setTransport(v, preferQUIC)
		if aux {
			ret = true
		}

	}
	return ret

}

func (c *P2Papp) startStreams(rendezvous string, peerid peer.ID) {

	stream1 := c.streamStart(peerid, c.textproto)
	go c.receiveTexthandler(stream1)
	stream2 := c.streamStart(peerid, c.audioproto)
	go c.receiveAudioHandler(stream2)

}
func (c *P2Papp) closeConns(ID peer.ID) {
	for _, v := range c.Host.Network().ConnsToPeer(ID) {
		v.Close()
	}
}

// func to get a stream with a peer of a given protcol
func (c *P2Papp) getStreamsFromPeerProto(peerID peer.ID, protocol protocol.ID) network.Stream {

	for _, v := range c.Host.Network().Conns() {
		if v.RemotePeer() == peerID {
			for _, s := range v.GetStreams() {

				if s.Protocol() == protocol {
					//check if stream is open
					if s.Stat().Direction == network.DirOutbound {
						return s
					}

				}
			}
		}
	}

	return nil
}
func selectAddrs(peeraddrs []multiaddr.Multiaddr, preferQUIC bool, preferTCP bool) []multiaddr.Multiaddr {
	var addrs []multiaddr.Multiaddr
	if preferQUIC {

		for _, v := range peeraddrs {
			if strings.Contains(v.String(), "quic") {
				addrs = append(addrs, v)
			}
		}

	}

	//if not prefer quic connect to tcp addresses and if there are no quic addreses
	if preferTCP {

		for _, v := range peeraddrs {
			if strings.Contains(v.String(), "tcp") {
				addrs = append(addrs, v)
			}
		}

	}
	return addrs

}
func (c *P2Papp) connecToPeersMDNS(peerChan <-chan peer.AddrInfo, rendezvous string, preferQUIC bool, preferTCP bool) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return
		default:

			for peer := range peerChan {

				if peer.ID == c.Host.ID() {
					continue
				}

				c.fmtPrintln("\nConnecting to:", peer.ID.String(), " ", peer.Addrs)

				addrs := selectAddrs(peer.Addrs, preferQUIC, preferTCP)
				if len(addrs) > 0 {
					peer.Addrs = addrs
				}
				err := c.Host.Connect(c.ctx, peer)
				c.fmtPrintln("Connected to:", peer.ID.String(), " ", peer.Addrs)
				if err != nil {
					c.fmtPrintln("Error connecting to peer:", err)
				}
				c.fmtPrintln("add MDNS")
				c.Add(rendezvous, peer.ID)

			}

		}
	}
}

func (c *P2Papp) clear() {
	for _, v := range c.Host.Network().Conns() {

		for _, s := range v.GetStreams() {
			s.Close()
			s.Reset()

		}
		v.Close()

	}

	c.Clear()
}

// func to start a stream with a peer only if there is no stream open and return the stream in any cases
func (c *P2Papp) streamStart(peerid peer.ID, ProtocolID protocol.ID) network.Stream {

	//stream := c.getStreamsFromPeerProto(peerid, ProtocolID)
	var stream network.Stream
	stream = nil
	if stream == nil {
		var err error
		stream, err = c.Host.NewStream(c.ctx, peerid, ProtocolID)
		if err != nil {
			c.fmtPrintln(err.Error())
			if err.Error() == "transient connection to peer" {
				stream, err = c.Host.NewStream(network.WithUseTransient(c.ctx, "relay"), peerid, ProtocolID)

			}
		}

		if err != nil {
			c.fmtPrintln("stream to ", peerid, "failed", ProtocolID)
			c.fmtPrintln("Stream failed:", err)

			return nil

		}

	}
	return stream

}

// func to connect to input peers using relay server
func (c *P2Papp) connectthrougRelays(peersid []peer.ID, rendezvous string, preferQUIC bool, ctx context.Context, ctx2 context.Context) {
	if len(peersid) == 0 {
		return
	}
	c.fmtPrintln("[*] Connecting to peers through Relays")
	var end = make(chan bool)
	go func() {
		peers, _ := c.Get(rendezvous, true)
		for _, server := range peers {
			serverpeerinfo := c.Host.Network().Peerstore().PeerInfo(server)
			if serverpeerinfo.Addrs == nil {
				continue
			}
			c.fmtPrintln("\t[*] Connecting using relay server:", serverpeerinfo.ID.String(), " ", serverpeerinfo.Addrs)

			for _, v := range peersid {

				//check if peer is already connected
				if c.Host.Network().Connectedness(v) == network.Connected {
					continue
				}
				relayaddr, err := ma.NewMultiaddr("/p2p/" + serverpeerinfo.ID.String() + "/p2p-circuit/p2p/" + v.String())
				if err != nil {
					c.fmtPrintln(err)
					continue

				}
				c.fmtPrintln("\t\t[*] Connecting to:", v.String(), " ", relayaddr)

				// Clear the backoff for the unreachable c.Host
				c.Host.Network().(*swarm.Swarm).Backoff().Clear(v)
				// Open a connection to the previously unreachable c.Host via the relay address
				peerrelayinfo := peer.AddrInfo{ID: v, Addrs: []ma.Multiaddr{relayaddr}}
				if c.connectToPeer(network.WithUseTransient(context.Background(), "relay"), peerrelayinfo, rendezvous, preferQUIC, true) {

					//delete peer from peers
					for i, p := range peersid {
						if p == v {
							peersid = append(peersid[:i], peersid[i+1:]...)
						}
					}
				}
			}

		}
		end <- true
	}()
	select {
	case <-end:
		c.fmtPrintln("[*] Connecting using Relays finished.")
		return
	case <-ctx.Done():
		c.fmtPrintln("[*] Connecting using Relays canceled due to global timer.")
		return
	case <-ctx2.Done():
		c.fmtPrintln("[*] Connecting using Relays canceled due to local timer.")
		return
	}

}

// func to reserve circuit with relay server and return all successful connections
func (c *P2Papp) connectRelay(rendezvous string, ctx context.Context, ctx2 context.Context) {
	aux, _ := c.Get(rendezvous, true)
	if len(aux) == 0 {
		return
	}
	var end = make(chan bool)

	go func() {
		c.fmtPrintln("[*] Reserving circuit with connected Hosts...")

		for _, v := range aux {

			// check if peer is  connected

			_, err := client.Reserve(c.ctx, c.Host, c.Host.Network().Peerstore().PeerInfo(v))
			if err == nil {
				c.fmtPrintln("\t[*] Reserved circuit with:", v.String())

			}

		}
		end <- true
	}()
	select {
	case <-end:
		c.fmtPrintln("[*] Reservation finished.")
		return
	case <-ctx.Done():
		c.fmtPrintln("[*] Reservation canceled due to global timer.")
		return
	case <-ctx2.Done():
		c.fmtPrintln("[*] Reservation canceled due to local timer.")
	}

}
