package controllers

// IndexController handles request to /
type IndexController struct {
	BaseController
}

// Get renders the index page
func (ic *IndexController) Get() {
	// controllers/base.go
	// https://beego.me/docs/mvc/view/view.md
	ic.Forward("page_title_index", "index.htm")
}
