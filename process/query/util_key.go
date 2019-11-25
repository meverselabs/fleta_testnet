package query

// tags
var (
	tagQuery = []byte{1, 0}
)

func toQueryKey(QueryID string) []byte {
	bs := make([]byte, 2+len(QueryID))
	copy(bs, tagQuery)
	copy(bs[2:], []byte(QueryID))
	return bs
}
