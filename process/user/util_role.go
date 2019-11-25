package user

var gRoleMap = map[string]bool{
	"CRC":  true,
	"CRA":  true,
	"SUBI": true,
	"PI":   true,
	"SM":   true,
	"DM":   true,
}

func IsAvailableRole(role string) bool {
	has, _ := gRoleMap[role]
	return has
}
