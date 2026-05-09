# edgeos-adblock

`update-dnsmasq` is a small CLI for Ubiquiti **EdgeOS / EdgeRouter** routers: it drives **dnsmasq**-based DNS blacklisting and redirection using the same configuration model as EdgeOS’s `service dns forwarding blacklist` tree (Vyatta-style definitions). This repository includes the **Vyatta config templates** under `.payload/` that install beside the binary on the router.

## Blocklist sources

Shipped examples and the in-repo **`Live`** fixture subscribe to **[HaGeZi DNS Blocklists](https://github.com/hagezi/dns-blocklists)** **Pro** tier in **dnsmasq** format: [`dnsmasq/pro.txt`](https://cdn.jsdelivr.net/gh/hagezi/dns-blocklists@latest/dnsmasq/pro.txt) (via jsDelivr). This is **not** HaGeZi Multi NORMAL (`multi.txt`). The hosts URL source uses the Vyatta tag **`hageziPro`**. See the HaGeZi project for formats, tiers, and changelog. EdgeOS **exclude** / **include** entries remain the supported way to tune behavior locally without vendoring a custom domain list.

## What it does

- **Downloads and refreshes** configured blacklist sources (remote URLs and local files), normalizes hosts/domain lists, and writes **dnsmasq** include fragments (default directory `/etc/dnsmasq.d` on-device, overridable with `-dir`).
- **Applies EdgeOS semantics**: domains vs hosts trees, per-source and global **whitelist / exclude** behavior, DNS redirect IP, disabled sources, and blacklist enable/disable—matching how sources are expressed in `config.boot`-style data.
- **Reloads dnsmasq** after updates when run on a live EdgeOS system (expects the stock init/service layout).
- **Fails over** to a backup config path (`-safe` / `/config/user-data/edgeos-adblock.failover.cfg`) when the primary config is unavailable.

Runtime logging uses the router’s log conventions (file under `/var/log` on Linux, `/tmp` on macOS when developing); verbose / debug flags adjust noise.

## CLI

Run **`update-dnsmasq -h`** (or `--help`) for current flags and defaults. Typical knobs include config file (`-f`), dnsmasq output directory (`-dir`), verbose (`-v`), and safe failover (`-safe`).

## Development

The **`update-dnsmasq`** CLI lives under **`cmd/update-dnsmasq/`** (standard Go project layout). Internal packages remain under **`internal/`**.

Build, test, and optional Docker image: **[docs/build-and-test.md](docs/build-and-test.md)** (`make help`, `make ci`, `make docker-build`).

## License

BSD — see **`license`**.
