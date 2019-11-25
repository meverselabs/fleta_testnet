package subject

// tags
var (
	tagSubject = []byte{1, 0}
)

func toSubjectKey(ID string) []byte {
	bs := make([]byte, 2+len(ID))
	copy(bs, tagSubject)
	copy(bs[2:], []byte(ID))
	return bs
}
