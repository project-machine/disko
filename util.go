package disko

import "fmt"

type uRange struct {
	Start, End uint64
}

func (r *uRange) Size() uint64 {
	return r.End - r.Start
}

// findRangeGaps returns a set of uRange to represent the un-used
// uint64 between min and max that are not included in ranges.
//  findRangeGaps({{10, 40}, {50, 100}}, 0, 110}) ==
//      {{0, 9}, {41, 49}, {101, 110}}
func findRangeGaps(ranges []uRange, min, max uint64) []uRange {
	// start 'ret' off with full range of min to max, then start cutting it up.
	ret := []uRange{{min, max}}

	for _, i := range ranges {
		for r := 0; r < len(ret); r++ {
			// 5 cases:
			if i.Start > ret[r].End || i.End < ret[r].Start {
				// a. i has no overlap
			} else if i.Start <= ret[r].Start && i.End >= ret[r].End {
				// b.) i is complete superset, so remove ret[r]
				ret = append(ret[:r], ret[r+1:]...)
				r--
			} else if i.Start > ret[r].Start && i.End < ret[r].End {
				// c.) i is strict subset: split ret[r]
				ret = append(
					append(ret[:r+1], uRange{i.End + 1, ret[r].End}),
					ret[r+1:]...)
				ret[r].End = i.Start - 1
				r++ // added entry is guaranteed to be 'a', so skip it.
			} else if i.Start <= ret[r].Start {
				// d.) overlap left edge to middle
				ret[r].Start = i.End + 1
			} else if i.Start <= ret[r].End {
				// e.) middle to right edge (possibly past).
				ret[r].End = i.Start - 1
			} else {
				panic(fmt.Sprintf("Error in findRangeGaps: %v, r=%d, ret=%v",
					i, r, ret))
			}
		}
	}

	return ret
}
