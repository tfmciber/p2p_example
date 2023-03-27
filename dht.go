package main

import (
	"fmt"
	"sync"
	"time"

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
		fmt.Println("Error creating DHT: ", err)
		panic(err)
	}

	if err = c.kdht.Bootstrap(c.ctx); err != nil {
		fmt.Println("Error bootstrapping DHT: ", err)
		panic(err)
	}

}

func (c *P2Papp) discoverPeers(RendezvousString string) []peer.AddrInfo {

	var peersFound []peer.AddrInfo

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

	routingDiscovery := drouting.NewRoutingDiscovery(c.kdht)

	// Advertise this node, so that it will be found by others but only once
	dutil.Advertise(c.ctx, routingDiscovery, RendezvousString)

	// Look for others who have announced and attempt to connect to them

	fmt.Println("[*] Searching for peers in DHT [", RendezvousString, "]")

	peers, err := routingDiscovery.FindPeers(c.ctx, RendezvousString)
	if err != nil {
		fmt.Println("Error finding peers: ", err)
		panic(err)

	}

	for peerr := range peers {

		if peerr.ID == c.Host.ID() {
			continue
		}
		peersFound = append(peersFound, peerr)
		fmt.Println("\t\t[*] Peer Found:", peer.Encode(peerr.ID))

	}

	c.kdht.Close()

	fmt.Println("[*] Finished peer discovery, ", len(peersFound), " peers found")
	return peersFound

}

func (c *P2Papp) dhtRoutine(quic bool, refresh uint) {

	for {
		select {

		case <-time.After(60 * time.Second):

			for rendezvous, s := range c.data {

				if s.timer == 0 {
					c.rendezvousS <- rendezvous
				} else {
					c.SetTimer(rendezvous, s.timer-1)

				}
			}

		case aux := <-c.rendezvousS:

			FoundPeersDHT := c.discoverPeers(aux)

			failed := c.connectToPeers(FoundPeersDHT, aux, quic, true)

			failed = c.requestConnection(failed, aux, quic)
			time.Sleep(5 * time.Second)
			c.connectRelay(aux)
			if len(failed) > 0 {
				c.connectthrougRelays(failed, aux, quic)

			}
			c.SetTimer(aux, refresh)

		}
	}
}
