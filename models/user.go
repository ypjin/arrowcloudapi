package models

import "time"

type Org struct {
	ID             string
	Name           string
	Admin          bool
	Node_acs_admin bool
}

// User holds the details of a user.
type User struct {
	ID        string `orm:"pk;column(_id)" json:"user_id"`
	Username  string `orm:"column(username)" json:"username"`
	Email     string `orm:"column(email)" json:"email"`
	Firstname string `orm:"column(firstname)" json:"firstname"`
	Lastname  string `orm:"column(lastname)" json:"lastname"`
	Orgs      []Org  `orm:"-" json:"orgs"`

	Password string `orm:"column(password)" json:"password"`
	Realname string `orm:"column(realname)" json:"realname"`
	Comment  string `orm:"column(comment)" json:"comment"`
	Deleted  int    `orm:"column(deleted)" json:"deleted"`
	Rolename string `json:"role_name"`
	//if this field is named as "RoleID", beego orm can not map role_id
	//to it.
	Role int `json:"role_id"`
	//	RoleList     []Role `json:"role_list"`
	HasAdminRole int       `orm:"column(sysadmin_flag)" json:"has_admin_role"`
	ResetUUID    string    `orm:"column(reset_uuid)" json:"reset_uuid"`
	Salt         string    `orm:"column(salt)" json:"-"`
	CreationTime time.Time `orm:"creation_time" json:"creation_time"`
	UpdateTime   time.Time `orm:"update_time" json:"update_time"`
}
