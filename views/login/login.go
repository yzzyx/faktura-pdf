package login

import (
	"net/http"

	"github.com/yzzyx/faktura-pdf/models"
	"github.com/yzzyx/faktura-pdf/session"
	"github.com/yzzyx/faktura-pdf/views"
)

// Login is the view-handler for the login-page
type Login struct {
	views.View
}

// New creates a new handler for the start page
func New() *Login {
	return &Login{}
}

// HandleGet shows the login page
func (v *Login) HandleGet() error {

	if v.FormValueBool("logout") {
		sr := v.GetData("session")
		if s, ok := sr.(session.Session); ok {
			session.Clear(s.ID)
			v.SetData("logged_in", false)
			v.SetData("session", nil)
			v.SetCookie(&http.Cookie{Name: "_fp_login", MaxAge: -1})
		}
	}
	return v.Render("login.html")
}

// HandlePost handles the login process
func (v *Login) HandlePost() error {

	username := v.FormValueString("username")
	password := v.FormValueString("password")

	user, err := models.UserGet(v.Ctx, username)
	if err != nil {
		return err
	}

	passwordValid, err := user.ValidatePassword(password)
	if err != nil {
		return err
	}

	if !passwordValid {
		v.SetData("invalidPassword", true)
		return v.Render("login.html")
	}

	s, err := session.New(user)
	if err != nil {
		return err
	}

	cookie := &http.Cookie{
		Name:  "_fp_login",
		Value: s.ID,
		//Secure:     false, // FIXME - set to secure if HTTPS is available
		HttpOnly: true,
	}
	v.SetCookie(cookie)

	return v.RedirectRoute("start")
}
