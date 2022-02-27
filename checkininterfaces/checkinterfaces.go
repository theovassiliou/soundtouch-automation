package main

import (
	"fmt"
	"net"
)

func main() {
	ifs, _ := net.Interfaces()

	for i, intf := range ifs {

		fmt.Print(i, " ", intf.Name, " ")
		a, _ := intf.Addrs()
		if len(a) > 0 {
			for _, n := range a {
				fmt.Println(n.String())
			}

		} else {
			fmt.Println()
		}
	}

}
