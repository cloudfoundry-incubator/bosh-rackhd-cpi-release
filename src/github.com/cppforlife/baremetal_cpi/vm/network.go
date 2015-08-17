package vm

import (
	"fmt"
	"net"
)

type Networks map[string]Network

type Network struct {
	Type string

	IP      string
	Netmask string
	Gateway string

	DNS     []string
	Default []string

	CloudProperties map[string]interface{}
}

func (ns Networks) Default() Network {
	var foundNetwork Network

	for _, foundNetwork = range ns {
		for _, networkDefault := range foundNetwork.Default {
			if networkDefault == "gateway" {
				return foundNetwork
			}
		}
	}

	return foundNetwork
}

func (n Network) IsDynamic() bool { return n.Type == "dynamic" }

func (n Network) IPWithSubnetMask() string {
	ones, _ := net.IPMask(net.ParseIP(n.Netmask).To4()).Size()
	return fmt.Sprintf("%s/%d", n.IP, ones)
}
