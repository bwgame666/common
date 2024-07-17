package libs

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
)

func isIPv4(ip net.IP) bool {
	return ip != nil && ip.To4() != nil
}

func IsIPv4(str string) bool {
	addr := net.ParseIP(str)
	return isIPv4(addr)
}

func IsIPv6(str string) bool {
	addr := net.ParseIP(str)
	return addr != nil && !isIPv4(addr) && addr.To16() != nil
}

func IsEOFOrNetReadError(err error) bool {
	if err == io.EOF {
		return true
	}
	ne, ok := err.(*net.OpError)
	return ok && ne.Op == "read"
}

func SplitAddress(addr string) (host string, port int, err error) {
	var portStr string

	host, portStr, err = net.SplitHostPort(addr)

	if err != nil {
		return "", 0, err
	}

	port, err = strconv.Atoi(portStr)

	if err != nil {
		return "", 0, err
	}

	return
}

func JoinAddress(host string, port int) string {
	return fmt.Sprintf("%s:%d", host, port)
}

var (
	ErrInvalidPortRange     = errors.New("invalid port range")
	ErrInvalidAddressFormat = errors.New("invalid address format")
)

type Address struct {
	Scheme  string
	Host    string
	MinPort int
	MaxPort int
	Path    string
}

func (a *Address) HostPortString(port int) string {
	return fmt.Sprintf("%s:%d", a.Host, port)
}

// return scheme://host:port/path
func (a *Address) String(port int) string {
	if a.Scheme == "" {
		return a.HostPortString(port)
	}

	return fmt.Sprintf("%s://%s:%d%s", a.Scheme, a.Host, port, a.Path)
}

// scheme://host:minPort~maxPort/path
func ParseAddress(addr string) (addrObj *Address, err error) {
	addrObj = new(Address)

	schemePos := strings.Index(addr, "://")

	if schemePos != -1 {
		addrObj.Scheme = addr[:schemePos]
		addr = addr[schemePos+3:]
	}

	colonPos := strings.Index(addr, ":")

	if colonPos != -1 {
		addrObj.Host = addr[:colonPos]
		addr = addr[colonPos+1:]
	} else {
		return nil, ErrInvalidAddressFormat
	}

	rangePos := strings.Index(addr, "~")

	var minStr, maxStr string
	if rangePos != -1 {
		minStr = addr[:rangePos]

		slashPos := strings.Index(addr, "/")

		if slashPos != -1 {
			maxStr = addr[rangePos+1 : slashPos]
			addrObj.Path = addr[slashPos:]
		} else {
			maxStr = addr[rangePos+1:]
		}
	} else {
		slashPos := strings.Index(addr, "/")

		if slashPos != -1 {
			addrObj.Path = addr[slashPos:]
			minStr = addr[rangePos+1 : slashPos]
		} else {
			minStr = addr[rangePos+1:]
		}
	}

	// extract min port
	addrObj.MinPort, err = strconv.Atoi(minStr)
	if err != nil {
		return nil, ErrInvalidPortRange
	}

	if maxStr != "" {
		// extract max port
		addrObj.MaxPort, err = strconv.Atoi(maxStr)
		if err != nil {
			return nil, ErrInvalidPortRange
		}
	} else {
		addrObj.MaxPort = addrObj.MinPort
	}

	return
}

func DetectPort(addr string, fn func(a *Address, port int) (interface{}, error)) (interface{}, error) {
	addrObj, err := ParseAddress(addr)
	if err != nil {
		return nil, err
	}

	for port := addrObj.MinPort; port <= addrObj.MaxPort; port++ {

		ln, err := fn(addrObj, port)
		if err == nil {
			return ln, nil
		}

		// hit max port
		if port == addrObj.MaxPort {
			return nil, err
		}
	}

	return nil, fmt.Errorf("unable to bind to %s", addr)
}

func GetLocalIP() string {
	list, err := GetPrivateIPv4()
	if err != nil {
		return ""
	}

	if len(list) == 0 {
		return ""
	}

	return list[0].String()
}

func GetPrivateIPv4() ([]*net.IPAddr, error) {
	addresses, err := activeInterfaceAddresses()
	if err != nil {
		return nil, fmt.Errorf("Failed to get interface addresses: %v", err)
	}

	var addrs []*net.IPAddr
	for _, rawAddr := range addresses {
		var ip net.IP
		switch addr := rawAddr.(type) {
		case *net.IPAddr:
			ip = addr.IP
		case *net.IPNet:
			ip = addr.IP
		default:
			continue
		}
		if ip.To4() == nil {
			continue
		}
		if !IsPrivate(ip) {
			continue
		}
		addrs = append(addrs, &net.IPAddr{IP: ip})
	}
	return addrs, nil
}

func GetPublicIPv6() ([]*net.IPAddr, error) {
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return nil, fmt.Errorf("Failed to get interface addresses: %v", err)
	}

	var addrs []*net.IPAddr
	for _, rawAddr := range addresses {
		var ip net.IP
		switch addr := rawAddr.(type) {
		case *net.IPAddr:
			ip = addr.IP
		case *net.IPNet:
			ip = addr.IP
		default:
			continue
		}
		if ip.To4() != nil {
			continue
		}
		if IsPrivate(ip) {
			continue
		}
		addrs = append(addrs, &net.IPAddr{IP: ip})
	}
	return addrs, nil
}

// privateBlocks contains non-forwardable address blocks which are used
// for private networks. RFC 6890 provides an overview of special
// address blocks.
var privateBlocks = []*net.IPNet{
	parseCIDR("10.0.0.0/8"),     // RFC 1918 IPv4 private network address
	parseCIDR("100.64.0.0/10"),  // RFC 6598 IPv4 shared address space
	parseCIDR("127.0.0.0/8"),    // RFC 1122 IPv4 loopback address
	parseCIDR("169.254.0.0/16"), // RFC 3927 IPv4 link local address
	parseCIDR("172.16.0.0/12"),  // RFC 1918 IPv4 private network address
	parseCIDR("192.0.0.0/24"),   // RFC 6890 IPv4 IANA address
	parseCIDR("192.0.2.0/24"),   // RFC 5737 IPv4 documentation address
	parseCIDR("192.168.0.0/16"), // RFC 1918 IPv4 private network address
	parseCIDR("::1/128"),        // RFC 1884 IPv6 loopback address
	parseCIDR("fe80::/10"),      // RFC 4291 IPv6 link local addresses
	parseCIDR("fc00::/7"),       // RFC 4193 IPv6 unique local addresses
	parseCIDR("fec0::/10"),      // RFC 1884 IPv6 site-local addresses
	parseCIDR("2001:db8::/32"),  // RFC 3849 IPv6 documentation address
}

func parseCIDR(s string) *net.IPNet {
	_, block, err := net.ParseCIDR(s)
	if err != nil {
		panic(fmt.Sprintf("Bad CIDR %s: %s", s, err))
	}
	return block
}

func IsPrivate(ip net.IP) bool {
	for _, priv := range privateBlocks {
		if priv.Contains(ip) {
			return true
		}
	}
	return false
}

// Returns addresses from interfaces that is up
func activeInterfaceAddresses() ([]net.Addr, error) {
	var upAddrs []net.Addr
	var loAddrs []net.Addr

	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("Failed to get interfaces: %v", err)
	}

	for _, iface := range interfaces {
		// Require interface to be up
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		addresses, err := iface.Addrs()
		if err != nil {
			return nil, fmt.Errorf("Failed to get interface addresses: %v", err)
		}

		if iface.Flags&net.FlagLoopback != 0 {
			loAddrs = append(loAddrs, addresses...)
			continue
		}

		upAddrs = append(upAddrs, addresses...)
	}

	if len(upAddrs) == 0 {
		return loAddrs, nil
	}

	return upAddrs, nil
}

func ClientIP(r *http.Request) string {
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	ip := strings.TrimSpace(strings.Split(xForwardedFor, ",")[0])
	if ip != "" {
		return ip
	}

	ip = strings.TrimSpace(r.Header.Get("X-Real-Ip"))
	if ip != "" {
		return ip
	}

	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		return ip
	}

	return ""
}
