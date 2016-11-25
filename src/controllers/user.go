package controllers

import (
	"net/http"

	"dao"
	"models"
	"service/auth"
	"utils/log"
)

// CommonController handles request from UI that doesn't expect a page, such as /SwitchLanguage /logout ...
type CommonController struct {
	BaseController
}

// Render returns nil.
func (cc *CommonController) Render() error {
	return nil
}

// Login handles login request from UI.
func (cc *CommonController) Login() {
	principal := cc.GetString("principal")
	password := cc.GetString("password")

	user, err := auth.Login(models.AuthModel{
		Principal: principal,
		Password:  password,
	})
	if err != nil {
		log.Errorf("Error occurred in UserLogin: %v", err)
		cc.CustomAbort(http.StatusUnauthorized, "")
	}

	if user == nil {
		cc.CustomAbort(http.StatusUnauthorized, "")
	}

	cc.SetSession("userId", user.UserID)
	cc.SetSession("username", user.Username)
}

// LogOut Habor UI
func (cc *CommonController) LogOut() {
	cc.DestroySession()
}

// SwitchLanguage User can swith to prefered language
func (cc *CommonController) SwitchLanguage() {
	lang := cc.GetString("lang")
	hash := cc.GetString("hash")
	if _, exist := supportLanguages[lang]; !exist {
		lang = defaultLang
	}
	cc.SetSession("lang", lang)
	cc.Data["Lang"] = lang
	cc.Redirect(cc.Ctx.Request.Header.Get("Referer")+hash, http.StatusFound)
}

// UserExists checks if user exists when user input value in sign in form.
func (cc *CommonController) UserExists() {
	target := cc.GetString("target")
	value := cc.GetString("value")

	user := models.User{}
	switch target {
	case "username":
		user.Username = value
	case "email":
		user.Email = value
	}

	exist, err := dao.UserExists(user, target)
	if err != nil {
		log.Errorf("Error occurred in UserExists: %v", err)
		cc.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
	cc.Data["json"] = exist
	cc.ServeJSON()
}
