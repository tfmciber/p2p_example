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

func initDHT(ctx context.Context, h host.Host) *dht.IpfsDHT {
	// Start a DHT, for use in peer discovery. We can't just make a new DHT
	// client because we want each peer to maintain its own local copy of the
	// DHT, so that the bootstrapping node of the DHT can go down without
	// inhibiting future peer discovery.
	kademliaDHT, err := dht.New(ctx, h)

	if err != nil {
		panic(err)
	}

	if err = kademliaDHT.Bootstrap(ctx); err != nil {
		fmt.Print("rfwrewrwe")
		panic(err)
	}
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
	return kademliaDHT

}

func discoverPeers(ctx context.Context, kademliaDHT *dht.IpfsDHT, h host.Host, RendezvousString string) <-chan peer.AddrInfo {

	routingDiscovery := drouting.NewRoutingDiscovery(kademliaDHT)

	// Look for others who have announced and attempt to connect to them

	fmt.Println("[*] Searching for peers in DHT [", RendezvousString, "]")

	peers, err := routingDiscovery.FindPeers(ctx, RendezvousString)
	if err != nil {
		panic(err)
	}
	// Advertise this node, so that it will be found by others but only once
	dutil.Advertise(ctx, routingDiscovery, RendezvousString)

	return peers

}
