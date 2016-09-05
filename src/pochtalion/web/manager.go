package web

import (
	"bufio"
	"bytes"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strings"

	"pochtalion"

	"github.com/labstack/echo"
	valid "gopkg.in/asaskevich/govalidator.v4"
)

var t *Template

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func SetupRouting(e *echo.Echo) {
	t = &Template{
		templates: template.Must(template.ParseGlob("public/*.html")),
	}

	e.SetRenderer(t)
	e.Get("/", NewMailing)
	e.Get("/newmailing", NewMailing)
	e.Post("/newmailing", CreateNewMailing)
}

func NewMailing(ctx echo.Context) error {
	var data = map[string]interface{}{}

	countSendingRaw := ctx.Request().Header().Get("sending")
	if len(countSendingRaw) > 0 {
		data["Sending"] = true
		data["Length"] = countSendingRaw
	}

	return ctx.Render(http.StatusOK, "newmailing", data)
}

type NewMailingForm struct {
	Title     string `form:"title"`
	From      string `form:"from"`
	RawEmails string `form:"emails"`
	RawText   string `form:"body"`

	Emails []string `form:"-" json:"emails"`

	Sending  bool   `form:"-" json:"-"`
	Error    bool   `form:"-" json:"-"`
	ErrorMsg string `form:"-" json:"-"`
}

func (f *NewMailingForm) ParseEmails() {
	r := bufio.NewReader(bytes.NewReader([]byte(f.RawEmails)))
	var err error
	var line []byte

	for ; err == nil; line, _, err = r.ReadLine() {
		email := string(line)
		if !valid.IsEmail(email) {
			continue
		}
		// break if EOF

		f.Emails = append(f.Emails, email)
	}
}

func (f *NewMailingForm) Format() {
	f.From = strings.TrimSpace(f.From)
	f.Title = strings.TrimSpace(f.Title)
	f.RawEmails = strings.TrimSpace(f.RawEmails)
	f.RawText = strings.TrimSpace(f.RawText)
}

func (f *NewMailingForm) Valid() error {
	if len(f.From) == 0 {
		return fmt.Errorf("from empty")
	}

	if len(f.Title) == 0 {
		return fmt.Errorf("title empty")
	}

	if len(f.RawEmails) == 0 || len(f.Emails) == 0 {
		return fmt.Errorf("emails empty")
	}

	if len(f.RawText) == 0 {
		return fmt.Errorf("text empty")
	}

	return nil
}

func CreateNewMailing(ctx echo.Context) error {
	form := new(NewMailingForm)
	if err := ctx.Bind(form); err != nil {
		form.ErrorMsg = fmt.Sprintln("Не валидные данные формы", err)
		form.Error = true
	}

	form.Format()
	form.ParseEmails()
	if err := form.Valid(); err != nil {
		form.ErrorMsg = fmt.Sprintln("Не валидные данные формы", err)
		form.Error = true
	}

	if form.Error {
		return ctx.Render(http.StatusBadRequest, "newmailing", form)
	}

	pochtalion.Sender.SendMailling(form.From, form.Title, form.RawText, form.Emails...)

	return ctx.Render(http.StatusBadRequest, "sending", map[string]interface{}{
		"Length":    len(form.Emails),
		"EmailsRaw": strings.Join(form.Emails, "\n"),
	})
}
