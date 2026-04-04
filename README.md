# hl-go

A client library written in Go for the [hamlib](https://github.com/hamlib/hamlib) network daemons rigctld, rotctld, and ampctld.

The library uses the extended response protocol to communicate with the hamlib server, using exclusively the long command names.
It can poll several important values periodically. It also supports the new [multicast feature](https://github.com/Hamlib/Hamlib/blob/master/README.multicast)
to receive changed values asynchronously from the hamlib server, without the need for polling.

## License
This software is published under the [MIT License](https://www.tldrlegal.com/l/mit).

Copyright [Florian Thienel](http://thecodingflow.com/)
