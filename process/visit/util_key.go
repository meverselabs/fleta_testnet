package visit

// tags
var (
	tagVisit = []byte{1, 0}
)

func toVisitKey(SubjectID string) []byte {
	bs := make([]byte, 2+len(SubjectID))
	copy(bs, tagVisit)
	copy(bs[2:], []byte(SubjectID))
	return bs
}
