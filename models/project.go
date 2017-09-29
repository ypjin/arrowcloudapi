package models

import (
	"time"
)

// Project holds the details of a project.
type Project struct {
	ProjectID       int64     `orm:"pk;column(project_id)" json:"project_id"`
	OwnerID         int       `orm:"column(owner_id)" json:"owner_id"`
	Name            string    `orm:"column(name)" json:"name"`
	CreationTime    time.Time `orm:"column(creation_time)" json:"creation_time"`
	CreationTimeStr string    `orm:"-" json:"creation_time_str"`
	Deleted         int       `orm:"column(deleted)" json:"deleted"`
	//UserID          int `json:"UserId"`
	OwnerName string `orm:"-" json:"owner_name"`
	Public    int    `orm:"column(public)" json:"public"`
	//This field does not have correspondent column in DB, this is just for UI to disable button
	Togglable bool `orm:"-"`

	UpdateTime time.Time `orm:"update_time" json:"update_time"`
	Role       int       `orm:"-" json:"current_user_role_id"`
	RepoCount  int       `orm:"-" json:"repo_count"`
}

// ProjectSorter holds an array of projects
type ProjectSorter struct {
	Projects []Project
}

// Len returns the length of array in ProjectSorter
func (ps *ProjectSorter) Len() int {
	return len(ps.Projects)
}

// Less defines the comparison rules of project
func (ps *ProjectSorter) Less(i, j int) bool {
	return ps.Projects[i].Name < ps.Projects[j].Name
}

// Swap swaps the position of i and j
func (ps *ProjectSorter) Swap(i, j int) {
	ps.Projects[i], ps.Projects[j] = ps.Projects[j], ps.Projects[i]
}
