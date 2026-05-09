// Command update-dnsmasq reads EdgeOS/Vyatta service dns forwarding blocklist
// configuration, downloads and normalizes configured sources, writes dnsmasq
// fragments (default /etc/dnsmasq.d), and reloads dnsmasq. It must run as root.
//
// See the repository README for router behavior and CLI flags; see
// docs/build-and-test.md for building and packaging.
package main
