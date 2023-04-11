package main

import (
	"context"
	"sync"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/peer"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
)

func (c *P2Papp) InitDHT() {
	// Start a DHT, for use in peer discovery. We can't just make a new DHT
	// client because we want each peer to maintain its own local copy of the
	// DHT, so that the bootstrapping node of the DHT can go down without
	// inhibiting future peer discovery.

	var err error

	c.kdht, err = dht.New(c.ctx, c.Host)

	if err != nil {
		c.fmtPrintln("Error creating DHT: ", err)
		panic(err)
	}

	if err = c.kdht.Bootstrap(c.ctx); err != nil {
		c.fmtPrintln("Error bootstrapping DHT: ", err)
		panic(err)
	}
	c.fmtPrintln("[*] DHT Initiated")

}

func (c *P2Papp) discoverPeers(RendezvousString string, ctx context.Context, ctx2 context.Context) []peer.AddrInfo {

	var peersFound []peer.AddrInfo

	var wg sync.WaitGroup
	var end = make(chan bool)
	go func() {

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
		dutil.Advertise(ctx, routingDiscovery, RendezvousString)

		// Look for others who have announced and attempt to connect to them

		c.fmtPrintln("[*] Searching for peers in DHT [", RendezvousString, "]")

		peers, err := routingDiscovery.FindPeers(ctx, RendezvousString)
		if err != nil {
			c.fmtPrintln("Error finding peers: ", err)
			panic(err)

		}

		for peerr := range peers {

			if peerr.ID == c.Host.ID() {
				continue
			}
			peersFound = append(peersFound, peerr)
			c.fmtPrintln("\t\t[*] Peer Found:", peer.Encode(peerr.ID))

		}

		c.fmtPrintln("[*] Finished peer discovery, ", len(peersFound), " peers found")
		end <- true
	}()
	select {
	case <-ctx.Done():
		c.fmtPrintln("[*] discoverPeers done")
		return nil
	case <-ctx2.Done():
		c.fmtPrintln("[*] discoverPeers done")
		return nil
	case <-end:
		return peersFound
	}

}
func (c *P2Papp) CancelRendezvous() {
	c.fmtPrintln("[*] DHT canceling")
	c.cancelRendezvous()
}
func (c *P2Papp) AddRendezvous(rendezvous string) {
	c.fmtPrintln("[*] DHT Adding Rendezvous", rendezvous)
	var ctx context.Context
	var end = make(chan bool)
	c.kdht, _ = dht.New(c.ctx, c.Host)

	ctx, c.cancelRendezvous = context.WithTimeout(context.Background(), 120*time.Second)
	defer c.cancelRendezvous()
	var cancelfunc1, cancelfunc2, cancelfunc3, cancelfunc4, cancelfunc5 context.CancelFunc
	var context1, context2, context3, context4, context5 context.Context

	go func() {

		context1, cancelfunc1 = context.WithTimeout(context.Background(), 60*time.Second)
		FoundPeersDHT := c.discoverPeers(rendezvous, ctx, context1)

		context2, cancelfunc2 = context.WithTimeout(context.Background(), 30*time.Second)
		failed := c.connectToPeers(FoundPeersDHT, rendezvous, c.preferquic, true, ctx, context2)
		connected, _ := c.Get(rendezvous)
		c.fmtPrintln("connected Users=", connected)
		time.Sleep(5 * time.Second)
		if len(connected) > 0 {
			context3, cancelfunc3 = context.WithTimeout(context.Background(), 15*time.Second)
			failed = c.requestConnection(failed, rendezvous, c.preferquic, ctx, context3)
			time.Sleep(5 * time.Second)
			context4, cancelfunc4 = context.WithTimeout(context.Background(), 5*time.Second)
			c.connectRelay(rendezvous, ctx, context4)
			context5, cancelfunc5 = context.WithTimeout(context.Background(), 10*time.Second)
			c.connectthrougRelays(failed, rendezvous, c.preferquic, ctx, context5)

		} else {
			c.fmtPrintln("[*] No connected peers")
			c.Add(rendezvous, "")

		}

		end <- true
		//clear all contexts

	}()
	select {
	case <-ctx.Done():
		//clear all contexts
		if cancelfunc1 != nil {
			cancelfunc1()
		}
		if cancelfunc2 != nil {
			cancelfunc2()
		}
		if cancelfunc3 != nil {
			cancelfunc3()
		}
		if cancelfunc4 != nil {
			cancelfunc4()
		}
		if cancelfunc5 != nil {
			cancelfunc5()
		}

		c.fmtPrintln("[*] ctx done")
		return
	case <-end:

		c.fmtPrintln("[*] DHT Ended")
		return

	}
}

func (c *P2Papp) DhtRoutine(quic bool) {
	go func() {
		for {
			select {

			case <-time.After(60 * time.Second):

				for rendezvous, s := range c.data {

					if s.timer == 0 {
						c.fmtPrintln("[*] DHT Timer done", rendezvous)
						c.AddRendezvous(rendezvous)
						c.SetTimer(rendezvous, c.refresh)
					} else {
						c.SetTimer(rendezvous, s.timer-1)

					}
				}

			}
		}
	}()
}
