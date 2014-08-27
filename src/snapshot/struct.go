package snapshot

/**
 * This file defines the data structure for storing snapshots.
 */

type PlayerSnapshot struct {
	// Time boundaries
	StartTimestamp		uint64
	EndTimestamp		uint64

	SummonerId		uint32
	GamesList		[]uint64

	// TODO: add rank to this snapshot

	// The relevant gameplay statistics for the period covered by this snapshot
	Stats 			[]PlayerStat

	// When this record was generated.
	CreationTimestamp	uint64	
}

type PlayerStat struct {
	Name		string
	Absolute	float32
	Normalized	float32
}
