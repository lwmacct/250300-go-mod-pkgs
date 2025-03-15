package mtsm

type Clean struct {
	Global  map[string]string `json:"global" note:"全局标签"`
	Metrics string            `json:"metrics"`
	Vector  []TsVector        `json:"vector"`
	Matrix  []TsMatrix        `json:"matrix"`
	Stats   struct {
		VectorLen int `json:"vector_len" note:"Vector 长度"`
		MatrixLen int `json:"matrix_len" note:"Matrix 长度"`
	} `json:"stats"`
}

type CleanOpts func(*Clean)

func NewClean(opts ...CleanOpts) *Clean {
	t := &Clean{}
	t.Global = map[string]string{}
	for _, opt := range opts {
		opt(t)
	}

	return t
}

func WithCleanVector(vector []TsVector) CleanOpts {
	return func(t *Clean) {
		t.Vector = vector
		t.Stats.VectorLen = len(t.Vector)
	}
}

func WithCleanMatrix(matrix []TsMatrix) CleanOpts {
	return func(t *Clean) {
		t.Matrix = matrix
		t.Stats.MatrixLen = len(t.Matrix)
	}
}

func (t *Clean) ToTvMatrix() error {
	for i := range t.Matrix {
		t.Matrix[i].ValueToTv()
	}
	return nil
}

func (t *Clean) ToTvVector() error {
	for i := range t.Vector {
		t.Vector[i].ValueToTv()
	}
	return nil
}

func (t *Clean) ClipLabels() {
	for i := range t.Vector {
		t.Metrics = t.Vector[i].Metric["__name__"]
		t.Vector[i].Metric = t.setGlobal(t.Vector[i].Metric)
	}

	for i := range t.Matrix {
		if i == 0 {
			t.Metrics = t.Matrix[i].Metric["__name__"]
		}
		t.Matrix[i].Metric = t.setGlobal(t.Matrix[i].Metric)
	}
}

// 删除 mapd 中的 __name__ 字段
// 将前缀 "g_" 标签加入到全局, 并移除
func (t *Clean) setGlobal(mapd map[string]string) map[string]string {
	findName := false
	for key, value := range mapd {
		if key[0:2] == "g_" {
			t.Global[key[2:]] = value
			delete(mapd, key)
		}
		if !findName && key == "__name__" {
			delete(mapd, key)
			findName = true
		}
	}
	return mapd
}
