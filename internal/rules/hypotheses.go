package rules

import "strings"

// Default returns the generic and verified v0.1 hypotheses.
func Default() []Hypothesis {
	return []Hypothesis{
		{Name: "timed-out-waiting", Explain: "The unit appears to have waited for a timeout.", Risk: "safe", Source: "systemd.service(5)", Evidence: []Check{messageCheck("timeout", 8, false)}},
		{Name: "slow-work", Explain: "The unit did work on the critical chain but the journal does not name a more specific cause.", Risk: "safe", Source: "systemd-analyze(1)", Evidence: []Check{entryCheck("unit logged during its delay", 3)}},
		{Name: "blocked-by-dependency", Explain: "The unit was held behind its preceding critical-chain dependency.", Risk: "safe", Source: "systemd.unit(5)", Evidence: []Check{previousCheck()}},
		{Name: "failed-and-retried", Explain: "The journal records a failure or restart.", Risk: "safe", Source: "systemctl(1)", Evidence: []Check{messageCheck("failed", 8, true), messageCheck("restart", 6, false)}},
		{Name: "shared-bottleneck", Explain: "Several units were slow together, suggesting shared I/O or resource contention.", Risk: "safe", Source: "systemd-analyze(1)", Evidence: []Check{overlapCheck()}},
		{Name: "device-timeout", Explain: "A mount or swap device did not appear in time.", Option: "Check the UUID with lsblk -f; then add nofail or correct the /etc/fstab entry.", Undo: "Restore the /etc/fstab backup.", Risk: "check-first", Source: "systemd.mount(5); fstab(5)", Units: []string{".device", ".mount", ".swap"}, Evidence: []Check{messageCheck("timed out", 10, true), messageCheck("uuid", 8, true)}},
		{Name: "network-online-wait", Explain: "The network-online wait service delayed the boot until connectivity was ready.", Option: "Check nmcli device status or journalctl -b -u NetworkManager before changing wait-online.", Undo: "Re-enable the wait-online service if dependent services need it.", Risk: "check-first", Source: "NetworkManager-wait-online.service(8); systemd-networkd-wait-online.service(8)", Units: []string{"wait-online"}, Evidence: []Check{messageCheck("network", 8, true), entryCheck("wait-online journal activity", 3)}},
		{Name: "fsck", Explain: "A filesystem check ran during boot; this is usually scheduled maintenance.", Risk: "safe", Source: "systemd-fsck(8)", Units: []string{"fsck"}, Evidence: []Check{entryCheck("filesystem check journal activity", 8)}},
		{Name: "failing-unit", Explain: "The unit failed during boot.", Risk: "safe", Source: "systemctl(1)", Evidence: []Check{messageCheck("failed", 10, true)}},
		{Name: "udev-settle", Explain: "udev device enumeration delayed a dependent service.", Risk: "check-first", Source: "systemd-udev-settle.service(8)", Units: []string{"udev-settle"}, Evidence: []Check{entryCheck("udev settle journal activity", 8)}},
	}
}

func messageCheck(needle string, weight int, stated bool) Check {
	return Check{Name: needle, Weight: weight, Stated: stated, Test: func(evidence *Evidence) (bool, string) {
		for _, entry := range evidence.Entries {
			if strings.Contains(strings.ToLower(entry.Msg), needle) || strings.Contains(strings.ToLower(entry.MsgID), needle) {
				return true, "journal: " + entry.Msg
			}
		}
		return false, ""
	}}
}

func entryCheck(name string, weight int) Check {
	return Check{Name: name, Weight: weight, Test: func(evidence *Evidence) (bool, string) {
		return len(evidence.Entries) > 0, "journal entries for " + evidence.Unit
	}}
}

func previousCheck() Check {
	return Check{Name: "critical-chain dependency", Weight: 3, Test: func(evidence *Evidence) (bool, string) {
		return evidence.PreviousUnit != "", "critical chain: after " + evidence.PreviousUnit
	}}
}

func overlapCheck() Check {
	return Check{Name: "overlapping slow units", Weight: 4, Test: func(evidence *Evidence) (bool, string) {
		return len(evidence.Overlapping) >= 4, "several units overlap on the critical chain"
	}}
}
