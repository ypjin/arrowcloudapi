package controllers

import (
	"dao"
	"utils/log"
)

// AdminOptionController handles requests to /admin_option
type AdminOptionController struct {
	BaseController
}

// Get renders the admin options  page
func (aoc *AdminOptionController) Get() {
	sessionUserID, ok := aoc.GetSession("userId").(string)
	if ok {
		isAdmin, err := dao.IsAdminRole(map[string]string{"UserID": sessionUserID})
		if err != nil {
			log.Errorf("Error occurred in IsAdminRole: %v", err)
		}
		if isAdmin {
			aoc.Forward("page_title_admin_option", "admin-options.htm")
			return
		}
	}
	aoc.Redirect("/dashboard", 302)
}
