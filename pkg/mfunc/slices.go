package mfunc

import "regexp"

var Slices = NewSlices()

type slices struct {
	regexpCache map[string]*regexp.Regexp
}
type slicesOpts func(*slices)

// 传入一个正则表达式, 删除匹配的成员
func (s *slices) RemoveMatch(array []string, regex string) []string {
	var re *regexp.Regexp
	// 使用缓存的正则表达式
	if cached, ok := s.regexpCache[regex]; ok {
		re = cached
	} else {
		re = regexp.MustCompile(regex)
		// 存入缓存
		s.regexpCache[regex] = re
	}

	// 使用过滤方式而不是在遍历中直接删除元素
	result := make([]string, 0, len(array))
	for _, str := range array {
		if !re.MatchString(str) {
			result = append(result, str)
		}
	}
	return result
}

func NewSlices(opts ...slicesOpts) *slices {
	t := &slices{
		regexpCache: make(map[string]*regexp.Regexp),
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}
