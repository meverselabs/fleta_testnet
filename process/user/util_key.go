package user

// tags
var (
	tagUser     = []byte{1, 0}
	tagUserRole = []byte{1, 1}
)

func toUserKey(ID string) []byte {
	bs := make([]byte, 2+len(ID))
	copy(bs, tagUser)
	copy(bs[2:], []byte(ID))
	return bs
}

func toUserRoleKey(ID string, Role string) []byte {
	bs := make([]byte, 2+len(ID)+1+len(Role))
	copy(bs, tagUser)
	copy(bs[2:], []byte(ID+"@"+Role))
	return bs
}
