#!/bin/sh
set -eu

if [ "$#" -ne 1 ] || [ -z "$1" ] || [ "${1#*/}" != "$1" ]; then
  echo "usage: $0 <fixture-name>" >&2
  exit 2
fi

output_dir="testdata/$1"
mkdir -p "$output_dir"
export LC_ALL=C SYSTEMD_COLORS=0 SYSTEMD_PAGER=cat

if [ "$(id -u)" -ne 0 ]; then
  echo "warning: not running as root; journal data may be partial" >&2
fi

systemd-analyze time >"$output_dir/time.txt"
systemd-analyze critical-chain >"$output_dir/critical-chain.txt" || true
systemd-analyze blame >"$output_dir/blame.txt" || true
journalctl -b -o json >"$output_dir/journal.jsonl" || true
systemctl get-default >"$output_dir/default-target.txt" || true

vm=false
if [ "$(systemd-detect-virt 2>/dev/null || true)" != "none" ]; then
  vm=true
fi
root_source=$(findmnt -no SOURCE / 2>/dev/null || true)
root_disk=$(lsblk -ndo PKNAME "$root_source" 2>/dev/null || true)
rotational=false
if [ -n "$root_disk" ] && [ "$(cat "/sys/block/$root_disk/queue/rotational" 2>/dev/null || echo 0)" = "1" ]; then
  rotational=true
fi
journal_partial=false
if ! journalctl -b _PID=1 -n 1 -o cat 2>/dev/null | grep -q .; then
  journal_partial=true
fi

{
  printf 'vm=%s\n' "$vm"
  printf 'rotational=%s\n' "$rotational"
  printf 'root=%s\n' "$([ "$(id -u)" -eq 0 ] && echo true || echo false)"
  printf 'journal_partial=%s\n' "$journal_partial"
  printf 'boots=%s\n' "$(journalctl --list-boots --no-pager 2>/dev/null | sed '/^[[:space:]]*$/d' | wc -l | tr -d ' ')"
  uname -r
  systemctl --version | sed -n '1p'
  . /etc/os-release 2>/dev/null || true
  printf 'os=%s %s\n' "${ID:-unknown}" "${VERSION_ID:-unknown}"
} >"$output_dir/meta.txt"
