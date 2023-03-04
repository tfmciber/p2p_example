package main

import (
	"context"
	"fmt"
	"sync"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
)

func initDHT(ctx context.Context, h host.Host) {
	// Start a DHT, for use in peer discovery. We can't just make a new DHT
	// client because we want each peer to maintain its own local copy of the
	// DHT, so that the bootstrapping node of the DHT can go down without
	// inhibiting future peer discovery.

	var wg sync.WaitGroup
	for _, peerAddr := range dht.DefaultBootstrapPeers {
		peerinfo, _ := peer.AddrInfoFromP2pAddr(peerAddr)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := h.Connect(ctx, *peerinfo); err != nil {

			}
		}()
	}
	wg.Wait()

}

func discoverPeers(ctx context.Context, h host.Host, RendezvousString string) <-chan peer.AddrInfo {

	//get current DHT in host

	routingDiscovery := drouting.NewRoutingDiscovery(kademliaDHT)

	// Look for others who have announced and attempt to connect to them

	fmt.Println("\t [*] Searching for peers in DHT [", RendezvousString, "]")

	peers, err := routingDiscovery.FindPeers(ctx, RendezvousString)
	if err != nil {
		panic(err)
	}
	// Advertise this node, so that it will be found by others but only once
	dutil.Advertise(ctx, routingDiscovery, RendezvousString)

	return peers

}
