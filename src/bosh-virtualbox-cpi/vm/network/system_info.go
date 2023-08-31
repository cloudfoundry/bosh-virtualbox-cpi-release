package network

import (
	"fmt"
	"math/big"
	"net"
	"runtime"
	"strings"
)

type SystemInfo struct {
	osVersion        string
	vBoxMajorVersion string
	VBoxMinorVersion string
}

func (n Networks) NewSystemInfo() (SystemInfo, error) {
	vBoxMajorVersion, VBoxMinorVersion, err := n.getVboxVersion()
	if err != nil {
		return SystemInfo{}, err
	}

	return SystemInfo{getOSVersion(), vBoxMajorVersion, VBoxMinorVersion}, nil
}

// GetFirstIP Get the first usable IP address of a subnet
func (s SystemInfo) GetFirstIP(subnet *net.IPNet) (net.IP, error) {
	return getIndexedIP(subnet, 0)
}

// GetLastIP Get the last usable IP address of a subnet
func (s SystemInfo) GetLastIP(subnet *net.IPNet) (net.IP, error) {
	size := rangeSize(subnet)
	if size <= 0 {
		return nil, fmt.Errorf("can't get range size of subnet. subnet: %q", subnet)
	}
	return getIndexedIP(subnet, int(size-1))
}

// IsMacOSXVBoxSpecial6or7Case Identify if you are system is running on MAC OS X and the used
// VirtualBox version is 6.1 or 7
func (s SystemInfo) IsMacOSXVBoxSpecial6or7Case() bool {
	if s.osVersion == "darwin" && (s.vBoxMajorVersion == "7") {
		return true
	} else {
		return false
	}
}

// getVboxVersion Extract the corresponding used Virtual Box version
func (n Networks) getVboxVersion() (string, string, error) {
	output, err := n.driver.Execute("--version")
	if err != nil {
		return "", "", err
	}

	output = strings.TrimSpace(output)
	matches := strings.Split(output, ".")

	if len(matches) > 3 {
		panic(fmt.Sprintf("Internal inconsistency: Expected len(%s matches) >= 3:", createdHostOnlyMatch))
	}

	return matches[0], matches[1], nil
}

// getOSVersion Extract the corresponding used operational system
func getOSVersion() string {
	return runtime.GOOS
}

// rangeSize Identify the range size of valid subnet addresses.
// The functionality is copied from https://github.com/tkestack/tke/blob/v1.9.2/pkg/util/ipallocator/allocator.go
func rangeSize(subnet *net.IPNet) int64 {
	ones, bits := subnet.Mask.Size()
	if bits == 32 && (bits-ones) >= 31 || bits == 128 && (bits-ones) >= 127 {
		return 0
	}

	if bits == 128 && (bits-ones) >= 16 {
		return int64(1) << uint(16)
	}
	return int64(1) << uint(bits-ones)
}

// addIPOffset adds the provided integer offset to a base big.Int representing a net.IP.
// The functionality is copied from https://github.com/tkestack/tke/blob/v1.9.2/pkg/util/ipallocator/allocator.go
func addIPOffset(base *big.Int, offset int) net.IP {
	return big.NewInt(0).Add(base, big.NewInt(int64(offset))).Bytes()
}

// bigForIP Creates a big.Int based on the provided net.IP.
// The functionality is copied from https://github.com/tkestack/tke/blob/v1.9.2/pkg/util/ipallocator/allocator.go
func bigForIP(ip net.IP) *big.Int {
	b := ip.To4()
	if b == nil {
		b = ip.To16()
	}
	return big.NewInt(0).SetBytes(b)
}

// getIndexedIP Get a net.IP that is subnet.IP + index in the contiguous IP space.
// The functionality is copied from https://github.com/tkestack/tke/blob/v1.9.2/pkg/util/ipallocator/allocator.go
func getIndexedIP(subnet *net.IPNet, index int) (net.IP, error) {
	ip := addIPOffset(bigForIP(subnet.IP), index)
	if !subnet.Contains(ip) {
		return nil,
			fmt.Errorf("can't generate IP with index %d from subnet. subnet too small. subnet: %q", index, subnet)
	}
	return ip, nil
}
