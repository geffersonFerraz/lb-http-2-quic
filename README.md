# A Frankenstein Load Balancer.

Receives on port 9999 and forwards traffic anywhere using QUIC (HTTP/3).

Manually forked from https://github.com/AeroNotix/go-quic-proxy and integrated the load balancer functionality.


## If the console displays any warnings regarding buffer size, execute the following command:

sysctl -w net.core.rmem_max=2500000
sysctl -w net.core.wmem_max=2500000

source( https://github.com/quic-go/quic-go/wiki/UDP-Buffer-Sizes )