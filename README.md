# edgeos-adblock

`update-dnsmasq` is a CLI for Ubiquiti **EdgeOS / EdgeRouter** routers. It reads **`service dns forwarding blocklist`** configuration (Vyatta-style), downloads configured URL and file sources, normalizes host and domain lists, writes **dnsmasq** fragments (default **`/etc/dnsmasq.d`** on device, overridable with **`-dir`**), and reloads **dnsmasq** where the stock init layout exists.

The repo includes **Vyatta templates** under **`.payload/`** for the blocklist configuration tree, plus install/remove helpers used with the Debian package.

## What it does

- Loads blocklist configuration from EdgeOS **`config.boot`**-style data (or **`-f`**).
- Processes **domains** and **hosts** trees with EdgeOS semantics: per-source and global **exclude** / **include**, DNS redirect IP, disabled sources, and enable/disable.
- Fetches remote lists over HTTPS and strips per-source prefixes when configured.
- Writes dnsmasq include files and reloads dnsmasq after updates on a live router.
- Optional **`-safe`** loads **`/config/user-data/edgeos-adblock.failover.cfg`** when the primary config is unavailable.

## How blocklists are supplied

Subscriptions are normal HTTP(S) URLs (and optional local files) declared under **`hosts`** and **`domains`** in the EdgeOS blocklist tree. This project does not ship a curated domain blocklist in-tree; you point sources at community lists and tune with EdgeOS excludes/includes.

Shipped examples and the **`Live`** test fixture use **[HaGeZi DNS Blocklists](https://github.com/hagezi/dns-blocklists)** in **dnsmasq** format—the **[Pro list](https://cdn.jsdelivr.net/gh/hagezi/dns-blocklists@latest/dnsmasq/pro.txt)** delivered via **[jsDelivr](https://cdn.jsdelivr.net)**. See **[HaGeZi’s repository](https://github.com/hagezi/dns-blocklists)** for formats, tiers, and changelog. In those snippets the hosts URL source is named **`hageziPro`** (Vyatta source tag).

## CLI

Run **`update-dnsmasq -h`** for current flags. Common options: **`-f`** (config), **`-dir`** (dnsmasq output directory), **`-v`** (verbose), **`-safe`** (failover config).

## Build and install

- **Develop and build from source:** **[docs/build-and-test.md](docs/build-and-test.md)** (`make help`, **`make test`**, **`make build`**, **`make pkgs`** in Docker; **`docker build`** for the slim runtime image).
- **Package layout:** the CLI lives under **`cmd/update-dnsmasq/`**; shared libraries under **`internal/`**. Debian packaging helpers include **`make_deb`** and **`Dockerfile`** as documented there.
- **Router install:** use your usual packaging or artifact workflow; **`.payload/post-install.sh`** illustrates CLI commands that provision blocklist sources and a periodic **`update-dnsmasq`** task. **`config.gateway.json`** is a UniFi **`config.gateway.json`** example for USG-style provisioning.

## License

BSD — see **`license`**.
