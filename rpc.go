package bloomfilter

type BFType struct {
	bf *Bloomfilter
}

func (b *BFType) Add(elems [][]byte, count *int) error {
	k := 0
	for _, elem := range elems {
		b.bf.Add(elem)
		k++
	}
	*count = k
	return nil
}

func (b *BFType) Check(elems [][]byte, checks *[]bool) error {
	checkRes := make([]bool, len(elems))
	for i, elem := range elems {
		checkRes[i] = b.bf.Check(elem)
	}
	*checks = checkRes

	return nil
}

func (b *BFType) Union(bf *Bloomfilter, count *float64) error {
	var err error
	*count, err = b.bf.Union(bf)

	return err
}
