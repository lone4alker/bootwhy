package analyze

import "time"

const (
	// tune: expected time relative to the median chain peer.
	PeerFactor = 3.0
	// tune: minimum expected unit time.
	MinExpected = 500 * time.Millisecond
	// tune: expectation scaling for virtual machines and spinning disks.
	VMFactor  = 1.5
	HDDFactor = 2.0
	// tune: delay budget cutoffs.
	FloorShare = 0.10
	FloorAbs   = time.Second
	// tune: normal boot thresholds.
	NormalAbs   = 2 * time.Second
	NormalShare = 0.10
	// tune: dominant contributor threshold.
	DominantShare = 0.60
)
