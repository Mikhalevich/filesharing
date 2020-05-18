package templates

import (
	"fmt"
	"html/template"
	"io"
	"os"
	"path"
	"path/filepath"
)

func templatePath(name string) string {
	executable, err := os.Executable()
	if err != nil {
		return ""
	}

	return path.Join(filepath.Dir(executable), "templates/html", name)
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

// FileInfo represents one file info for templating
type FileInfo struct {
	Name    string
	Size    int64
	ModTime int64
}

type TemplateView struct {
	TemplateBase
	Title        string
	FileInfoList []FileInfo
}

func NewTemplateView(title string, list []FileInfo) *TemplateView {
	return &TemplateView{
		TemplateBase: *NewTemplateBase("view.html"),
		Title:        title,
		FileInfoList: list,
	}
}

func (t *TemplateView) Execute(wr io.Writer) error {
	return t.TemplateBase.ExecuteTemplate(wr, *t)
}
