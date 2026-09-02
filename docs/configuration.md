# Configuration

`edgeos-adblock` reads one EdgeOS configuration tree:

```text
service dns forwarding adblock
├── disabled <true|false>
├── dns-redirect-ip <IPv4-or-IPv6>
├── allow <fqdn>                         [repeatable]
├── block <fqdn>                         [repeatable]
└── source <name>                        [repeatable tagged node]
    ├── description <text>
    ├── url <URL>                        [exactly one of url/file]
    ├── file <path>                      [exactly one of url/file]
    ├── prefix <text>
    └── dns-redirect-ip <IPv4-or-IPv6>
```

There are no compatibility aliases or separate source kinds. Hosts files,
plain domain lists, and dnsmasq lists all normalize to fully qualified domain
names and produce the same dnsmasq address rules.

## Nodes

| Node | Required | Default | Meaning |
| --- | --- | --- | --- |
| `disabled` | No | `false` | When `true`, removes generated blocking and skips source processing. |
| `dns-redirect-ip` | No | `0.0.0.0` | Destination returned for blocked domains. |
| `allow <fqdn>` | No | — | Excludes the domain and its descendants from every generated block rule. Repeatable. |
| `block <fqdn>` | No | — | Adds an explicit domain without requiring a source. Repeatable. |
| `source <name>` | No | — | Defines a uniquely named local or remote source. Repeatable. |
| `source … description` | No | — | Human-readable source description. |
| `source … url` | Conditional | — | HTTP or HTTPS plain-text source. Mutually exclusive with `file`. |
| `source … file` | Conditional | — | Readable local plain-text source. Mutually exclusive with `url`. |
| `source … prefix` | No | Empty | Requires and removes this prefix before domain extraction from ordinary input lines. |
| `source … dns-redirect-ip` | No | Global value | Overrides `dns-redirect-ip` for this source. |

A configured source must have exactly one of `url` or `file`. Source names must
be unique and may contain letters, numbers, underscores, and hyphens. `allow`
takes precedence over `block` and every source. Duplicate domains are emitted
only once across the generated fragments.

## Example

```text
configure
set service dns forwarding adblock dns-redirect-ip '0.0.0.0'

set service dns forwarding adblock source ads description 'Example hosts source'
set service dns forwarding adblock source ads prefix '0.0.0.0 '
set service dns forwarding adblock source ads url 'https://example.invalid/hosts.txt'

set service dns forwarding adblock source internal file '/config/user-data/internal-adblock.txt'
set service dns forwarding adblock source internal dns-redirect-ip '192.0.2.1'

set service dns forwarding adblock allow 'telemetry.example.com'
set service dns forwarding adblock block 'ads.internal.example'

commit
save
exit
```

## Accepted source content

The updater ignores blank lines and lines beginning with `#`, `//`, or `<`.
It extracts domain names from plain lists and hosts-style lines. When `prefix`
is set, ordinary lines must begin with that exact prefix; the prefix is removed
before extraction. These dnsmasq forms are recognized directly regardless of
the configured prefix:

```text
local=/ads.example.com/
address=/ads.example.com/0.0.0.0
```

All output is lowercase and sorted. Comments and duplicate domains are
discarded.

## Generated files

The updater writes dnsmasq fragments under `/etc/dnsmasq.d`:

```text
source.<source-name>.edgeos-adblock.conf
block.explicit.edgeos-adblock.conf
```

Each line is rendered as:

```text
address=/<fqdn>/<dns-redirect-ip>
```

Stale `*.edgeos-adblock.conf` files are removed. Disabling or deleting the
configuration therefore removes rules produced by earlier runs.

## Scheduler

The Debian package creates this package-managed daily job on first install:

```text
system task-scheduler task edgeos_adblock
├── executable path /config/scripts/edgeos-adblock
└── interval 1d
```

The scheduler is separate from the `service dns forwarding adblock` schema.
