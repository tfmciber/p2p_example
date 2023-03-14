package main

import (
	"fmt"
	"sync"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/peer"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
)

func (c *P2Papp) initDHT() {
	// Start a DHT, for use in peer discovery. We can't just make a new DHT
	// client because we want each peer to maintain its own local copy of the
	// DHT, so that the bootstrapping node of the DHT can go down without
	// inhibiting future peer discovery.
	var err error
	c.kdht, err = dht.New(c.ctx, c.Host)

	if err != nil {
		panic(err)
	}

	if err = c.kdht.Bootstrap(c.ctx); err != nil {

		panic(err)
	}
	var wg sync.WaitGroup
	for _, peerAddr := range dht.DefaultBootstrapPeers {
		peerinfo, _ := peer.AddrInfoFromP2pAddr(peerAddr)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := c.Host.Connect(c.ctx, *peerinfo); err != nil {

			}
		}()
	}
	wg.Wait()

}

func (c *P2Papp) discoverPeers(RendezvousString string) <-chan peer.AddrInfo {

	routingDiscovery := drouting.NewRoutingDiscovery(c.kdht)

	// Look for others who have announced and attempt to connect to them

	fmt.Println("[*] Searching for peers in DHT [", RendezvousString, "]")

	peers, err := routingDiscovery.FindPeers(c.ctx, RendezvousString)
	if err != nil {
		panic(err)
	}
	// Advertise this node, so that it will be found by others but only once
	dutil.Advertise(c.ctx, routingDiscovery, RendezvousString)

	return peers

}
