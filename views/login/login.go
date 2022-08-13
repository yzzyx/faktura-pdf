package login

import (
	"net/http"

	"github.com/yzzyx/faktura-pdf/models"
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
		if v.Session.ID != "" {
			err := models.SessionRemove(v.Ctx, v.Session.ID)
			if err != nil {
				return err
			}
			v.SetData("logged_in", false)
			v.SetData("session", nil)
			v.Session = models.Session{}
			v.SetCookie(&http.Cookie{Name: "_fp_login", MaxAge: -1})
		}
	}
	v.SetData("r", v.FormValueString("r"))

	return v.Render("login.html")
}

// HandlePost handles the login process
func (v *Login) HandlePost() error {

	username := v.FormValueString("username")
	password := v.FormValueString("password")
	redirect := v.FormValueString("r")

	user, err := models.UserGet(v.Ctx, models.UserFilter{Username: username})
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

	companyList, err := models.CompanyList(v.Ctx, models.CompanyFilter{UserID: user.ID})
	if err != nil {
		return err
	}

	s := models.Session{
		User: user,
	}

	if len(companyList) == 1 {
		s.Company = companyList[0]
	}

	s.ID, err = models.SessionSave(v.Ctx, s)
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

	if len(companyList) == 0 {
		// Redirect to company creation page
		return v.RedirectRoute("company-view", "id", "-1")
	} else if len(companyList) > 1 {
		// Redirect to company selection page
		u, err := v.URL("company-list")
		if err != nil {
			return err
		}
		q := u.Query()
		q.Add("r", redirect)
		u.RawQuery = q.Encode()
		v.Redirect(u.String())
		return nil
	}

	if redirect != "" {
		v.Redirect(redirect)
		return nil
	}
	return v.RedirectRoute("start")
}
