package main

import (
	"fmt"
	"net"
)

func main() {
	ifs, _ := net.Interfaces()

	for i, intf := range ifs {

		fmt.Println(i, intf.Name)
		a, _ := intf.Addrs()
		for _, n := range a {
			fmt.Println(n.String())
		}
	}

}
