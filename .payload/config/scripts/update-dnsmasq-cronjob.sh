#!/bin/bash
# See repo AGENTS.md: this wrapper does not modify Vyatta configuration (sleep + exec only).
# Cron script runs update-dnsmasq at random times within a 3 hour window
# Since cron will run this script, we have to escape the modulus operator, 
# otherwise cron will interpret it as a newline.

random=$(/usr/bin/awk 'BEGIN{srand();printf("%d", 65536*rand())}')
seconds=${1}

[[ ${seconds} -lt 1 ]] && seconds=1
[[ ${seconds} -gt 86400 ]] && seconds=86000 

sleep $(( random % seconds ))
/config/scripts/update-dnsmasq