package routers

import (
	"api"
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

	beego.Router("/login", &controllers.AuthController{}, "post:Login")
	beego.Router("/log_out", &controllers.AuthController{}, "get:LogOut")
	beego.Router("/reset", &controllers.AuthController{}, "post:ResetPassword")
	beego.Router("/userExists", &controllers.AuthController{}, "post:UserExists")
	beego.Router("/sendEmail", &controllers.AuthController{}, "get:SendEmail")
	beego.Router("/language", &controllers.AuthController{}, "get:SwitchLanguage")

	beego.Router("/optional_menu", &controllers.OptionalMenuController{})
	beego.Router("/navigation_header", &controllers.NavigationHeaderController{})
	beego.Router("/navigation_detail", &controllers.NavigationDetailController{})
	beego.Router("/sign_in", &controllers.SignInController{})

	beego.Router("/stack/:name", &api.StackAPI{}, "put:Deploy;delete:Delete")
	beego.Router("/stack/services", &api.StackAPI{}, "get:CheckServices")
	beego.Router("/stacks", &api.StackAPI{}, "get:List")
	beego.Router("/stack/log", &api.StackAPI{}, "get:Log")

}
