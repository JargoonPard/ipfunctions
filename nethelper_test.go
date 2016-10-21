package nethelper

import (
	"math"
	"net"
	"testing"
)

func TestIPConversion(t *testing.T) {
	var result uint32

	result = 172 << 24
	result = result + (16 << 16)
	result = result + (3 << 8)

	_, sub1, _ := net.ParseCIDR("172.16.3.0/24")
	value := convertIPtoInt(sub1.IP)

	if value != result {
		t.Errorf("Expected %v, got %v", result, value)
	}
}

func TestStartAndEndRanges(t *testing.T) {
	_, sub1, _ := net.ParseCIDR("172.16.0.0/24")
	range1 := StartAndEndRanges(*sub1)

	start := convertIPtoInt(sub1.IP)
	end := start + uint32(math.Pow(2, 8)) - 1

	if range1.Start != start {
		t.Errorf("START: Expected %v, got %v", start, range1.Start)
	}

	if range1.End != end {
		t.Errorf("END: Expected %v, got %v", end, range1.End)
	}
}

func TestGapFinderSubAtStart(t *testing.T) {
	_, network, _ := net.ParseCIDR("172.16.0.0/16")

	subs := buildTestSubnets(1)
	gaps := getGapRanges(*network, subs)

	if len(gaps) != 1 {
		t.Errorf("Expected a lenth of 1 but got %v", len(gaps))
	}
}

func TestIntToIP(t *testing.T) {
	input := uint32(127 << 24)

	s := IntToIP(input)
	if s != "127.0.0.0" {
		t.Errorf("Expected 127.0.0.0 but got %v", s)
	}

	input = uint32(255<<24) + uint32(255<<16) + uint32(255<<8) + 255
	s = IntToIP(input)
	if s != "255.255.255.255" {
		t.Errorf("Expected 255.255.255.255 but got %v", s)
	}

	input = 0
	s = IntToIP(input)
	if s != "0.0.0.0" {
		t.Errorf("Expected 0.0.0.0 but got %v", s)
	}
}

func TestInvalidSubnet(t *testing.T) {
	var subnets []Range

	_, network, _ := net.ParseCIDR("10.0.0.0/16")
	_, subnet, _ := net.ParseCIDR("172.16.0.0/22")
	subnets = append(subnets, StartAndEndRanges(*subnet))

	gaps := getGapRanges(*network, subnets)

	if len(gaps) != 0 {
		t.Errorf("Expected no gaps but got %v", len(gaps))
	}
}

func TestGapFinder(t *testing.T) {
	_, network, _ := net.ParseCIDR("172.16.0.0/16")

	subs := buildTestSubnets(5)
	By(startIP).Sort(subs)

	gaps := getGapRanges(*network, subs)

	if len(gaps) != 3 {
		t.Errorf("Expected a lenth of 3 but got %v", len(gaps))
	}

	size := gaps[0].size()
	if size != 496 {
		t.Errorf("Expected a gap of 496 but it is %v", size)
	}

	size = gaps[1].size()
	if size != 3072 {
		t.Errorf("Expected a gap of 3072 but it is %v", size)
	}

	size = gaps[2].size()
	if size != 49152 {
		t.Errorf("Expected a gap of 49152 but it is %v", size)
	}
}

func TestFindSubnet(t *testing.T) {
	_, network, _ := net.ParseCIDR("172.16.0.0/16")
	subs := buildTestSubnets(5)

	result := FindSubnet(*network, subs, 24)

	if result != 2886729984 {
		t.Errorf("Expected a starting IP of 2886729984 but got %v", result)
	}

	result = FindSubnet(*network, subs, 21)

	if result != 2886731776 {
		t.Errorf("Expected a starting IP of 2886731776 but got %v", result)
	}

	result = FindSubnet(*network, subs, 20)

	if result != 2886750208 {
		t.Errorf("Expected a starting IP of 2886750208 but got %v", result)
	}

	result = FindSubnet(*network, subs, 16)

	if result != 0 {
		t.Errorf("Expected a starting IP of 0 but got %v", result)
	}
}

func TestGapFinderSubInMiddle(t *testing.T) {
	_, network, _ := net.ParseCIDR("172.16.0.0/16")

	_, sub1, _ := net.ParseCIDR("172.16.32.0/19")
	subRange := StartAndEndRanges(*sub1)

	subs := make([]Range, 1)
	subs[0] = subRange

	gaps := getGapRanges(*network, subs)

	if len(gaps) != 2 {
		t.Errorf("Expected a lenth of 2 but got %v", len(gaps))
	}
}

func buildTestSubnets(count int) []Range {
	var subnets []Range

	switch count {
	case 5:
		_, sub1, _ := net.ParseCIDR("172.16.3.0/24")
		subnets = append(subnets, StartAndEndRanges(*sub1))
		fallthrough
	case 4:
		_, sub2, _ := net.ParseCIDR("172.16.2.0/24")
		subnets = append(subnets, StartAndEndRanges(*sub2))
		fallthrough
	case 3:
		_, sub3, _ := net.ParseCIDR("172.16.32.0/19")
		subnets = append(subnets, StartAndEndRanges(*sub3))
		fallthrough
	case 2:
		_, sub4, _ := net.ParseCIDR("172.16.16.0/20")
		subnets = append(subnets, StartAndEndRanges(*sub4))
		fallthrough
	case 1:
		_, sub5, _ := net.ParseCIDR("172.16.0.0/28")
		subnets = append(subnets, StartAndEndRanges(*sub5))
	}

	return subnets
}
