#!/bin/vbash

# Set up the Vyatta environment (see AGENTS.md — transactional CLI only, never raw config files)
declare -i DEC
source /opt/vyatta/etc/functions/script-template
API=/bin/cli-shell-api
CFGRUN=/opt/vyatta/sbin/vyatta-cfg-cmd-wrapper
shopt -s expand_aliases

alias begin='${CFGRUN} begin'
alias cleanup='${CFGRUN} cleanup'
alias commit='${CFGRUN} commit'
alias delete='${CFGRUN} delete'
alias end='${CFGRUN} end'
alias save='sudo ${CFGRUN} save'
alias set='${CFGRUN} set'

alias bold='tput bold'
alias normal='tput sgr0'
alias reverse='tput smso'
alias underline='tput smul'

alias black='tput setaf 0'
alias blink='tput blink'
alias blue='tput setaf 4'
alias cyan='tput setaf 6'
alias green='tput setaf 2'
alias lime='tput setaf 190'
alias magenta='tput setaf 5'
alias powder='tput setaf 153'
alias purple='tput setaf 171'
alias red='tput setaf 1'
alias tan='tput setaf 3'
alias white='tput setaf 7'
alias yellow='tput setaf 3'

# Setup the echo_logger function
echo_logger() {
	local MSG
	shopt -s checkwinsize
	COLUMNS=$(tput cols)
	DEC+=1
	CTR=$( printf "%03x" ${DEC} )
	TIME=$(date +%H:%M:%S.%3N)

	case "${1}" in
	E)
		shift
		MSG="$(red)$(bold)ERRO$(normal)[${CTR}]${TIME}: ${@} failed!"
		;;
	F)
		shift
		MSG="$(red)$(bold)FAIL$(normal)[${CTR}]${TIME}: ${@}"
		;;
	FE)
		shift
		MSG="$(red)$(bold)CRIT$(normal)[${CTR}]${TIME}: ${@}"
		;;
	I)
		shift
		MSG="$(green)INFO$(normal)[${CTR}]${TIME}: ${@}"
		;;
	S)
		shift
		MSG="$(green)$(bold)NOTI$(normal)[${CTR}]${TIME}: ${@}"
		;;
	T)
		shift
		MSG="$(tan)$(bold)TRYI$(normal)[${CTR}]${TIME}: ${@}"
		;;
	W)
		shift
		MSG="$(yellow)$(bold)WARN$(normal)[${CTR}]${TIME}: ${@}"
		;;
	*)
		echo "ERROR: usage: echo_logger MSG TYPE(E, F, FE, I, S, T, W) MSG."
		exit 1
		;;
	esac

	# MSG=$(echo "${MSG}" | ansi)
	let COLUMNS=${#MSG}-${#@}+${COLUMNS}
	echo "post-install: ${MSG}" | fold -sw ${COLUMNS}
}

# Set the group so that the admin user will be able to commit configs
set_vyattacfg_grp() {
if [[ 'vyattacfg' != $(id -ng) ]]; then
  exec sg vyattacfg -c "$0 $@"
fi
}

# Function to output command status of success or failure to screen and log
try() {
	if eval "${@}"; then
		echo_logger I "${@}"
		return 0
	else
		echo_logger E "${@}"
		return 1
	fi
}

noblocklist() {
	${API} existsActive service dns forwarding blocklist && return 1
	return 0
}

# Load the [service dns forwarding blocklist] configuration (HaGeZi Pro hosts list only; add excludes/sources in configure).
update_dns_config() {
	try begin
	try set service dns forwarding blocklist dns-redirect-ip 0.0.0.0
	# HaGeZi Pro (dnsmasq/pro.txt); Vyatta hosts source tag hageziPro.
	try set service dns forwarding blocklist hosts source hageziPro description '"HaGeZi DNS Blocklists — Pro (dnsmasq)"'
	try set service dns forwarding blocklist hosts source hageziPro prefix ''
	try set service dns forwarding blocklist hosts source hageziPro url 'https://cdn.jsdelivr.net/gh/hagezi/dns-blocklists@latest/dnsmasq/pro.txt'
	try set system task-scheduler task update_edgeos_adblock executable arguments 10800
	try set system task-scheduler task update_edgeos_adblock executable path /config/scripts/update-dnsmasq-cronjob.sh
	try set system task-scheduler task update_edgeos_adblock interval 1d
	try commit || {
		echo_logger FE "Configuration commit failed — aborting post-install"
		try end 2>/dev/null || true
		exit 1
	}
	try save || {
		echo_logger FE "Configuration save failed — aborting post-install"
		try end 2>/dev/null || true
		exit 1
	}
	try end
}

# echo "$@"
# Set group to vyattacfg
set_vyattacfg_grp

# Set UPGRADE flag
UPGRADE=0
[[ "${1}" == "configure" ]] && [[ -z "${2}" ]] && UPGRADE=1

noblocklist && UPGRADE=1

# Only run the post installation script if this is a first time installation
if [[ ${UPGRADE} == 1 ]] ; then
	echo "Installing edgeos-adblock configuration settings..."
	update_dns_config
fi