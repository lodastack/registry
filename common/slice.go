package common

func ContainString(sl []string, v string) (int, bool) {
	for index, vv := range sl {
		if vv == v {
			return index, true
		}
	}
	return 0, false
}

func AddIfNotContain(sl []string, v string) ([]string, bool) {
	if v == "" {
		return sl, false
	}
	if _, ok := ContainString(sl, v); !ok {
		sl = append(sl, v)
		return sl, true
	}
	return sl, false
}

// RemoveIfContain remove the first string in the string list.
func RemoveIfContain(sl []string, v string) ([]string, bool) {
	index, ok := ContainString(sl, v)
	if !ok {
		return sl, false
	}
	copy(sl[index:], sl[index+1:])
	sl = sl[:len(sl)-1]
	return sl, true
}

func Reverse(ori []string) []string {
	if len(ori) == 0 {
		return nil
	}
	result := make([]string, len(ori))
	for i, j := 0, len(ori)-1; i <= j; i, j = i+1, j-1 {
		result[i], result[j] = ori[j], ori[i]
	}
	return result
}
