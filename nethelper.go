package nethelper

import (
	"encoding/binary"
	"math"
	"net"
	"sort"
	"strconv"
	"strings"
)

//Range is a simple type that represents the start and
//end IP address expressed as uint32 values
type Range struct {
	Start uint32
	End   uint32
}

func (nr Range) size() uint32 {
	return nr.End - nr.Start + 1 //add 1 because both values are inclusive in the range
}

//FindSubnet finds the start address of a subnet that meets the cidr requirments
//from within the network that has the existing subnets already defined
func FindSubnet(network net.IPNet, subnets []Range, cidr int) net.IPNet {
	By(startIP).Sort(subnets)

	gaps := getGapRanges(network, subnets)

	startAddress := findSubnetStart(gaps, cidr)
	//startAddress := FindSubnet(network, subnets, cidr)
	ip := net.ParseIP(IntToIP(startAddress))
	_, subnet, _ := net.ParseCIDR(ip.String() + "/" + strconv.Itoa(cidr))

	return *subnet
}

func findSubnetStart(gaps []Range, cidrSize int) uint32 {
	blocksize := uint32(math.Pow(2, float64((32 - cidrSize))))

	for _, g := range gaps {
		if g.size() < blocksize {
			continue
		}
		alignedStart := g.Start + blocksize - (g.Start % blocksize)
		if (g.End - alignedStart + 1) >= blocksize {
			return alignedStart
		}
	}

	return 0
}

//StartAndEndRanges given an IPNet will return a Range object
//containing the starting and ending IP address values of the
//network
func StartAndEndRanges(network net.IPNet) Range {
	s := convertIPtoInt(network.IP)
	mask := convertIPtoInt(net.IP(network.Mask))
	e := s - mask - 1 //subtract 1 because the endpoints are inclusive in the network

	return Range{Start: s, End: e}
}

func startIP(net1, net2 *Range) bool {
	return net1.Start < net2.Start
}

//IntToIP converts a uint32 representation of a network address into
//a . notation string representation (e.g. 127.0.0.0)
func IntToIP(input uint32) string {
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, input)

	return strings.Join(bytesToString(bs), ".")
}

func convertIPtoInt(input net.IP) uint32 {
	return (uint32(input[0]) << 24) + (uint32(input[1]) << 16) + (uint32(input[2]) << 8) + uint32(input[3])
}

func bytesToString(bs []byte) []string {

	result := make([]string, 4)
	for i := range bs {
		idx := 3 - i
		result[idx] = strconv.Itoa(int(bs[i]))
	}

	return result
}

//By is the definition of the function that is used to
//sort two Range objects
type By func(net1, net2 *Range) bool

//Sort is the implementation of the sort function
//used to sort a slice of Range objects
func (by By) Sort(networks []Range) {
	ns := &networkSorter{
		networks: networks,
		by:       by,
	}
	sort.Sort(ns)
}

type networkSorter struct {
	networks []Range
	by       func(net1, net2 *Range) bool
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

//getGapRanges is used to create a collection of range objects that
//represent the available IP ranges within the network that can be
//used to create new subnets
func getGapRanges(network net.IPNet, subnets []Range) []Range {
	var gaps []Range
	networkRange := StartAndEndRanges(network)
	start := networkRange.Start

	for _, sub := range subnets {
		ip := net.ParseIP(IntToIP(sub.Start))
		if network.Contains(ip) && start < sub.Start {
			gap := Range{Start: start, End: sub.Start - 1}
			gaps = append(gaps, gap)
		}

		start = sub.End + 1
	}

	if start < networkRange.End {
		gap := Range{Start: start, End: networkRange.End}
		gaps = append(gaps, gap)
	}

	return gaps
}
