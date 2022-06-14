package utils

type Empty struct{}
type StringSet map[string]Empty

func NewStringSet(items ...string) StringSet {
	ss := StringSet{}
	ss.Insert(items...)
	return ss
}

func (s StringSet) Insert(items ...string) StringSet {
	for _, item := range items {
		s[item] = Empty{}
	}
	return s
}

func (s StringSet) Has(item string) bool {
	_, contained := s[item]
	return contained
}

func (s StringSet) List() []string {
	res := make([]string, 0, len(s))
	for key := range s {
		res = append(res, key)
	}
	return res
}
