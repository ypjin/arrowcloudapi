package controllers

import (
	"net/http"

	"arrowcloudapi/dao"
	"arrowcloudapi/models"
	"arrowcloudapi/service/auth"
	"arrowcloudapi/utils/log"
)

// AuthController handles user authentication requests
type AuthController struct {
	BaseController
}

// Render returns nil.
func (ac *AuthController) Render() error {
	return nil
}

// Login handles login request from UI.
func (ac *AuthController) Login() {
	principal := ac.GetString("principal")
	password := ac.GetString("password")

	user, err := auth.Login(models.AuthModel{
		Principal: principal,
		Password:  password,
	})
	if err != nil {
		log.Errorf("Error occurred in UserLogin: %v", err)
		ac.CustomAbort(http.StatusUnauthorized, "")
	}

	if user == nil {
		ac.CustomAbort(http.StatusUnauthorized, "")
	}

	ac.SetSession("userId", user.ID)
	ac.SetSession("username", user.Username)
}

// LogOut Habor UI
func (ac *AuthController) LogOut() {
	ac.DestroySession()
}

// SwitchLanguage User can swith to prefered language
func (ac *AuthController) SwitchLanguage() {
	lang := ac.GetString("lang")
	hash := ac.GetString("hash")
	if _, exist := supportLanguages[lang]; !exist {
		lang = defaultLang
	}
	ac.SetSession("lang", lang)
	ac.Data["Lang"] = lang
	ac.Redirect(ac.Ctx.Request.Header.Get("Referer")+hash, http.StatusFound)
}

// UserExists checks if user exists when user input value in sign in form.
func (ac *AuthController) UserExists() {
	target := ac.GetString("target")
	value := ac.GetString("value")

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
		ac.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
	ac.Data["json"] = exist
	ac.ServeJSON()
}
