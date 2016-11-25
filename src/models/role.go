package models

const (
	//PROJECTADMIN project administrator
	PROJECTADMIN = 1
	//DEVELOPER developer
	DEVELOPER = 2
	//GUEST guest
	GUEST = 3
)

// Role holds the details of a role.
type Role struct {
	RoleID   int    `orm:"pk;column(role_id)" json:"role_id"`
	RoleCode string `orm:"column(role_code)" json:"role_code"`
	Name     string `orm:"column(name)" json:"role_name"`

	RoleMask int `orm:"role_mask" json:"role_mask"`
}
