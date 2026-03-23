package parser

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/lerity-yao/czt-contrib/cztctl/api/spec"
	"github.com/lerity-yao/czt-contrib/cztctl/pkg/parser/extension/ast"
	"github.com/lerity-yao/czt-contrib/cztctl/pkg/parser/extension/token"
)

const (
	atServerGroupKey = "group"
)

// API is the parsed extension file.
type API struct {
	Filename     string
	Syntax       *ast.SyntaxStmt
	info         *ast.InfoStmt
	importStmt   []ast.ImportStmt
	TypeStmt     []ast.TypeStmt
	ServiceStmts []*ast.ServiceStmt
	importSet    map[string]struct{}
}

// Analyzer converts AST to spec.
type Analyzer struct {
	api  *API
	spec *spec.ApiSpec
}

// Parse parses the given extension file and returns the parsed spec.
// mode is determined by file extension: .cron → ParseModeCron, .rabbitmq → ParseModeRabbitMQ.
func Parse(filename string, src interface{}) (*spec.ApiSpec, error) {
	mode := detectMode(filename)
	p := New(filename, src, mode)
	tree := p.Parse()
	if err := p.CheckErrors(); err != nil {
		return nil, err
	}

	importSet := map[string]struct{}{}
	api, err := convert2API(tree, importSet)
	if err != nil {
		return nil, err
	}

	if err = api.parseReverse(mode); err != nil {
		return nil, err
	}

	if err = api.selfCheck(); err != nil {
		return nil, err
	}

	result := new(spec.ApiSpec)
	analyzer := Analyzer{api: api, spec: result}

	if err = analyzer.convert2Spec(mode); err != nil {
		return nil, err
	}

	return result, nil
}

func detectMode(filename string) ParseMode {
	if strings.HasSuffix(filename, ".rabbitmq") {
		return ParseModeRabbitMQ
	}
	return ParseModeCron
}

func convert2API(a *ast.AST, importSet map[string]struct{}) (*API, error) {
	api := &API{
		Filename:  a.Filename,
		importSet: importSet,
	}

	if len(a.Stmts) == 0 {
		return api, nil
	}

	one := a.Stmts[0]
	syntax, ok := one.(*ast.SyntaxStmt)
	if !ok {
		syntax = &ast.SyntaxStmt{
			Syntax: ast.NewTokenNode(token.Token{Type: token.IDENT, Text: token.Syntax}),
			Assign: ast.NewTokenNode(token.Token{Type: token.ASSIGN, Text: "="}),
			Value:  ast.NewTokenNode(token.Token{Type: token.STRING, Text: `"v1"`}),
		}
	}
	api.Syntax = syntax

	var hasSyntax, hasInfo bool
	for i := 0; i < len(a.Stmts); i++ {
		one := a.Stmts[i]
		switch val := one.(type) {
		case *ast.SyntaxStmt:
			if hasSyntax {
				return nil, ast.DuplicateStmtError(val.Pos(), "duplicate syntax statement")
			}
			hasSyntax = true
		case *ast.InfoStmt:
			if hasInfo {
				return nil, ast.DuplicateStmtError(val.Pos(), "duplicate info statement")
			}
			hasInfo = true
			api.info = val
		case ast.ImportStmt:
			api.importStmt = append(api.importStmt, val)
		case ast.TypeStmt:
			api.TypeStmt = append(api.TypeStmt, val)
		case *ast.ServiceStmt:
			api.ServiceStmts = append(api.ServiceStmts, val)
		}
	}

	return api, nil
}

func (api *API) mergeAPI(in *API) error {
	api.TypeStmt = append(api.TypeStmt, in.TypeStmt...)
	api.ServiceStmts = append(api.ServiceStmts, in.ServiceStmts...)
	return nil
}

func (api *API) parseReverse(mode ParseMode) error {
	list, err := api.parseImportedAPI(api.importStmt, mode)
	if err != nil {
		return err
	}
	for _, e := range list {
		if err = api.mergeAPI(e); err != nil {
			return err
		}
	}
	return nil
}

func (api *API) parseImportedAPI(imports []ast.ImportStmt, mode ParseMode) ([]*API, error) {
	var list []*API
	if len(imports) == 0 {
		return list, nil
	}

	var importValueSet = map[string]token.Token{}
	for _, imp := range imports {
		switch val := imp.(type) {
		case *ast.ImportLiteralStmt:
			importValueSet[strings.ReplaceAll(val.Value.Token.Text, `"`, "")] = val.Value.Token
		case *ast.ImportGroupStmt:
			for _, v := range val.Values {
				importValueSet[strings.ReplaceAll(v.Token.Text, `"`, "")] = v.Token
			}
		}
	}

	dir := filepath.Dir(api.Filename)
	for impPath, tok := range importValueSet {
		if !filepath.IsAbs(impPath) {
			impPath = filepath.Join(dir, impPath)
		}

		if _, ok := api.importSet[impPath]; ok {
			continue
		}
		api.importSet[impPath] = struct{}{}

		p := New(impPath, "", mode)
		tree := p.Parse()
		if err := p.CheckErrors(); err != nil {
			return nil, fmt.Errorf("%s: %w", tok.Position.String(), err)
		}

		nestedAPI, err := convert2API(tree, api.importSet)
		if err != nil {
			return nil, err
		}

		if err = nestedAPI.parseReverse(mode); err != nil {
			return nil, err
		}

		list = append(list, nestedAPI)
	}

	return list, nil
}

func (api *API) selfCheck() error {
	if err := api.checkTypeStmt(); err != nil {
		return err
	}
	if err := api.checkServiceStmt(); err != nil {
		return err
	}
	return api.checkTypeDeclareContext()
}

func (api *API) checkTypeStmt() error {
	seen := map[string]token.Position{}
	for _, v := range api.TypeStmt {
		switch val := v.(type) {
		case *ast.TypeLiteralStmt:
			name := val.Expr.Name.Token.Text
			if pos, ok := seen[name]; ok && pos != val.Expr.Name.Token.Position {
				return ast.DuplicateStmtError(val.Expr.Name.Pos(), "duplicate type expression")
			}
			seen[name] = val.Expr.Name.Token.Position
		case *ast.TypeGroupStmt:
			for _, expr := range val.ExprList {
				name := expr.Name.Token.Text
				if pos, ok := seen[name]; ok && pos != expr.Name.Token.Position {
					return ast.DuplicateStmtError(expr.Name.Pos(), "duplicate type expression")
				}
				seen[name] = expr.Name.Token.Position
			}
		}
	}
	return nil
}

func (api *API) checkServiceStmt() error {
	handlerSeen := map[string]token.Position{}
	for _, v := range api.ServiceStmts {
		for _, item := range v.Routes {
			if item.AtHandler == nil {
				continue
			}
			name := item.AtHandler.Name.Token.Text
			if pos, ok := handlerSeen[name]; ok && pos != item.AtHandler.Name.Token.Position {
				return ast.DuplicateStmtError(item.AtHandler.Name.Pos(), "duplicate handler expression")
			}
			handlerSeen[name] = item.AtHandler.Name.Token.Position
		}
	}
	return nil
}

func (api *API) checkTypeDeclareContext() error {
	typeMap := map[string]struct{}{}
	for _, v := range api.TypeStmt {
		switch tp := v.(type) {
		case *ast.TypeLiteralStmt:
			typeMap[tp.Expr.Name.Token.Text] = struct{}{}
		case *ast.TypeGroupStmt:
			for _, v := range tp.ExprList {
				typeMap[v.Name.Token.Text] = struct{}{}
			}
		}
	}
	return api.checkTypeContext(typeMap)
}

func (api *API) checkTypeContext(declareContext map[string]struct{}) error {
	em := newErrorManager()
	for _, v := range api.TypeStmt {
		switch tp := v.(type) {
		case *ast.TypeLiteralStmt:
			em.add(api.checkTypeExprContext(declareContext, tp.Expr.DataType))
		case *ast.TypeGroupStmt:
			for _, v := range tp.ExprList {
				em.add(api.checkTypeExprContext(declareContext, v.DataType))
			}
		}
	}
	return em.error()
}

func (api *API) checkTypeExprContext(declareContext map[string]struct{}, tp ast.DataType) error {
	switch val := tp.(type) {
	case *ast.ArrayDataType:
		return api.checkTypeExprContext(declareContext, val.DataType)
	case *ast.BaseDataType:
		if token.IsBaseType(val.Base.Token.Text) {
			return nil
		}
		if _, ok := declareContext[val.Base.Token.Text]; !ok {
			return ast.SyntaxError(val.Base.Pos(), "unresolved type <%s>", val.Base.Token.Text)
		}
		return nil
	case *ast.MapDataType:
		manager := newErrorManager()
		manager.add(api.checkTypeExprContext(declareContext, val.Key))
		manager.add(api.checkTypeExprContext(declareContext, val.Value))
		return manager.error()
	case *ast.PointerDataType:
		return api.checkTypeExprContext(declareContext, val.DataType)
	case *ast.SliceDataType:
		return api.checkTypeExprContext(declareContext, val.DataType)
	case *ast.StructDataType:
		manager := newErrorManager()
		for _, e := range val.Elements {
			manager.add(api.checkTypeExprContext(declareContext, e.DataType))
		}
		return manager.error()
	}
	return nil
}

// ==================== Analyzer: AST → Spec ====================

func (a *Analyzer) convert2Spec(mode ParseMode) error {
	a.fillInfo()

	if err := a.fillTypes(); err != nil {
		return err
	}

	if err := a.fillService(mode); err != nil {
		return err
	}

	sort.SliceStable(a.spec.Types, func(i, j int) bool {
		return a.spec.Types[i].Name() < a.spec.Types[j].Name()
	})

	return nil
}

func (a *Analyzer) fillInfo() {
	properties := make(map[string]string)
	if a.api.info != nil {
		for _, kv := range a.api.info.Values {
			key := kv.Key.Token.Text
			properties[strings.TrimSuffix(key, ":")] = kv.Value.RawText()
		}
	}
	a.spec.Info.Properties = properties
	infoKeyValue := make(map[string]string)
	for key, value := range properties {
		titleKey := strings.Title(strings.TrimSuffix(key, ":"))
		infoKeyValue[titleKey] = value
	}
	a.spec.Info.Title = infoKeyValue[infoTitleKey]
	a.spec.Info.Desc = infoKeyValue[infoDescKey]
	a.spec.Info.Version = infoKeyValue[infoVersionKey]
	a.spec.Info.Author = infoKeyValue[infoAuthorKey]
	a.spec.Info.Email = infoKeyValue[infoEmailKey]
}

func (a *Analyzer) fillTypes() error {
	for _, item := range a.api.TypeStmt {
		switch v := item.(type) {
		case *ast.TypeLiteralStmt:
			if err := a.fillTypeExpr(v.Expr); err != nil {
				return err
			}
		case *ast.TypeGroupStmt:
			for _, expr := range v.ExprList {
				if err := a.fillTypeExpr(expr); err != nil {
					return err
				}
			}
		}
	}

	var types []spec.Type
	for _, item := range a.spec.Types {
		switch v := item.(type) {
		case spec.DefineStruct:
			var members []spec.Member
			for _, member := range v.Members {
				switch mv := member.Type.(type) {
				case spec.DefineStruct:
					tp, err := a.findDefinedType(mv.RawName)
					if err != nil {
						return err
					}
					member.Type = tp
				}
				members = append(members, member)
			}
			v.Members = members
			types = append(types, v)
		default:
			return fmt.Errorf("unknown type %+v", v)
		}
	}
	a.spec.Types = types
	return nil
}

func (a *Analyzer) fillTypeExpr(expr *ast.TypeExpr) error {
	head, _ := expr.CommentGroup()
	switch val := expr.DataType.(type) {
	case *ast.StructDataType:
		var members []spec.Member
		for _, item := range val.Elements {
			m, err := a.fieldToMember(item)
			if err != nil {
				return err
			}
			members = append(members, m)
		}
		a.spec.Types = append(a.spec.Types, spec.DefineStruct{
			RawName: expr.Name.Token.Text,
			Members: members,
			Docs:    head.List(),
		})
		return nil
	default:
		return ast.SyntaxError(expr.Pos(), "expected <struct> expr, got <%T>", expr.DataType)
	}
}

func (a *Analyzer) fillService(mode ParseMode) error {
	var groups []spec.Group
	for _, item := range a.api.ServiceStmts {
		var group spec.Group
		if item.AtServerStmt != nil {
			group.Annotation.Properties = a.convertKV(item.AtServerStmt.Values)
		}

		for _, astRoute := range item.Routes {
			head, leading := astRoute.CommentGroup()
			route := spec.Route{
				Method:  astRoute.Route.Name.Token.Text,
				Doc:     head.List(),
				Comment: leading.List(),
			}

			if astRoute.AtDoc != nil {
				route.AtDoc = a.convertAtDoc(astRoute.AtDoc)
			}
			if astRoute.AtHandler != nil {
				route.Handler = astRoute.AtHandler.Name.Token.Text
				handlerHead, handlerLeading := astRoute.AtHandler.CommentGroup()
				route.HandlerDoc = handlerHead.List()
				route.HandlerComment = handlerLeading.List()
			}

			// cron extensions
			if astRoute.AtCron != nil {
				cronText := astRoute.AtCron.Value.Token.Text
				route.Cron = strings.Trim(cronText, `"`)
			}
			if astRoute.AtCronRetry != nil {
				retryText := astRoute.AtCronRetry.Value.Token.Text
				if v, err := strconv.Atoi(retryText); err == nil {
					route.CronRetry = v
				}
			}

			// rabbitmq extension: queue name is in Method
			if mode == ParseModeRabbitMQ {
				route.Queue = route.Method
			}

			// request param type
			if astRoute.Route.Request != nil && astRoute.Route.Request.Body != nil {
				requestType, err := a.getType(astRoute.Route.Request)
				if err != nil {
					return err
				}
				route.RequestType = requestType
			}

			if err := a.fillRouteType(&route); err != nil {
				return err
			}

			group.Routes = append(group.Routes, route)

			name := item.Name.Format("")
			if len(a.spec.Service.Name) > 0 && a.spec.Service.Name != name {
				return ast.SyntaxError(item.Name.Pos(), "multiple service names defined <%s> and <%s>", name, a.spec.Service.Name)
			}
			a.spec.Service.Name = name
		}
		groups = append(groups, group)
	}

	a.spec.Service.Groups = groups
	return nil
}

func (a *Analyzer) convertAtDoc(atDoc ast.AtDocStmt) spec.AtDoc {
	var ret spec.AtDoc
	switch val := atDoc.(type) {
	case *ast.AtDocLiteralStmt:
		ret.Text = val.Value.Token.Text
	case *ast.AtDocGroupStmt:
		ret.Properties = a.convertKV(val.Values)
	}
	return ret
}

func (a *Analyzer) convertKV(kv []*ast.KVExpr) map[string]string {
	ret := map[string]string{}
	for _, v := range kv {
		key := strings.TrimSuffix(v.Key.Token.Text, ":")
		ret[key] = v.Value.RawText()
	}
	return ret
}

func (a *Analyzer) fieldToMember(field *ast.ElemExpr) (spec.Member, error) {
	var name []string
	for _, v := range field.Name {
		name = append(name, v.Token.Text)
	}

	tp, err := a.astTypeToSpec(field.DataType)
	if err != nil {
		return spec.Member{}, err
	}

	head, leading := field.CommentGroup()
	m := spec.Member{
		Name:     strings.Join(name, ", "),
		Type:     tp,
		Docs:     head.List(),
		Comment:  leading.String(),
		IsInline: field.IsAnonymous(),
	}
	if field.Tag != nil {
		m.Tag = field.Tag.Token.Text
	}

	return m, nil
}

func (a *Analyzer) astTypeToSpec(in ast.DataType) (spec.Type, error) {
	switch v := in.(type) {
	case *ast.BaseDataType:
		raw := v.RawText()
		if token.IsBaseType(raw) {
			return spec.PrimitiveType{RawName: raw}, nil
		}
		return spec.DefineStruct{RawName: raw}, nil
	case *ast.AnyDataType:
		return nil, ast.SyntaxError(v.Pos(), "unsupported any type")
	case *ast.StructDataType:
		var members []spec.Member
		for _, item := range v.Elements {
			m, err := a.fieldToMember(item)
			if err != nil {
				return nil, err
			}
			members = append(members, m)
		}
		if v.RawText() == "{}" {
			return nil, ast.SyntaxError(v.Pos(), "unsupported empty struct")
		}
		return spec.NestedStruct{
			RawName: v.RawText(),
			Members: members,
		}, nil
	case *ast.InterfaceDataType:
		return spec.InterfaceType{RawName: v.RawText()}, nil
	case *ast.MapDataType:
		if !v.Key.CanEqual() {
			return nil, ast.SyntaxError(v.Pos(), "map key must be equal data type")
		}
		if v.Value.ContainsStruct() {
			return nil, ast.SyntaxError(v.Pos(), "map value unsupported nested struct")
		}
		value, err := a.astTypeToSpec(v.Value)
		if err != nil {
			return nil, err
		}
		return spec.MapType{
			RawName: v.RawText(),
			Key:     v.Key.RawText(),
			Value:   value,
		}, nil
	case *ast.PointerDataType:
		raw := v.DataType.RawText()
		if token.IsBaseType(raw) {
			return spec.PointerType{RawName: v.RawText(), Type: spec.PrimitiveType{RawName: raw}}, nil
		}
		value, err := a.astTypeToSpec(v.DataType)
		if err != nil {
			return nil, err
		}
		return spec.PointerType{RawName: v.RawText(), Type: value}, nil
	case *ast.ArrayDataType:
		if v.Length.Token.Type == token.ELLIPSIS {
			return nil, ast.SyntaxError(v.Pos(), "array length unsupported dynamic length")
		}
		if v.ContainsStruct() {
			return nil, ast.SyntaxError(v.Pos(), "array elements unsupported nested struct")
		}
		value, err := a.astTypeToSpec(v.DataType)
		if err != nil {
			return nil, err
		}
		return spec.ArrayType{RawName: v.RawText(), Value: value}, nil
	case *ast.SliceDataType:
		if v.ContainsStruct() {
			return nil, ast.SyntaxError(v.Pos(), "slice elements unsupported nested struct")
		}
		value, err := a.astTypeToSpec(v.DataType)
		if err != nil {
			return nil, err
		}
		return spec.ArrayType{RawName: v.RawText(), Value: value}, nil
	}
	return nil, ast.SyntaxError(in.Pos(), "unsupported type <%T>", in)
}

func (a *Analyzer) fillRouteType(route *spec.Route) error {
	if route.RequestType != nil {
		switch route.RequestType.(type) {
		case spec.DefineStruct:
			tp, err := a.findDefinedType(route.RequestType.Name())
			if err != nil {
				return err
			}
			route.RequestType = tp
		}
	}
	return nil
}

func (a *Analyzer) findDefinedType(name string) (spec.Type, error) {
	for _, item := range a.spec.Types {
		if _, ok := item.(spec.DefineStruct); ok {
			if item.Name() == name {
				return item, nil
			}
		}
	}
	return nil, errors.New("type " + name + " not defined")
}

func (a *Analyzer) getType(expr *ast.BodyStmt) (spec.Type, error) {
	body := expr.Body
	var tp spec.Type
	var err error
	var rawText = body.Format("")

	if token.IsBaseType(body.Value.Token.Text) {
		tp = spec.PrimitiveType{RawName: body.Value.Token.Text}
	} else {
		tp, err = a.findDefinedType(body.Value.Token.Text)
		if err != nil {
			return nil, err
		}
	}

	if body.LBrack != nil {
		if body.Star != nil {
			return spec.ArrayType{
				RawName: rawText,
				Value:   spec.PointerType{RawName: rawText, Type: tp},
			}, nil
		}
		return spec.ArrayType{RawName: rawText, Value: tp}, nil
	}
	if body.Star != nil {
		return spec.PointerType{RawName: rawText, Type: tp}, nil
	}
	return tp, nil
}
