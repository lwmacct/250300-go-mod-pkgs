package mfunc

// NewCounter 返回一个计数器, 每次调用返回值都会加1
func NewCounter(start int) func() int {
	val := start - 1
	return func() int {
		val++
		return val
	}
}
