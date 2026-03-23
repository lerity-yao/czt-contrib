package gogen

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/lerity-yao/czt-contrib/cztctl/api/apiutil"
	"github.com/lerity-yao/czt-contrib/cztctl/api/spec"
	"github.com/lerity-yao/czt-contrib/cztctl/pkg/golang"
	"github.com/lerity-yao/czt-contrib/cztctl/util"
	"github.com/lerity-yao/czt-contrib/cztctl/util/pathx"
)

// FileGenConfig describes the configuration for generating a file.
type FileGenConfig struct {
	Dir             string
	Subdir          string
	Filename        string
	TemplateName    string
	Category        string
	TemplateFile    string
	BuiltinTemplate string
	Data            any
}

// GenFile generates a file from a template.
func GenFile(c FileGenConfig) error {
	fp, created, err := apiutil.MaybeCreateFile(c.Dir, c.Subdir, c.Filename)
	if err != nil {
		return err
	}
	if !created {
		return nil
	}
	defer fp.Close()

	var text string
	if len(c.Category) == 0 || len(c.TemplateFile) == 0 {
		text = c.BuiltinTemplate
	} else {
		text, err = pathx.LoadTemplate(c.Category, c.TemplateFile, c.BuiltinTemplate)
		if err != nil {
			return err
		}
	}

	t := template.Must(template.New(c.TemplateName).Parse(text))
	buffer := new(bytes.Buffer)
	err = t.Execute(buffer, c.Data)
	if err != nil {
		return err
	}

	code := golang.FormatCode(buffer.String())
	_, err = fp.WriteString(code)
	return err
}

// WriteProperty writes a struct field to writer.
func WriteProperty(writer io.Writer, name, tag, comment string, tp spec.Type, indent int) error {
	util.WriteIndent(writer, indent)
	var (
		err            error
		isNestedStruct bool
	)
	structType, ok := tp.(spec.NestedStruct)
	if ok {
		isNestedStruct = true
	}
	if len(comment) > 0 {
		comment = strings.TrimPrefix(comment, "//")
		comment = "//" + comment
	}

	if isNestedStruct {
		_, err = fmt.Fprintf(writer, "%s struct {\n", strings.Title(name))
		if err != nil {
			return err
		}

		if err := writeMember(writer, structType.Members); err != nil {
			return err
		}

		_, err := fmt.Fprintf(writer, "} %s", tag)
		if err != nil {
			return err
		}

		if len(comment) > 0 {
			_, err = fmt.Fprintf(writer, " %s", comment)
			if err != nil {
				return err
			}
		}
		_, err = fmt.Fprint(writer, "\n")
		if err != nil {
			return err
		}
	} else {
		if len(comment) > 0 {
			_, err = fmt.Fprintf(writer, "%s %s %s %s\n", strings.Title(name), tp.Name(), tag, comment)
			if err != nil {
				return err
			}
		} else {
			_, err = fmt.Fprintf(writer, "%s %s %s\n", strings.Title(name), tp.Name(), tag)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func writeMember(writer io.Writer, members []spec.Member) error {
	for _, member := range members {
		if member.IsInline {
			if _, err := fmt.Fprintf(writer, "%s\n", strings.Title(member.Type.Name())); err != nil {
				return err
			}
			continue
		}
		if err := WriteProperty(writer, member.Name, member.Tag, member.GetComment(), member.Type, 1); err != nil {
			return err
		}
	}
	return nil
}
