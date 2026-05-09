#!/usr/bin/env bash
# Verifies Makefile: .DEFAULT_GOAL := help and ordered ##@ headers for this repo.
# Compatible with bash 3.2 (macOS).
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MF="$ROOT/Makefile"

grep -q '^\.DEFAULT_GOAL := help' "$MF" || {
  echo "makefile-conventions: missing .DEFAULT_GOAL := help in Makefile" >&2
  exit 1
}

expected=(Setup Build Test Clean)
tmp=$(mktemp)
grep '^##@' "$MF" | sed 's/^##@[[:space:]]*//' >"$tmp"
count=$(wc -l <"$tmp" | tr -d ' ')
if [[ "$count" -ne "${#expected[@]}" ]]; then
  echo "makefile-conventions: expected ${#expected[@]} category headers, got $count" >&2
  cat "$tmp" >&2
  rm -f "$tmp"
  exit 1
fi
i=0
while IFS= read -r line; do
  if [[ "$line" != "${expected[$i]}" ]]; then
    echo "makefile-conventions: category $i: expected '${expected[$i]}', got '$line'" >&2
    rm -f "$tmp"
    exit 1
  fi
  i=$((i + 1))
done <"$tmp"
rm -f "$tmp"

echo "makefile-conventions: OK"
