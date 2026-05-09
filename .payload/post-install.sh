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

# Load the [service dns forwarding blocklist] configuration
update_dns_config() {
	try begin
	# try set service dns forwarding blocklist disabled false
	try set service dns forwarding blocklist dns-redirect-ip 0.0.0.0
	try set service dns forwarding blocklist domains include adk2x.com
	try set service dns forwarding blocklist domains include adsrvr.org
	try set service dns forwarding blocklist domains include adtechus.net
	try set service dns forwarding blocklist domains include advertising.com
	try set service dns forwarding blocklist domains include centade.com
	try set service dns forwarding blocklist domains include doubleclick.net
	try set service dns forwarding blocklist domains include fastplayz.com
	try set service dns forwarding blocklist domains include free-counter.co.uk
	try set service dns forwarding blocklist domains include hilltopads.net
	try set service dns forwarding blocklist domains include intellitxt.com
	try set service dns forwarding blocklist domains include kiosked.com
	try set service dns forwarding blocklist domains include patoghee.in
	try set service dns forwarding blocklist domains include themillionaireinpjs.com
	try set service dns forwarding blocklist domains include traktrafficflow.com
	try set service dns forwarding blocklist domains include wwwpromoter.com
	try set service dns forwarding blocklist domains source NoBitCoin description '"Blocking Web Browser Bitcoin Mining"'
	try set service dns forwarding blocklist domains source NoBitCoin prefix '0.0.0.0'
	try set service dns forwarding blocklist domains source NoBitCoin url 'https://raw.githubusercontent.com/hoshsadiq/adblock-nocoin-list/master/hosts.txt'
	try set service dns forwarding blocklist domains source OISD description '"OISD Domains Small"'
	try set service dns forwarding blocklist domains source OISD url 'https://small.oisd.nl/domainswild2'
	try set service dns forwarding blocklist domains source simple_tracking description '"Basic tracking list by Disconnect"'
	try set service dns forwarding blocklist domains source simple_tracking url 'https://s3.amazonaws.com/lists.disconnect.me/simple_tracking.txt'
	try set service dns forwarding blocklist exclude 1e100.net
	try set service dns forwarding blocklist exclude 2o7.net
	try set service dns forwarding blocklist exclude adjust.com
	try set service dns forwarding blocklist exclude adobedtm.com
	try set service dns forwarding blocklist exclude akamai.net
	try set service dns forwarding blocklist exclude akamaihd.net
	try set service dns forwarding blocklist exclude amazon.com
	try set service dns forwarding blocklist exclude amazonaws.com
	try set service dns forwarding blocklist exclude ampproject.org
	try set service dns forwarding blocklist exclude android.clients.google.com
	try set service dns forwarding blocklist exclude apple.com
	try set service dns forwarding blocklist exclude apresolve.spotify.com
	try set service dns forwarding blocklist exclude ask.com
	try set service dns forwarding blocklist exclude avast.com
	try set service dns forwarding blocklist exclude avira-update.com
	try set service dns forwarding blocklist exclude bannerbank.com
	try set service dns forwarding blocklist exclude bazaarvoice.com
	try set service dns forwarding blocklist exclude bing.com
	try set service dns forwarding blocklist exclude bit.ly
	try set service dns forwarding blocklist exclude bitdefender.com
	try set service dns forwarding blocklist exclude bonsaimirai.us9.list-manage.com
	try set service dns forwarding blocklist exclude c.s-microsoft.com
	try set service dns forwarding blocklist exclude cdn.ravenjs.com
	try set service dns forwarding blocklist exclude cdn.visiblemeasures.com
	try set service dns forwarding blocklist exclude clientconfig.passport.net
	try set service dns forwarding blocklist exclude clients2.google.com
	try set service dns forwarding blocklist exclude clients4.google.com
	try set service dns forwarding blocklist exclude cloudfront.net
	try set service dns forwarding blocklist exclude coremetrics.com
	try set service dns forwarding blocklist exclude dickssportinggoods.com
	try set service dns forwarding blocklist exclude dl.dropboxusercontent.com
	try set service dns forwarding blocklist exclude dropbox.com
	try set service dns forwarding blocklist exclude ebay.com
	try set service dns forwarding blocklist exclude edgesuite.net
	try set service dns forwarding blocklist exclude evernote.com
	try set service dns forwarding blocklist exclude express.co.uk
	try set service dns forwarding blocklist exclude feedly.com
	try set service dns forwarding blocklist exclude freedns.afraid.org
	try set service dns forwarding blocklist exclude github.com
	try set service dns forwarding blocklist exclude githubusercontent.com
	try set service dns forwarding blocklist exclude global.ssl.fastly.net
	try set service dns forwarding blocklist exclude google.com
	try set service dns forwarding blocklist exclude googleads.g.doubleclick.net
	try set service dns forwarding blocklist exclude googleadservices.com
	try set service dns forwarding blocklist exclude googleapis.com
	try set service dns forwarding blocklist exclude googletagmanager.com
	try set service dns forwarding blocklist exclude googleusercontent.com
	try set service dns forwarding blocklist exclude gstatic.com
	try set service dns forwarding blocklist exclude gvt1.com
	try set service dns forwarding blocklist exclude gvt1.net
	try set service dns forwarding blocklist exclude hb.disney.go.com
	try set service dns forwarding blocklist exclude herokuapp.com
	try set service dns forwarding blocklist exclude hp.com
	try set service dns forwarding blocklist exclude hulu.com
	try set service dns forwarding blocklist exclude i.s-microsoft.com
	try set service dns forwarding blocklist exclude images-amazon.com
	try set service dns forwarding blocklist exclude live.com
	try set service dns forwarding blocklist exclude logmein.com
	try set service dns forwarding blocklist exclude m.weeklyad.target.com
	try set service dns forwarding blocklist exclude magnetmail1.net
	try set service dns forwarding blocklist exclude microsoft.com
	try set service dns forwarding blocklist exclude microsoftonline.com
	try set service dns forwarding blocklist exclude msdn.com
	try set service dns forwarding blocklist exclude msecnd.net
	try set service dns forwarding blocklist exclude msftncsi.com
	try set service dns forwarding blocklist exclude mywot.com
	try set service dns forwarding blocklist exclude nsatc.net
	try set service dns forwarding blocklist exclude outlook.office365.com
	try set service dns forwarding blocklist exclude paypal.com
	try set service dns forwarding blocklist exclude pop.h-cdn.co
	try set service dns forwarding blocklist exclude products.office.com
	try set service dns forwarding blocklist exclude quora.com
	try set service dns forwarding blocklist exclude rackcdn.com
	try set service dns forwarding blocklist exclude rarlab.com
	try set service dns forwarding blocklist exclude s.youtube.com
	try set service dns forwarding blocklist exclude schema.org
	try set service dns forwarding blocklist exclude shopify.com
	try set service dns forwarding blocklist exclude skype.com
	try set service dns forwarding blocklist exclude smacargo.com
	try set service dns forwarding blocklist exclude sourceforge.net
	try set service dns forwarding blocklist exclude spclient.wg.spotify.com
	try set service dns forwarding blocklist exclude spotify.com
	try set service dns forwarding blocklist exclude spotify.edgekey.net
	try set service dns forwarding blocklist exclude spotilocal.com
	try set service dns forwarding blocklist exclude ssl-on9.com
	try set service dns forwarding blocklist exclude ssl-on9.net
	try set service dns forwarding blocklist exclude sstatic.net
	try set service dns forwarding blocklist exclude static.chartbeat.com
	try set service dns forwarding blocklist exclude storage.googleapis.com
	try set service dns forwarding blocklist exclude twimg.com
	try set service dns forwarding blocklist exclude video-stats.l.google.com
	try set service dns forwarding blocklist exclude viewpoint.com
	try set service dns forwarding blocklist exclude weeklyad.target.com
	try set service dns forwarding blocklist exclude weeklyad.target.com.edgesuite.net
	try set service dns forwarding blocklist exclude windows.net
	try set service dns forwarding blocklist exclude www.msftncsi.com
	try set service dns forwarding blocklist exclude xboxlive.com
	try set service dns forwarding blocklist exclude yimg.com
	try set service dns forwarding blocklist exclude ytimg.com
	try set service dns forwarding blocklist hosts exclude cfvod.kaltura.com
	try set service dns forwarding blocklist hosts include beap.gemini.yahoo.com
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