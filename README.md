#This repository is archived and will no longer receive updates.

# Using Go and a Raspberry Pi to Troubleshoot Wifi

Build a wifi scanner for fun and - well, fun anyway!

## Trouble in Paradise

This past summer, my wife and I sold everything we owned and moved with our two dogs to Hawaii. It's been everything we thought it
would be: beautiful sun, warm sand, cool surf - you name it.  We've also run into some things we didn't expect: wifi problems.

Now, that's not a Hawaii problem. It's limited in scope to the apartment we are renting. We live for now in a single-room studio
apartment attached to our landlord's house. Part of the rent includes free internet! YAY! However, said internet is provided by the
wifi router in the landlord's apartment. BOO!

In all honesty, it works OK. Ish. Ok, it doesn't work well, and I'm not sure why. The router is literally on the other side of the
wall, but our signal is spotty and we have some trouble staying connected. Back home, the wifi router we owned would cross through
many walls and some floors. Certainly, it covered an area larger than the 600 sq. foot apartment we live in!

What does a good techie do in such a situation? Why, investigate, of course!

Luckily the "everything we own" that we sold before moving here did not include among it's ranks a Raspberry Pi Zero W. So small! So
portable! Of course I took it to Hawaii with me!  My bright idea is to use the Pi and it's built-in wifi adapter, write a little (simple) program in Go that will measure the actual Wifi signal being received from the router and display that output. I'm going to make it super simple, quick and dirty, and worry about making it better later.  I just want to know what's up with the wifi, dang it!

Hunting around on Google for even a minute turns up a relatively useful Go package for working with wifi:
"github.com/mdlayher/wifi". Sounds promising!

## Getting information about the Wifi interfaces

The plan is to query the statistics for the wifi interface and return the signal strength, so I need to find the interfaces on the
device. Luckily the "wifi" package has a method to query them, so creating a file named `main.go`, I can do just that:

<!-- markdownlint-disable MD010 -->
```golang
package main

import (
	"fmt"

	"github.com/mdlayher/wifi"
)

func main() {

	c, err := wifi.New()
	defer c.Close()

	if err != nil {
		panic(err)
	}

	interfaces, err := c.Interfaces()

	for _, x := range interfaces {
		fmt.Printf("%+v\n", x)
	}

}
```
<!-- markdownlint-enable MD010 -->

So, what's going on here? After importing the "github.com/mdlayher/wifi" module, it an be used in the main function create a
new Client (type *Client).  The new client (named `c`) can then get a list of the interfaces on the system with `c.Interfaces()`.
Then it's possible to loop over the slice of Interface pointers, and print information about them.

The adding the "+" to `%+v` prints the names of the fields in the `*Interface` struct, too, which is helpful for identifying what you're seeing without having to refer back to documentation.

Running the code above, you get a list of the wifi interfaces on your machine.  In my case, it returned:

```text
&{Index:0 Name: HardwareAddr:5c:5f:67:f3:0a:a7 PHY:0 Device:3 Type:P2P device Frequency:0}
&{Index:3 Name:wlp2s0 HardwareAddr:5c:5f:67:f3:0a:a7 PHY:0 Device:1 Type:station Frequency:2412}
```

Note the MAC address, `HardwareAddr`, is the same for both lines, meaning this is the same physical hardware. This is confirmed by `PHY: 0`. The [Wifi Go Docs](https://godoc.org/github.com/mdlayher/wifi#Interface) note that `PHY` is the physical device to which the interface belongs.

The first interface has no name, and is of `TYPE:P2P`. The second, named "wpl2s0" is of `TYPE:Station`. The Wifi module Go Docs list the [different types of interfaces](https://godoc.org/github.com/mdlayher/wifi#InterfaceType) and describes what they are. According to the docs, the "P2P" type indicates "an interface is a device within a peer-to-peer client network". I believe, and please correct me in the comments if I'm wrong, that this interface is for [Wifi Direct](https://en.wikipedia.org/wiki/Wi-Fi_Direct), a standard for allowing two Wifi devices to connect without an intermediate access point.

The "Station" type indicates "an interface is part of a managed basic service set (BSS) of client devices with a controlling access point". This is the standard function for a wireless device that most will be used to: as a client connected to an access point. This is the interface that matters for testing the quality of the wifi.

## Getting the Station Information from the interface

Using this information, the loop over the interfaces can be updated to retrieve the information we're looking for:

<!-- markdownlint-disable MD010 -->
```golang
	for _, x := range interfaces {
		if x.Type == wifi.InterfaceTypeStation {
			// c.StationInfo(x) returns a slice of all
			// the staton information about the interface
			info, err := c.StationInfo(x)
			if err != nil {
				fmt.Printf("Station err: %s\n", err)
			}
			for _, x := range info {
				fmt.Printf("%+v\n", x)
			}
		}
  }
```
<!-- markdownlint-enable MD010 -->

First, checking that `x.Type` (the type of the Interface) is `wifi.InterfaceTypeStation` - a Station
interface. As mentioned, that's the only type that matters for this exercise. This is an unfortunate naming collision - the interface "type" is not a "type" in the Golang sense. In fact, what we're working here is a Go `type` named `InterfaceType` to represent the type of interface. Whew, that took me a minute to figure out!

So, assuming the interface is of the _correct_ type, the station information can be retrieved with `c.StationInfo(x)`; using the
client `StationInfo()` method to get the info about the interface, `x`.

This returns a slice of `*StationInfo` pointers.  I'm not sure quite why there's a slice.  Perhaps the interface can have multiple
StationInfo responses?  In any case, we can loop over the slice and use the same `+%v` trick to print the keys and values for the
StationInfo struct.

Running the above returns:

```txt
&{HardwareAddr:70:5a:9e:71:2e:d4 Connected:17m10s Inactive:1.579s ReceivedBytes:2458563 TransmittedBytes:1295562 ReceivedPackets:6355 TransmittedPackets:6135 ReceiveBitrate:2000000 TransmitBitrate:43300000 Signal:-79 TransmitRetries:2306 TransmitFailed:4 BeaconLoss:2}
```

We're interested in the "Signal", and possibly "TransmitFailed" and "BeaconLoss". The signal is reported in units of dBm, or
decibel-milliwatts.

## A quick aside: How to read Wifi dBm

* -30 is the best possible signal strength - not realistic or necessary
* -67 is very good - for apps that need reliable packet delivery like streaming media
* -70 is fair - minimum reliable packet delivery, fine for email and web
* -80 is poor - absolute basic connectivity, unreliable packet delivery
* -90 is unusable - approaching the "noise floor"

<!-- markdownlint-disable MD036 -->
_Note that dBm is logarithmic scale:  -60 is 1000x lower than -30_
<!-- markdownlint-enable MD036 -->

REF: [MetaGeek Wifi Signal Strength Basics - https://www.metageek.com/training/resources/wifi-signal-strength-basics.html](https://www.metageek.com/training/resources/wifi-signal-strength-basics.html)

## Making this a real "scanner"

So, looking at the signal from the result above: -79.  YIKES, not good.  That single result is not especially helpful. That's just a
point-in-time reference, and only valid for the particular physical space the Wifi network adapter was in at that instant.  What
would be more useful would be a continuous reading, so it's possible to see how the signal changes as the Raspberry Pi is moved
around. The main function can be tweaked again to accomplish this:

<!-- markdownlint-disable MD010 -->
```golang
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
```
<!-- markdownlint-enable MD010 -->

First, we name a variable `i` of type `*wifi.Interface`.  Since it's outside the loop, we can use it to store the interface
information. Any variable created inside the loop is inaccessible outside the scope of that loop.

Then, we can break the loop into two.  The first loop ranges over the interfaces returned by `c.Interfaces()`, and if that interface
is a Station type, stores that in the `i` variable we created earlier, and breaks out of the loop.

The second loop is an infinite loop, so it'll just run over and over until we CTRL-C to end the program.  This loop takes that interface information and retrieves the station information, as we did before, and prints out the signal information. Then it sleeps for one second, and runs again, printing the signal information over and over until we quit.

So, running that:

```txt
[chris@marvin wifi-monitor]$ go run main.go
Signal: -81
Signal: -81
Signal: -79
Signal: -81
```

Oof.  Not good.

## Mapping the apartment

However, that information is good to know, at least.  With an attached screen or eInk display and a battery (or a looooong extension cable) I can walk the Pi around the apartment and map out where the dead spots are.

Spoiler alert: with the landlord's access point in the apartment next door, the big dead spot for me is a cone shape eminating from
the refrigerator in the kitchen area of the studio apartment...the refrigerator that shares a wall with the landlord's apartment!

I think in Dungeons and Dragons lingo, this is a "Cone of Silence".  Or at least a "Cone of Poor Internet".

Anyway, this code can be compiled directly on the Raspberry Pi with `go build -o wifi_scanner` and the resulting binary, "wifi_scanner" can be shared with any other ARM devices (of the same version).  Alternatively, it can be compiled on a regular system with the right libraries for ARM devices.

Happy Pi Scanning!  May your wifi router not be behind your refrigerator!

_The code used for this article can be found in my Github repo:_ [https://github.com/clcollins/goPiWiFi](https://github.com/clcollins/goPiWiFi)
