package mtsm

type TsVector struct {
	Metric map[string]string `json:"metric"`
	Value  []interface{}     `json:"value,omitempty"`
	Dts    []float64         `json:"dts,omitempty" note:"时间戳"`
	Val    []float64         `json:"val,omitempty" note:"值"`
}

// ValueToTv 将 Vector 的 Value 字段转换为 Valuet 和 Valuev 切片。
func (t *TsVector) ValueToTv() {
	if t.Value == nil {
		return
	}
	t.Dts = []float64{}
	t.Val = []float64{}
	t.Dts = append(t.Dts, convertToFloat64(t.Value[0]))
	t.Val = append(t.Val, convertToFloat64(t.Value[1]))
	t.Value = nil
}
