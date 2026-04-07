package parser

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	"github.com/lerity-yao/czt-contrib/cztctl/api/parser/g4/ast"
	"github.com/lerity-yao/czt-contrib/cztctl/api/parser/g4/gen/cztctl"
	"github.com/lerity-yao/czt-contrib/cztctl/api/spec"
	extParser "github.com/lerity-yao/czt-contrib/cztctl/pkg/parser/extension/parser"
	"github.com/lerity-yao/czt-contrib/cztctl/util/env"
)

// cronRouteSeparators defines legal separators for .cron routeName: - :
var cronRouteSeparators = map[rune]bool{'-': true, ':': true}

// rabbitmqRouteSeparators defines legal separators for .rabbitmq routeName: . -
var rabbitmqRouteSeparators = map[rune]bool{'.': true, '-': true}

type parser struct {
	ast  *ast.Api
	spec *spec.ApiSpec
}

// Parse parses the extension api file (.cron / .rabbitmq).
// Default uses ANTLR4 parser; uses hand-written parser when CZTCTL_EXPERIMENTAL=on.
// Toggle via: cztctl env -w CZTCTL_EXPERIMENTAL=on
func Parse(filename string, src interface{}) (*spec.ApiSpec, error) {
	if env.UseExperimental() {
		return extParser.Parse(filename, src)
	}

	astParser := ast.NewParser(ast.WithParserPrefix(filepath.Base(filename)), ast.WithParserDebug())
	parsedApi, err := astParser.Parse(filename)
	if err != nil {
		return nil, err
	}

	apiSpec := new(spec.ApiSpec)
	p := parser{ast: parsedApi, spec: apiSpec}
	err = p.convert2Spec(filename)
	if err != nil {
		return nil, err
	}

	return apiSpec, nil
}

func (p parser) convert2Spec(filename string) error {
	p.fillInfo()
	p.fillSyntax()
	p.fillImport()
	err := p.fillTypes()
	if err != nil {
		return err
	}
	return p.fillService(filename)
}

func (p parser) fillInfo() {
	properties := make(map[string]string)
	if p.ast.Info != nil {
		for _, kv := range p.ast.Info.Kvs {
			properties[kv.Key.Text()] = kv.Value.Text()
		}
	}
	p.spec.Info.Properties = properties
	p.spec.Info.Title = properties["title"]
	p.spec.Info.Desc = properties["desc"]
	p.spec.Info.Version = properties["version"]
	p.spec.Info.Author = properties["author"]
	p.spec.Info.Email = properties["email"]
}

func (p parser) fillSyntax() {
	if p.ast.Syntax != nil {
		p.spec.Syntax = spec.ApiSyntax{
			Version: p.ast.Syntax.Version.Text(),
			Doc:     p.stringExprs(p.ast.Syntax.DocExpr),
			Comment: p.stringExprs([]ast.Expr{p.ast.Syntax.CommentExpr}),
		}
	}
}

func (p parser) fillImport() {
	if len(p.ast.Import) > 0 {
		for _, item := range p.ast.Import {
			p.spec.Imports = append(p.spec.Imports, spec.Import{
				Value:   item.Value.Text(),
				Doc:     p.stringExprs(item.DocExpr),
				Comment: p.stringExprs([]ast.Expr{item.CommentExpr}),
			})
		}
	}
}

func (p parser) fillTypes() error {
	for _, item := range p.ast.Type {
		switch v := (item).(type) {
		case *ast.TypeStruct:
			members := make([]spec.Member, 0, len(v.Fields))
			for _, item := range v.Fields {
				members = append(members, p.fieldToMember(item))
			}
			p.spec.Types = append(p.spec.Types, spec.DefineStruct{
				RawName: v.Name.Text(),
				Members: members,
				Docs:    p.stringExprs(v.Doc()),
			})
		default:
			return fmt.Errorf("unknown type %+v", v)
		}
	}

	var types []spec.Type
	for _, item := range p.spec.Types {
		switch v := (item).(type) {
		case spec.DefineStruct:
			var members []spec.Member
			for _, member := range v.Members {
				switch v := member.Type.(type) {
				case spec.DefineStruct:
					tp, err := p.findDefinedType(v.RawName)
					if err != nil {
						return err
					}
					member.Type = *tp
				}
				members = append(members, member)
			}
			v.Members = members
			types = append(types, v)
		default:
			return fmt.Errorf("unknown type %+v", v)
		}
	}
	p.spec.Types = types
	return nil
}

func (p parser) findDefinedType(name string) (*spec.Type, error) {
	for _, item := range p.spec.Types {
		if _, ok := item.(spec.DefineStruct); ok {
			if item.Name() == name {
				return &item, nil
			}
		}
	}
	return nil, fmt.Errorf("type %s not defined", name)
}

func (p parser) fieldToMember(field *ast.TypeField) spec.Member {
	var name string
	var tag string
	if !field.IsAnonymous {
		name = field.Name.Text()
		if field.Tag != nil {
			tag = field.Tag.Text()
		}
	}
	return spec.Member{
		Name:     name,
		Type:     p.astTypeToSpec(field.DataType),
		Tag:      tag,
		Comment:  p.commentExprs(field.Comment()),
		Docs:     p.stringExprs(field.Doc()),
		IsInline: field.IsAnonymous,
	}
}

func (p parser) astTypeToSpec(in ast.DataType) spec.Type {
	switch v := (in).(type) {
	case *ast.Literal:
		raw := v.Literal.Text()
		if cztctl.IsBasicType(raw) {
			return spec.PrimitiveType{RawName: raw}
		}
		return spec.DefineStruct{RawName: raw}
	case *ast.Interface:
		return spec.InterfaceType{RawName: v.Literal.Text()}
	case *ast.Map:
		return spec.MapType{RawName: v.MapExpr.Text(), Key: v.Key.Text(), Value: p.astTypeToSpec(v.Value)}
	case *ast.Array:
		return spec.ArrayType{RawName: v.ArrayExpr.Text(), Value: p.astTypeToSpec(v.Literal)}
	case *ast.Pointer:
		raw := v.Name.Text()
		if cztctl.IsBasicType(raw) {
			return spec.PointerType{RawName: v.PointerExpr.Text(), Type: spec.PrimitiveType{RawName: raw}}
		}
		return spec.PointerType{RawName: v.PointerExpr.Text(), Type: spec.DefineStruct{RawName: raw}}
	}
	panic(fmt.Sprintf("unsupported type %+v", in))
}

func (p parser) stringExprs(docs []ast.Expr) []string {
	var result []string
	for _, item := range docs {
		if item == nil {
			continue
		}
		result = append(result, item.Text())
	}
	return result
}

func (p parser) commentExprs(comment ast.Expr) string {
	if comment == nil {
		return ""
	}
	return comment.Text()
}

func (p parser) fillService(filename string) error {
	isCron := strings.HasSuffix(filename, ".cron")

	var groups []spec.Group
	for _, item := range p.ast.Service {
		var group spec.Group
		p.fillAtServer(item, &group)

		for _, astRoute := range item.ServiceApi.ServiceRoute {
			route := spec.Route{
				AtServerAnnotation: spec.Annotation{},
				Method:             astRoute.Route.RouteName.Text(),
				Doc:                p.stringExprs(astRoute.Route.DocExpr),
				Comment:            p.stringExprs([]ast.Expr{astRoute.Route.CommentExpr}),
			}

			if astRoute.AtHandler != nil {
				route.Handler = astRoute.AtHandler.Name.Text()
				route.HandlerDoc = append(route.HandlerDoc, p.stringExprs(astRoute.AtHandler.DocExpr)...)
				route.HandlerComment = append(route.HandlerComment, p.stringExprs([]ast.Expr{astRoute.AtHandler.CommentExpr})...)
			}

			if astRoute.Route.Req != nil {
				route.RequestType = p.astTypeToSpec(astRoute.Route.Req.Name)
			}

			if astRoute.AtDoc != nil {
				properties := make(map[string]string)
				for _, kv := range astRoute.AtDoc.Kv {
					properties[kv.Key.Text()] = kv.Value.Text()
				}
				route.AtDoc.Properties = properties
				if astRoute.AtDoc.LineDoc != nil {
					route.AtDoc.Text = astRoute.AtDoc.LineDoc.Text()
				}
			}

			// Semantic isolation: .cron fills Cron/CronRetry, .rabbitmq fills Queue
			if isCron {
				if astRoute.AtCron != nil {
					route.Cron = strings.Trim(astRoute.AtCron.CronExpr.Text(), `"`)
				}
				if astRoute.AtCronRetry != nil {
					retryStr := astRoute.AtCronRetry.RetryCount.Text()
					retry, err := strconv.Atoi(retryStr)
					if err != nil {
						return fmt.Errorf("invalid @cronRetry value: %s", retryStr)
					}
					route.CronRetry = retry
				}
			} else {
				// .rabbitmq: routeName is the queue name
				route.Queue = astRoute.Route.RouteName.Text()
			}

			// Validate routeName separators by file type
			allowedSeps := rabbitmqRouteSeparators
			fileType := ".rabbitmq"
			if isCron {
				allowedSeps = cronRouteSeparators
				fileType = ".cron"
			}
			for _, ch := range route.Method {
				if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) && ch != '_' && ch != '$' {
					if !allowedSeps[ch] {
						return fmt.Errorf("route [%s] contains illegal separator '%c' for %s file", route.Method, ch, fileType)
					}
				}
			}

			err := p.fillRouteType(&route)
			if err != nil {
				return err
			}

			for _, char := range route.Handler {
				if !unicode.IsDigit(char) && !unicode.IsLetter(char) {
					return fmt.Errorf("route [%s] handler [%s] invalid, handler name should only contains letter or digit",
						route.Method, route.Handler)
				}
			}

			group.Routes = append(group.Routes, route)
			name := item.ServiceApi.Name.Text()
			if len(p.spec.Service.Name) > 0 && p.spec.Service.Name != name {
				return fmt.Errorf("multiple service names defined %s and %s",
					name, p.spec.Service.Name)
			}
			p.spec.Service.Name = name
		}
		groups = append(groups, group)
	}
	p.spec.Service.Groups = groups
	return nil
}

func (p parser) fillAtServer(item *ast.Service, group *spec.Group) {
	if item.AtServer != nil {
		properties := make(map[string]string)
		for _, kv := range item.AtServer.Kv {
			properties[kv.Key.Text()] = kv.Value.Text()
		}
		group.Annotation.Properties = properties
	}
}

func (p parser) fillRouteType(route *spec.Route) error {
	if route.RequestType != nil {
		switch route.RequestType.(type) {
		case spec.DefineStruct:
			tp, err := p.findDefinedType(route.RequestType.Name())
			if err != nil {
				return err
			}
			route.RequestType = *tp
		}
	}
	return nil
}
