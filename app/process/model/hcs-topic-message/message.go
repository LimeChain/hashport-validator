package hcstopicmessage

type HCSMessage struct {
	ConsensusTimestamp string `json:"consensus_timestamp"`
	TopicId            string `json:"topic_id"`
	Message            string `json:"message"`
	RunningHash        string `json:"running_hash"`
	SequenceNumber     int64  `json:"sequence_number"`
}

type HCSMessages struct {
	Messages []HCSMessage
}
