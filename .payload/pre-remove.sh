#!/bin/vbash

# Set up the Vyatta environment (see AGENTS.md — transactional CLI only, never raw config files)
declare -i DEC
source /opt/vyatta/etc/functions/script-template
API=/bin/cli-shell-api
CFGRUN=/opt/vyatta/sbin/vyatta-cfg-cmd-wrapper
DATE=$(date +'%FT%H%M%S')

shopt -s expand_aliases

alias begin='${CFGRUN} begin'
alias cleanup='${CFGRUN} cleanup'
alias commit='${CFGRUN} commit'
alias delete='${CFGRUN} delete'
alias end='${CFGRUN} end'
alias save='sudo ${CFGRUN} save'
alias set='${CFGRUN} set'
alias show='_vyatta_op_run show'

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
	CTR=$(printf "%03x" ${DEC})
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
	echo "pre-remove: ${MSG}" | fold -sw ${COLUMNS}
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

isblocklist() {
	${API} existsActive service dns forwarding blocklist && return 0
	return 1
}

# Back up [service dns forwarding blocklist]
backup_dns_config() {
	if isblocklist; then
		file="/config/user-data/edgeos-adblock.${DATE}.cmds"
		echo_logger I "Backing up blocklist configuration to: ${file}"
		echo "edit service dns forwarding" > ${file} 
		${API} showConfig service dns forwarding blocklist \
		--show-commands --show-active-only | \
		grep blocklist >> ${file} || \
		echo_logger E 'Blocklist configuration backup failed!'
	fi
}

# Delete the [service dns forwarding blocklist] configuration
delete_dns_config() {
	try begin
	try delete system task-scheduler task update_edgeos_adblock
	try delete service dns forwarding blocklist
	try commit || {
		echo_logger FE "Configuration commit failed — aborting pre-remove"
		try end 2>/dev/null || true
		exit 1
	}
	try save || {
		echo_logger FE "Configuration save failed — aborting pre-remove"
		try end 2>/dev/null || true
		exit 1
	}
	try end
}

# Remove dnsmasq configuration files
delete_dnsmasq_config() {
	rm -f /etc/dnsmasq.d/*edgeos-adblock.conf
}

restart_dnsmasq() {
	if [[ -f /bin/systemctl ]]; then 
		/bin/systemctl restart dnsmasq 
	else
		/etc/init.d/dnsmasq restart
	fi
}

# echo "$@"

# Back up the existing blocklist configuration
backup_dns_config

# Only run the pre-installation script if this is a first time installation
if [[ "${1}" == "remove" ]] ; then
	echo "Deleting edgeos-adblock configuration settings..."
	delete_dns_config
	delete_dnsmasq_config
fi

set_vyattacfg_grp
restart_dnsmasq
