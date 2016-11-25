package routers

import (
	"controllers"

	"github.com/astaxie/beego"
)

func init() {

	beego.SetStaticPath("static/resources", "static/resources")
	beego.SetStaticPath("static/vendors", "static/vendors")

	// beego.Router("/", &controllers.MainController{})

	//Page Controllers:
	beego.Router("/", &controllers.IndexController{})
	beego.Router("/dashboard", &controllers.DashboardController{})
	beego.Router("/project", &controllers.ProjectController{})
	beego.Router("/repository", &controllers.RepositoryController{})
	beego.Router("/sign_up", &controllers.SignUpController{})
	beego.Router("/add_new", &controllers.AddNewController{})
	beego.Router("/account_setting", &controllers.AccountSettingController{})
	beego.Router("/change_password", &controllers.ChangePasswordController{})
	beego.Router("/admin_option", &controllers.AdminOptionController{})
	beego.Router("/forgot_password", &controllers.ForgotPasswordController{})
	beego.Router("/reset_password", &controllers.ResetPasswordController{})
	beego.Router("/search", &controllers.SearchController{})

	beego.Router("/login", &controllers.CommonController{}, "post:Login")
	beego.Router("/log_out", &controllers.CommonController{}, "get:LogOut")
	beego.Router("/reset", &controllers.CommonController{}, "post:ResetPassword")
	beego.Router("/userExists", &controllers.CommonController{}, "post:UserExists")
	beego.Router("/sendEmail", &controllers.CommonController{}, "get:SendEmail")
	beego.Router("/language", &controllers.CommonController{}, "get:SwitchLanguage")

	beego.Router("/optional_menu", &controllers.OptionalMenuController{})
	beego.Router("/navigation_header", &controllers.NavigationHeaderController{})
	beego.Router("/navigation_detail", &controllers.NavigationDetailController{})
	beego.Router("/sign_in", &controllers.SignInController{})

}
