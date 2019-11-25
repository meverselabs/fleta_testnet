package study

// tags
var (
	tagStudy    = []byte{1, 0}
	tagTextData = []byte{2, 0}
)

func toStudyKey(ID string) []byte {
	bs := make([]byte, 2+len(ID))
	copy(bs, tagStudy)
	copy(bs[2:], []byte(ID))
	return bs
}

func toTextDataKey(ID string) []byte {
	bs := make([]byte, 2+len(ID))
	copy(bs, tagTextData)
	copy(bs[2:], []byte(ID))
	return bs
}
