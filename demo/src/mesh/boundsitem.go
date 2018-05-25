package mesh

type BoundsItem struct {
	bmin [2]float32
	bmax [2]float32
	i    int
}

type BoundsSorter struct {
	lst      []BoundsItem
	compFunc func(a, b *BoundsItem) bool
}

func NewBoundsSorter(lst []BoundsItem, compFunc func(a, b *BoundsItem) bool) *BoundsSorter {
	return &BoundsSorter{
		lst:      lst,
		compFunc: compFunc,
	}
}

func (bs *BoundsSorter) Len() int {
	return len(bs.lst)
}

func (bs *BoundsSorter) Swap(i, j int) {
	bs.lst[i], bs.lst[j] = bs.lst[j], bs.lst[i]
}

func (bs *BoundsSorter) Less(i, j int) bool {
	return bs.compFunc(&bs.lst[i], &bs.lst[j])
}
