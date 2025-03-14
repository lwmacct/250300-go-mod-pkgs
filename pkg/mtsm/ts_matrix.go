package mtsm

type TsMatrix struct {
	Metric map[string]string `json:"metric"`
	Values [][]interface{}   `json:"values,omitempty"`
	Dts    []float64         `json:"dts,omitempty" note:"时间戳"`
	Val    []float64         `json:"val,omitempty" note:"值"`
}

// ValueToTv 将 Matrix 的 Values 字段转换为 Valuet 和 Valuev 切片。
func (m *TsMatrix) ValueToTv() {
	if m.Values == nil {
		return
	}
	m.Dts = []float64{}
	m.Val = []float64{}
	for _, pair := range m.Values {
		m.Dts = append(m.Dts, convertToFloat64(pair[0]))
		m.Val = append(m.Val, convertToFloat64(pair[1]))
	}
	m.Values = nil
}
