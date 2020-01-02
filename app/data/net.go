package data

import (
	"bytes"
	"net"
	"strings"
	"net/http"
)

type ipRange struct {
	start net.IP
	end net.IP
}

func inRange(r ipRange, ipAddress net.IP) bool {
	return bytes.Compare(ipAddress, r.start) >= 0 && bytes.Compare(ipAddress, r.end) < 0
}

var privateRanges = []ipRange{
	ipRange{
		start: net.ParseIP("10.0.0.0"),
		end:   net.ParseIP("10.255.255.255"),
	},
	ipRange{
		start: net.ParseIP("100.64.0.0"),
		end:   net.ParseIP("100.127.255.255"),
	},
	ipRange{
		start: net.ParseIP("172.16.0.0"),
		end:   net.ParseIP("172.31.255.255"),
	},
	ipRange{
		start: net.ParseIP("192.0.0.0"),
		end:   net.ParseIP("192.0.0.255"),
	},
	ipRange{
		start: net.ParseIP("192.168.0.0"),
		end:   net.ParseIP("192.168.255.255"),
	},
	ipRange{
		start: net.ParseIP("198.18.0.0"),
		end:   net.ParseIP("198.19.255.255"),
	},
}


// isPrivateSubnet - check to see if this ip is in a private subnet
func isPrivateSubnet(ipAddress net.IP) bool {
	if ipCheck := ipAddress.To4(); ipCheck != nil {
		for _, r := range privateRanges {
			if inRange(r, ipAddress){
				return true
			}
		}
	}
	return false
}

func getIPAdress(r *http.Request) string {
	var ipAddress string
	for _, h := range []string{"X-Forwarded-For", "X-Real-Ip"} {
		for _, ip := range strings.Split(r.Header.Get(h), ",") {
			ip = strings.TrimSpace(ip)
			realIP := net.ParseIP(ip)
			if !realIP.IsGlobalUnicast() || isPrivateSubnet(realIP) {
				continue
			}
			
			return ip
		}
	}
	return ipAddress
}