package main

import (
	"fmt"
	"html/template"
)

var (
	funcs     = template.FuncMap{"increment": func(i int) int { i++; return i }}
	templates = template.Must(template.New("fileSharing").Funcs(funcs).ParseFiles("res/view.html", "res/login.html", "res/register.html"))
)

type TemplateBase struct {
	Errors map[string]string
}

func (self *TemplateBase) AddError(name string, errorValue string, params ...interface{}) {
	self.Errors[name] = fmt.Sprintf(errorValue, params...)
}

type TemplatePassword struct {
	TemplateBase
	Password string
}

func NewTemplatePassword() *TemplatePassword {
	var info TemplatePassword
	info.Errors = make(map[string]string)
	return &info
}

type TemplateRegister struct {
	TemplateBase
	StorageName string
	Password    string
}

func NewTemplateRegister() *TemplateRegister {
	return &TemplateRegister{
		TemplateBase: TemplateBase{
			Errors: make(map[string]string),
		},
	}
}
