## packetloss-cover-up

## Why does this exist?
It shouldn't. In a sane world everyone would have access to fast, reliable internet.
In reality, speeds vary and packets get lost.

The latter leads to slow TCP connections with retransmits and annoying behavior in UDP-based realtime applications with audio or video glitches.

## What does this tool do?
One instance sits behind your UDP-based client, one instance sits in front of your UDP-based server.
In such a setup, this tool duplicates all outgoing UDP packets in the hope that at least one will make it through.
All packets are prepended with frame numbers so that the target can properly discard the duplicate packets.

It only supports this process for upstream packets, i.e. client-sent packets.
In theory, it would also work the other way round, but I haven't had a use case for it yet (my downstream is reliable).

## What are the drawbacks?
- It uses twice the (upstream) bandwidth by design.
- It's unstable because it does not have proper error handling.
- It supports a single client only.
- As a proxy, it will masquerade the real client IP address for the target server (`nextHop`).

## Is there any practical use case?
I use it with [Jamulus](https://github.com/corrados/jamulus), an UDP-based real-time music playing software.
For me, it reduces the number of lost packets by about 66%, leading to vastly improved audio quality.

## Usage
```
# Local machine
$ packetloss-cover-up -nextHop your-remote-ip:20000 -listenAddr 127.0.0.1:20000 -wrapUpstream

# Remote machine with a reliable link to the nextHop (e.g. a Jamulus server):
$ packetloss-cover-up -nextHop maybe-a-jamulus-server:22124 -listenAddr your-remote-ip:20000 -unwrapUpstream

# Now point your client to 127.0.0.1:2000 instead of maybe-a-jamulus-server:22124.
```

## License
This implementation is licensed under [AGPLv3](LICENSE.AGPLv3).

## Author
packetloss-cover-up was implemented by [Christian Hoffmann](https://hoffmann-christian.info) in 2021.
