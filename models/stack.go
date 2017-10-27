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

	UpdateTime             time.Time `orm:"update_time" json:"update_time"`
	OriginalComposeFile    string    `orm:"column(compose_file_original)" json:"compose_file_original"`
	TransformedComposeFile string    `orm:"column(compose_file_transformed)" json:"compose_file_transformed"`
	// comma separated NFS folder names used for volumes (no path info)
	VolumeFolders string `orm:"column(volume_folders)" json:"volume_folders"`
	// comma separated service names which need to be exposed
	PublicServices string `orm:"column(public_services)" json:"public_services"`
}
