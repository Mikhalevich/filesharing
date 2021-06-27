package template

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
)

var (
	//go:embed html
	content embed.FS

	//go:embed res
	resources embed.FS

	funcs       = template.FuncMap{"increment": func(i int) int { i++; return i }}
	pcTemplates = template.Must(template.New("fileSharing").Funcs(funcs).ParseFS(content, "html/*.html"))
)

func Resources() fs.FS {
	return resources
}

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
	Title             string
	NeedPermanentLink bool
	FileInfoList      []FileInfo
}

func NewTemplateView(title string, needPermanentLink bool, list []FileInfo) *TemplateView {
	return &TemplateView{
		TemplateBase:      *NewTemplateBase("view.html"),
		Title:             title,
		NeedPermanentLink: needPermanentLink,
		FileInfoList:      list,
	}
}

func (t *TemplateView) Execute(wr io.Writer) error {
	return t.TemplateBase.ExecuteTemplate(wr, *t)
}
