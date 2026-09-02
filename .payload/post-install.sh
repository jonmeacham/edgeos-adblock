#!/bin/vbash

# EdgeOS configuration must be changed through the transactional Vyatta CLI.
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
	echo "post-install: configuration failed" >&2
	cfg end 2>/dev/null || true
	exit 1
}

# Preserve an existing user configuration during package upgrades.
if "${API}" existsActive service dns forwarding adblock; then
	exit 0
fi

echo "Installing edgeos-adblock configuration settings..."
cfg begin || abort_config
cfg set service dns forwarding adblock dns-redirect-ip 0.0.0.0 || abort_config
cfg set service dns forwarding adblock source hageziPro description '"HaGeZi DNS Blocklists — Pro (dnsmasq)"' || abort_config
cfg set service dns forwarding adblock source hageziPro url 'https://cdn.jsdelivr.net/gh/hagezi/dns-blocklists@latest/dnsmasq/pro.txt' || abort_config
cfg set system task-scheduler task edgeos_adblock executable path /config/scripts/edgeos-adblock || abort_config
cfg set system task-scheduler task edgeos_adblock interval 1d || abort_config
cfg commit || abort_config
sudo "${CFGRUN}" save || abort_config
cfg end || abort_config
