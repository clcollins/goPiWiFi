package main

import (
	"fmt"
	"time"

	"github.com/mdlayher/wifi"
)

func main() {

	c, err := wifi.New()
	defer c.Close()

	if err != nil {
		panic(err)
	}

	interfaces, err := c.Interfaces()

	var i *wifi.Interface

	for _, x := range interfaces {
		if x.Type == wifi.InterfaceTypeStation {
			// Loop through the interfaces, and assign the station
			// to var x
			// We could hardcode the station by name, or index,
			// or hardwareaddr, but this is more portable, if less efficient
			i = x
			break
		}
	}

	for {
		// c.StationInfo(x) returns a slice of all
		// the staton information about the interface
		info, err := c.StationInfo(i)
		if err != nil {
			fmt.Printf("Station err: %s\n", err)
		}

		for _, x := range info {
			fmt.Printf("Signal: %d\n", x.Signal)
		}

		time.Sleep(time.Second)
	}
}

func recoverNoStation() {
	if r := recover(); r != nil {
		fmt.Println("No station ...", r)
	}
}
