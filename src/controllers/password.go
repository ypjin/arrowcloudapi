package controllers

import (
	"bytes"
	"net/http"
	"os"
	"regexp"
	"text/template"

	"dao"
	"models"
	"utils"
	"utils/log"

	"github.com/astaxie/beego"
)

type messageDetail struct {
	Hint string
	URL  string
	UUID string
}

// SendEmail verifies the Email address and contact SMTP server to send reset password Email.
func (ac *AuthController) SendEmail() {

	email := ac.GetString("email")

	pass, _ := regexp.MatchString(`^(([^<>()[\]\\.,;:\s@\"]+(\.[^<>()[\]\\.,;:\s@\"]+)*)|(\".+\"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$`, email)

	if !pass {
		ac.CustomAbort(http.StatusBadRequest, "email_content_illegal")
	} else {

		queryUser := models.User{Email: email}
		exist, err := dao.UserExists(queryUser, "email")
		if err != nil {
			log.Errorf("Error occurred in UserExists: %v", err)
			ac.CustomAbort(http.StatusInternalServerError, "Internal error.")
		}
		if !exist {
			ac.CustomAbort(http.StatusNotFound, "email_does_not_exist")
		}

		messageTemplate, err := template.ParseFiles("views/reset-password-mail.tpl")
		if err != nil {
			log.Errorf("Parse email template file failed: %v", err)
			ac.CustomAbort(http.StatusInternalServerError, err.Error())
		}

		message := new(bytes.Buffer)

		harborURL := os.Getenv("HARBOR_URL")
		if harborURL == "" {
			harborURL = "localhost"
		}
		uuid := utils.GenerateRandomString()
		err = messageTemplate.Execute(message, messageDetail{
			Hint: ac.Tr("reset_email_hint"),
			URL:  harborURL,
			UUID: uuid,
		})

		if err != nil {
			log.Errorf("Message template error: %v", err)
			ac.CustomAbort(http.StatusInternalServerError, "internal_error")
		}

		config, err := beego.AppConfig.GetSection("mail")
		if err != nil {
			log.Errorf("Can not load app.conf: %v", err)
			ac.CustomAbort(http.StatusInternalServerError, "internal_error")
		}

		mail := utils.Mail{
			From:    config["from"],
			To:      []string{email},
			Subject: ac.Tr("reset_email_subject"),
			Message: message.String()}

		err = mail.SendMail()

		if err != nil {
			log.Errorf("Send email failed: %v", err)
			ac.CustomAbort(http.StatusInternalServerError, "send_email_failed")
		}

		user := models.User{ResetUUID: uuid, Email: email}
		dao.UpdateUserResetUUID(user)

	}

}

// ForgotPasswordController handles requests to /forgot_password
type ForgotPasswordController struct {
	BaseController
}

// Get renders forgot password page
func (fpc *ForgotPasswordController) Get() {
	fpc.Forward("page_title_forgot_password", "forgot-password.htm")
}

// ResetPasswordController handles request to /resetPassword
type ResetPasswordController struct {
	BaseController
}

// Get checks if reset_uuid in the reset link is valid and render the result page for user to reset password.
func (rpc *ResetPasswordController) Get() {

	resetUUID := rpc.GetString("reset_uuid")
	if resetUUID == "" {
		log.Error("Reset uuid is blank.")
		rpc.Redirect("/", http.StatusFound)
		return
	}

	queryUser := models.User{ResetUUID: resetUUID}
	user, err := dao.GetUser(queryUser)
	if err != nil {
		log.Errorf("Error occurred in GetUser: %v", err)
		rpc.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}

	if user != nil {
		rpc.Data["ResetUuid"] = user.ResetUUID
		rpc.Forward("page_title_reset_password", "reset-password.htm")
	} else {
		rpc.Redirect("/", http.StatusFound)
	}
}

// ResetPassword handles request from the reset page and reset password
func (ac *AuthController) ResetPassword() {

	resetUUID := ac.GetString("reset_uuid")
	if resetUUID == "" {
		ac.CustomAbort(http.StatusBadRequest, "Reset uuid is blank.")
	}

	queryUser := models.User{ResetUUID: resetUUID}
	user, err := dao.GetUser(queryUser)
	if err != nil {
		log.Errorf("Error occurred in GetUser: %v", err)
		ac.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
	if user == nil {
		log.Error("User does not exist")
		ac.CustomAbort(http.StatusBadRequest, "User does not exist")
	}

	password := ac.GetString("password")

	if password != "" {
		user.Password = password
		err = dao.ResetUserPassword(*user)
		if err != nil {
			log.Errorf("Error occurred in ResetUserPassword: %v", err)
			ac.CustomAbort(http.StatusInternalServerError, "Internal error.")
		}
	} else {
		ac.CustomAbort(http.StatusBadRequest, "password_is_required")
	}
}
