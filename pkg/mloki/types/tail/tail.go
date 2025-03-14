package tail

// path: $
type Root struct {
	Streams []Stream  `json:"streams"`
	Dropped []Dropped `json:"dropped_entries"`
}

// path: $.streams
type Stream struct {
	Stream map[string]string `json:"stream" note:"这是数据包含的 label"`
	Values [][]string        `json:"values" note:"数组成员=2, 第一个是时间戳, 第二个是值"`
}

// path: $.dropped_entries
type Dropped struct {
	Labels    map[string]string `json:"labels"`
	Timestamp string            `json:"timestamp"`
}
