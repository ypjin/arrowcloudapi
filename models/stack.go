package models

import "time"

// Stack holds the details of a service stack.
type Stack struct {
	ID              string    `orm:"pk;column(_id)" json:"stack_id"`
	Name            string    `orm:"column(name)" json:"name"`
	UserID          string    `orm:"column(user_id)" json:"user_id"`
	OrgID           string    `orm:"column(org_id)" json:"org_id"`
	CreationTime    time.Time `orm:"column(creation_time)" json:"creation_time"`
	CreationTimeStr string    `orm:"-" json:"creation_time_str"`
	Deleted         int       `orm:"column(deleted)" json:"deleted"`
	//UserID          int `json:"UserId"`
	UserName string `orm:"-" json:"user_name"`

	UpdateTime  time.Time `orm:"update_time" json:"update_time"`
	ComposeFile string    `orm:"column(compose_file)" json:"compose_file"`
}
