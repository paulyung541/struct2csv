package struct2csv

type HeaderAutoIncrementConv struct {
	max uint64
}

func NewHeaderAutoIncrementConv() *HeaderAutoIncrementConv {
	return &HeaderAutoIncrementConv{}
}

func (h *HeaderAutoIncrementConv) ConvertHeader(s string) KeyType {
	hs := newKeyAuto(h.max)
	h.max = uint64(hs)
	return hs
}

type HeaderOriginalStringConv struct {
}

func NewHeaderOriginalStringConv() *HeaderOriginalStringConv {
	return &HeaderOriginalStringConv{}
}

func (h *HeaderOriginalStringConv) ConvertHeader(s string) KeyType {
	return newKeyString(s)
}
