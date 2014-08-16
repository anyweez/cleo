package libproc

type PlayerSnapshotContainer struct {
	Snapshot   []byte // unmarshals into proto.PlayerSnapshot
	Timestamp  uint64
	SummonerId uint32
}
