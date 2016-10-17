package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"net"
	"sort"
	"strconv"
	"strings"
)

type NetworkRange struct {
	Start uint32
	End   uint32
}

func (nr NetworkRange) size() uint32 {
	return nr.End - nr.Start
}

func main() {

	_, network, _ := net.ParseCIDR("172.16.0.0/16")
	_, sub1, _ := net.ParseCIDR("172.16.0.0/28")
	_, sub2, _ := net.ParseCIDR("172.16.2.0/24")

	subnets := []NetworkRange{startAndEndRanges(*sub1), startAndEndRanges(*sub2)}
	var cidrSize uint = 24

	subnet := findSubnet(*network, subnets, cidrSize)

	fmt.Printf("subnet definition should be %v/%v based on the gaps and a cidr of %v\n", intToIP(subnet), cidrSize, cidrSize)
}

func findSubnet(network net.IPNet, subnets []NetworkRange, cidr uint) uint32 {
	By(startIP).Sort(subnets)
	myrange := startAndEndRanges(network)

	gaps := getGapRanges(myrange.Start, myrange.End, subnets)

	return findSubnetStart(gaps, cidr)
}

func findSubnetStart(gaps []NetworkRange, cidrSize uint) uint32 {
	blocksize := uint32(math.Pow(2, float64((32 - cidrSize))))

	for _, g := range gaps {
		alignedStart := g.Start + blocksize - (g.Start % blocksize)
		if (g.End - alignedStart + 1) >= blocksize {
			return alignedStart
		}
	}

	return 0
}

func startAndEndRanges(network net.IPNet) NetworkRange {
	s := convertIPtoInt(network.IP)
	mask := convertIPtoInt(net.IP(network.Mask))
	e := s - mask

	return NetworkRange{Start: s, End: e}
}

func startIP(net1, net2 *NetworkRange) bool {
	return net1.Start < net2.Start
}

func convertIPtoInt(input net.IP) uint32 {
	first := uint32(input[0])
	second := uint32(input[1])
	third := uint32(input[2])
	fourth := uint32(input[3])

	var result uint32

	result = (first * uint32(math.Pow(2, 24))) + (second * uint32(math.Pow(2, 16))) + (third * uint32(math.Pow(2, 8))) + fourth

	return result
}

func intToIP(input uint32) string {
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, input)

	return strings.Join(bytesToString(bs), ".")
}

func bytesToString(bs []byte) []string {

	result := make([]string, 4)
	for i := range bs {
		idx := 3 - i
		result[idx] = strconv.Itoa(int(bs[i]))
	}

	return result
}

type By func(net1, net2 *NetworkRange) bool

func (by By) Sort(networks []NetworkRange) {
	ns := &networkSorter{
		networks: networks,
		by:       by,
	}
	sort.Sort(ns)
}

type networkSorter struct {
	networks []NetworkRange
	by       func(net1, net2 *NetworkRange) bool
}

func (n *networkSorter) Len() int {
	return len(n.networks)
}

func (n *networkSorter) Swap(i, j int) {
	n.networks[i], n.networks[j] = n.networks[j], n.networks[i]
}

func (n *networkSorter) Less(i, j int) bool {
	return n.by(&n.networks[i], &n.networks[j])
}

func getGapRanges(start, end uint32, subnets []NetworkRange) []NetworkRange {
	//gaps := make([]NetworkRange, 0)
	var gaps []NetworkRange

	for _, sub := range subnets {
		if start < sub.Start {
			gap := NetworkRange{Start: start, End: sub.Start - 1}
			gaps = append(gaps, gap)
		}

		start = sub.End + 1
	}

	if start < end {
		gap := NetworkRange{Start: start, End: end}
		gaps = append(gaps, gap)
	}

	return gaps
}
