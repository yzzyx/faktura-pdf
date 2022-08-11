package register

import (
	"net/http"

	"github.com/yzzyx/faktura-pdf/models"
	"github.com/yzzyx/faktura-pdf/session"
	"github.com/yzzyx/faktura-pdf/views"
)

// Register is the view-handler for the register-page
type Register struct {
	views.View
}

// Register creates a new handler for the registration page
func New() *Register {
	return &Register{}
}

// HandleGet shows the registration page
func (v *Register) HandleGet() error {
	return v.Render("register.html")
}

// HandlePost handles the login process
func (v *Register) HandlePost() error {

	username := v.FormValueString("username")
	name := v.FormValueString("name")
	password := v.FormValueString("password")

	user, err := models.UserGet(v.Ctx, username)
	if err != nil {
		return err
	}

	if user.ID > 0 {
		v.SetData("userExists", true)
		return v.Render("register.html")
	}

	if len(password) < 8 {
		v.SetData("simplePassword", true)
		return v.Render("register.html")
	}

	user = models.User{
		Username: username,
		Name:     name,
		Email:    username,
	}

	err = user.SetPassword(password)
	if err != nil {
		return err
	}

	// FIXME - add support for email validation

	user.ID, err = models.UserSave(v.Ctx, user)
	if err != nil {
		return err
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
