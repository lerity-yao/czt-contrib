package snake

const (
	defaultWorkerIDBits         = 10
	defaultSequenceBits         = 12
	defaultEpoch          int64 = 1704067200000
	defaultTimeDifference       = 5
	envPodIP                    = "POD_IP"
)

type Conf struct {
	WorkerIDBits   uint8 `json:",default=10"`            // worker id bits, suggest not exceed 10 bits
	SequenceBits   uint8 `json:",default=12"`            // sequence number bits, suggest not exceed 12 bits
	Epoch          int64 `json:",default=1704067200000"` // epoch timestamp in milliseconds
	TimeDifference int64 `json:",default=5"`             // max time difference in milliseconds
	WorkerID       int64 `json:",optional"`
}

func (c *Conf) Validate() error {

	if c.Epoch <= 0 {
		c.Epoch = defaultEpoch
	}

	if c.WorkerIDBits <= 0 || c.WorkerIDBits > 63 {
		c.WorkerIDBits = defaultWorkerIDBits
	}
	if c.SequenceBits <= 0 || c.SequenceBits > 63 {
		c.SequenceBits = defaultSequenceBits
	}

	if c.TimeDifference <= 0 {
		c.TimeDifference = defaultTimeDifference
	}

	return nil
}
