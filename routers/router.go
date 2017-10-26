package routers

import (
	"arrowcloudapi/api"
	"arrowcloudapi/controllers"

	"github.com/astaxie/beego"
)

func init() {

	beego.SetStaticPath("static/resources", "static/resources")
	beego.SetStaticPath("static/vendors", "static/vendors")

	// beego.Router("/", &controllers.MainController{})

	//Page Controllers:
	beego.Router("/login", &controllers.AuthController{}, "post:Login")
	beego.Router("/log_out", &controllers.AuthController{}, "get:LogOut")

	beego.Router("/stack/:name", &api.StackAPI{}, "post:Deploy;delete:Delete")
	beego.Router("/stack/services", &api.StackAPI{}, "get:CheckServices")
	beego.Router("/stacks", &api.StackAPI{}, "get:List")
	beego.Router("/stack/log", &api.StackAPI{}, "get:GetServiceLog")

}
