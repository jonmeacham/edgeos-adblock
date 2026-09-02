#!/bin/vbash

# Package upgrades keep the active configuration; removal deletes it.
[[ "${1:-}" == remove ]] || exit 0

source /opt/vyatta/etc/functions/script-template
API=/bin/cli-shell-api
CFGRUN=/opt/vyatta/sbin/vyatta-cfg-cmd-wrapper

if [[ "$(id -ng)" != vyattacfg ]]; then
	exec sg vyattacfg -c "$0 $*"
fi

cfg() {
	"${CFGRUN}" "$@"
}

abort_config() {
	echo "pre-remove: configuration failed" >&2
	cfg end 2>/dev/null || true
	exit 1
}

echo "Deleting edgeos-adblock configuration settings..."
cfg begin || abort_config
cfg delete system task-scheduler task edgeos_adblock || true
if "${API}" existsActive service dns forwarding adblock; then
	cfg delete service dns forwarding adblock || abort_config
fi
cfg commit || abort_config
sudo "${CFGRUN}" save || abort_config
cfg end || abort_config

rm -f /etc/dnsmasq.d/*.edgeos-adblock.conf
if [[ -x /bin/systemctl ]]; then
	/bin/systemctl restart dnsmasq
else
	/etc/init.d/dnsmasq restart
fi
