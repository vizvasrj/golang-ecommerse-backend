package common

// UserRole type
type UserRole string

const (
	RoleAdmin    UserRole = "ROLE ADMIN"
	RoleMember   UserRole = "ROLE MEMBER"
	RoleMerchant UserRole = "ROLE MERCHANT"
)

func (r UserRole) String() string {
	return string(r)
}

// GetUserRoleFromString takes a string and returns the corresponding UserRole type.
func GetUserRoleFromString(role string) UserRole {
	switch role {
	case "ROLE ADMIN":
		return RoleAdmin
	case "ROLE MEMBER":
		return RoleMember
	case "ROLE MERCHANT":
		return RoleMerchant
	default:
		// Return an empty UserRole if the input string does not match any valid role
		return ""
	}
}

func GetUserRole(role interface{}) UserRole {
	roleStr, ok := role.(string)
	if !ok {
		// Return an empty UserRole if the input cannot be decoded to a string
		return ""
	}

	switch roleStr {
	case "ROLE ADMIN":
		return RoleAdmin
	case "ROLE MEMBER":
		return RoleMember
	case "ROLE MERCHANT":
		return RoleMerchant
	default:
		// Return an empty UserRole if the input string does not match any valid role
		return ""
	}
}
