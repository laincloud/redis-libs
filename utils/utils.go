package utils

func MergerSortedSlices(slices ...[]int) []int {
	if len(slices) == 0 {
		return nil
	}
	if len(slices) == 1 {
		return slices[0]
	}
	size := len(slices)/2 + (len(slices) & 1)
	res := make([][]int, size, size)
	p := 0
	for i := 0; i < len(slices)/2; i++ {
		res[p] = MergerTwoSortedSlices(slices[i*2], slices[i*2+1])
		p++
	}
	if len(slices)&1 == 1 {
		res[p] = slices[len(slices)-1]
	}
	return MergerSortedSlices(res...)
}

func MergerTwoSortedSlices(sorted1 []int, sorted2 []int) []int {
	s1, s2 := 0, 0
	len1, len2 := len(sorted1), len(sorted2)
	res := make([]int, len1+len2, len1+len2)
	p := 0
	for s1 < len1 && s2 < len2 {
		if sorted1[s1] <= sorted2[s2] {
			res[p] = sorted1[s1]
			s1++
		} else {
			res[p] = sorted2[s2]
			s2++
		}
		p++
	}
	if s1 < len1 {
		for s1 < len1 {
			res[p] = sorted1[s1]
			p++
			s1++
		}
	} else {
		for s2 < len2 {
			res[p] = sorted1[s2]
			p++
			s2++
		}
	}
	return res
}
