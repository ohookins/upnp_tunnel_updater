# UPnP-based tunnelbroker.net endpoint updater

[Hurricane Electric](https://www.tunnelbroker.net/) offers free IPv6 over IPv6
tunnels, but requires the endpoint to be aware of your own IPv4 endpoint in
order for the tunnel to operate (which makes perfect sense - it is an IP-based
tunnel which has no awareness of connections and needs to know where to forward
inbound packets).

If you are on a dynamic IP address (especially one that changes as frequently
as mine), this poses a small problem. There are a variety of methods available
to you such as Dynamic DNS updaters, adding shell scripting to a field in your
router if it is Linux-based, etc. Actually there is no shortage of solutions at
all. I chose to implement something using Universal Plug'n'Play for several
reasons:
* My router's ability to run DynDNS updating scripts is fairly limited, and
quite unreliable.
* I cannot (easily) have a script running that updates both my tunnel broker
endpoint and _actual_ DynDNS.
* It seems unnecessary to have to go out to the internet (e.g. one of the many
"what is my IP"-style sites) just to figure out your public address.
* I wanted to learn about UPnP.
* There's always time for more Go programming.

# References
* [UPnP Device Architecture](http://upnp.org/specs/arch/UPnP-arch-DeviceArchitecture-v1.0.pdf)
 * This describes the general structure of the protocol in most detail.
* [WAN IP Connection Service](http://upnp.org/specs/gw/UPnP-gw-WANIPConnection-v1-Service.pdf)
 * The actual service that needs to be queried after discovery.
* [UPnP Hacks](http://www.upnp-hacks.org/upnp.html)
 * Has some practical examples of usage.

# Building
* Download and configure Golang (at least 1.2 recommended)
* ```go fmt; go build```

# Running
```
./upnp_tunnel_updater \
  -user-id=<USERID> \
  -password=<PASSWORD> \
  -tunnel-id=<TUNNEL_ID>
```

# TODO
* Flag presence checks.
* Local cache file to prevent spurious update attempts of remote config.
* Subscribing to the event endpoint and remaining resident for continuous
updates.
* Code cleanup.
