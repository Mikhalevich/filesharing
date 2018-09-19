package templates

import (
	"fmt"
	"html/template"
	"io"
	"path"
	"runtime"

	"github.com/Mikhalevich/filesharing/fs"
)

func templatePath(name string) string {
	_, filename, _, _ := runtime.Caller(0)
	return path.Join(path.Dir(filename), "html", name)
}

var (
	funcs       = template.FuncMap{"increment": func(i int) int { i++; return i }}
	pcTemplates = template.Must(template.New("fileSharing").Funcs(funcs).ParseFiles(
		templatePath("view.html"),
		templatePath("login.html"),
		templatePath("register.html")))
)

type TemplateBase struct {
	Name   string
	Errors map[string]string
}

func NewTemplateBase(name string) *TemplateBase {
	return &TemplateBase{
		Name:   name,
		Errors: make(map[string]string),
	}
}

func (t *TemplateBase) AddError(name string, errorValue string, params ...interface{}) {
	t.Errors[name] = fmt.Sprintf(errorValue, params...)
}

func (t *TemplateBase) ExecuteTemplate(wr io.Writer, data interface{}) error {
	return pcTemplates.ExecuteTemplate(wr, t.Name, data)
}

type TemplatePassword struct {
	TemplateBase
	Password string
}

func NewTemplatePassword() *TemplatePassword {
	return &TemplatePassword{
		TemplateBase: *NewTemplateBase("login.html"),
	}
}

func (t *TemplatePassword) Execute(wr io.Writer) error {
	return t.TemplateBase.ExecuteTemplate(wr, *t)
}

type TemplateRegister struct {
	TemplateBase
	StorageName string
	Password    string
}

func NewTemplateRegister() *TemplateRegister {
	return &TemplateRegister{
		TemplateBase: *NewTemplateBase("register.html"),
	}
}

func (t *TemplateRegister) Execute(wr io.Writer) error {
	return t.TemplateBase.ExecuteTemplate(wr, *t)
}

type TemplateView struct {
	TemplateBase
	Title        string
	FileInfoList []fs.FileInfo
}

func NewTemplateView(title string, list []fs.FileInfo) *TemplateView {
	return &TemplateView{
		TemplateBase: *NewTemplateBase("view.html"),
		Title:        title,
		FileInfoList: list,
	}
}

func (t *TemplateView) Execute(wr io.Writer) error {
	return t.TemplateBase.ExecuteTemplate(wr, *t)
}
