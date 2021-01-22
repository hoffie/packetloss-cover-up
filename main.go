package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"net"
	"time"
)

var nextHop string
var listenAddr string
var unwrapUpstream bool
var wrapUpstream bool

func init() {
	flag.StringVar(&nextHop, "nextHop", "", "address of the next hop to connect to")
	flag.StringVar(&listenAddr, "listenAddr", "127.0.0.1:20000", "where to listen for incoming packets")
	flag.BoolVar(&wrapUpstream, "wrapUpstream", false, "decides whether to wrap the upstream packets")
	flag.BoolVar(&unwrapUpstream, "unwrapUpstream", false, "decides whether to unwrap the upstream packets")
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	fmt.Println("vim-go")
	flag.Parse()
	a, err := net.ResolveUDPAddr("udp", listenAddr)
	check(err)
	s, err := net.ListenUDP("udp", a)
	check(err)

	a, err = net.ResolveUDPAddr("udp", nextHop)
	check(err)
	nh, err := net.DialUDP("udp", nil, a)
	check(err)

	var lastClient net.Addr
	// downstream / receiving side
	go func() {
		buf := make([]byte, 8192)
		for {
			n, err := nh.Read(buf)
			check(err)
			n, err = s.WriteTo(buf[:n], lastClient)
			check(err)
			//fmt.Printf("got %d bytes in downstream: %v\n", n, buf[:n])
		}
	}()

	// upstream / sending side
	var buf [8192]byte
	var wrapBuf [8192]byte
	readPktIdx := uint16(0)
	dupes := 0
	lastKnownPktIdx := uint16(0)
	writePktIdx := uint16(0)
	writePrefix := 0
	if wrapUpstream {
		writePrefix = 2 // len(uint16)
	}
	readPrefix := 0
	if unwrapUpstream {
		readPrefix = 2
	}

	statForwarded := 0.0
	statDiscarded := 0.0
	statRecovered := 0.0
	statLost := 0.0
	go func() {
		for range time.NewTicker(1 * time.Second).C {
			if statForwarded == 0 {
				continue
			}
			fmt.Printf("total: %.0f, discarded: %.0f (%.2f%%), lost: %.0f (%.2f%%), recovered: %.0f (%.2f%%)\n",
				statForwarded,
				statDiscarded, 100.0*statDiscarded/statForwarded,
				statLost, 100.0*statLost/statForwarded,
				statRecovered, 100.0*statRecovered/statForwarded,
			)
		}
	}()
	for {
		var n int
		n, lastClient, err = s.ReadFrom(buf[0:])
		check(err)
		if unwrapUpstream {
			readPktIdx = binary.LittleEndian.Uint16(buf[0:readPrefix])
			if readPktIdx == lastKnownPktIdx {
				// duplicate packet, discard
				dupes++
				continue
			}
			if dupes != 1 {
				// as we are sending duplicate packets, we expect a single dupe.
				// if we see less or more, something is wrong. count that.
				statRecovered++
			}
			if readPktIdx != lastKnownPktIdx+1 {
				if readPktIdx < lastKnownPktIdx+1 {
					// it doesn't make sense to forward out-of-order packets to the next hop.
					statDiscarded++
					continue
				}
				statLost++
			}
			lastKnownPktIdx = readPktIdx
			dupes = 0
		}
		if wrapUpstream {
			writePktIdx++
			binary.LittleEndian.PutUint16(wrapBuf[0:], writePktIdx)
		}
		copy(wrapBuf[writePrefix:], buf[readPrefix:n])
		statForwarded++
		send := func() {
			nh.Write(wrapBuf[:n-readPrefix+writePrefix])
		}
		send()
		if wrapUpstream {
			go func() {
				// send same again to account for packet loss
				time.Sleep(200 * time.Microsecond)
				send()
			}()
		}
	}
}
