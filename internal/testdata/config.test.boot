adblock {
    disabled false
    dns-redirect-ip 192.168.168.1
    allow allow.example.invalid
    block blocked.example.invalid
    source hageziPro {
        description "HaGeZi DNS Blocklists — Pro (dnsmasq)"
        prefix ""
        url https://cdn.jsdelivr.net/gh/hagezi/dns-blocklists@latest/dnsmasq/pro.txt
    }
    source tasty {
        description "Synthetic domain file"
        dns-redirect-ip 10.10.10.10
        file ../../internal/testdata/adblock.source.txt
    }
}
