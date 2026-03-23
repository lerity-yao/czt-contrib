package rabbitmq

import (
	"fmt"

	"github.com/lerity-yao/czt-contrib/cztctl/util/pathx"
)

const (
	category             = "rabbitmq"
	configTemplateFile   = "config.tpl"
	contextTemplateFile  = "context.tpl"
	etcTemplateFile      = "etc.tpl"
	handlerTemplateFile  = "handler.tpl"
	listenerTemplateFile = "listener.tpl"
	logicTemplateFile    = "logic.tpl"
	mainTemplateFile     = "main.tpl"
	typesTemplateFile    = "types.tpl"
)

var templates = map[string]string{
	configTemplateFile:   configTemplate,
	contextTemplateFile:  contextTemplate,
	etcTemplateFile:      etcTemplate,
	handlerTemplateFile:  handlerTemplate,
	logicTemplateFile:    logicTemplate,
	mainTemplateFile:     mainTemplate,
	listenerTemplateFile: listenerTemplate,
	typesTemplateFile:    typesTemplate,
}

// Category returns the category of the rabbitmq files.
func Category() string {
	return category
}

// Clean cleans the generated deployment files.
func Clean() error {
	return pathx.Clean(category)
}

// GenTemplates generates rabbitmq template files.
func GenTemplates() error {
	return pathx.InitTemplates(category, templates)
}

// RevertTemplate reverts the given template file to the default value.
func RevertTemplate(name string) error {
	content, ok := templates[name]
	if !ok {
		return fmt.Errorf("%s: no such file name", name)
	}
	return pathx.CreateTemplate(category, name, content)
}

// Update updates the template files to the templates built in current cztctl.
func Update() error {
	err := Clean()
	if err != nil {
		return err
	}
	return pathx.InitTemplates(category, templates)
}
