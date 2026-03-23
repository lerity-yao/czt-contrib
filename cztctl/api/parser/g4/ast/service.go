package ast

import (
	"fmt"
	"sort"

	"github.com/lerity-yao/czt-contrib/cztctl/api/parser/g4/gen/cztctl"
)

// Service describes service for api syntax
type Service struct {
	AtServer   *AtServer
	ServiceApi *ServiceApi
}

// KV defines a slice for KvExpr
type KV []*KvExpr

// AtServer describes server metadata for api syntax
type AtServer struct {
	AtServerToken Expr
	Lp            Expr
	Rp            Expr
	Kv            KV
}

// ServiceApi describes service ast for api syntax
type ServiceApi struct {
	ServiceToken Expr
	Name         Expr
	Lbrace       Expr
	Rbrace       Expr
	ServiceRoute []*ServiceRoute
}

// ServiceRoute describes service route ast for api syntax
type ServiceRoute struct {
	AtDoc       *AtDoc
	AtCron      *AtCron
	AtCronRetry *AtCronRetry
	AtHandler   *AtHandler
	Route       *Route
}

// AtDoc describes service comments ast for api syntax
type AtDoc struct {
	AtDocToken Expr
	Lp         Expr
	Rp         Expr
	LineDoc    Expr
	Kv         []*KvExpr
}

// AtHandler describes service handler ast for api syntax
type AtHandler struct {
	AtHandlerToken Expr
	Name           Expr
	DocExpr        []Expr
	CommentExpr    Expr
}

// AtCron describes @cron annotation ast
type AtCron struct {
	AtCronToken Expr
	CronExpr    Expr
}

// AtCronRetry describes @cronRetry annotation ast
type AtCronRetry struct {
	AtCronRetryToken Expr
	RetryCount       Expr
}

// Route describes route ast for cron/rabbitmq syntax
type Route struct {
	RouteName   Expr
	Req         *Body
	DocExpr     []Expr
	CommentExpr Expr
}

// Body describes request body ast for api syntax
type Body struct {
	Lp   Expr
	Rp   Expr
	Name DataType
}

// VisitServiceSpec implements from cztctl.BaseCztctlParserVisitor
func (v *ApiVisitor) VisitServiceSpec(ctx *cztctl.ServiceSpecContext) any {
	var serviceSpec Service
	if ctx.AtServer() != nil {
		serviceSpec.AtServer = ctx.AtServer().Accept(v).(*AtServer)
	}
	serviceSpec.ServiceApi = ctx.ServiceApi().Accept(v).(*ServiceApi)
	return &serviceSpec
}

// VisitAtServer implements from cztctl.BaseCztctlParserVisitor
func (v *ApiVisitor) VisitAtServer(ctx *cztctl.AtServerContext) any {
	var atServer AtServer
	atServer.AtServerToken = v.newExprWithTerminalNode(ctx.ATSERVER())
	atServer.Lp = v.newExprWithToken(ctx.GetLp())
	atServer.Rp = v.newExprWithToken(ctx.GetRp())
	for _, each := range ctx.AllKvLit() {
		atServer.Kv = append(atServer.Kv, each.Accept(v).(*KvExpr))
	}
	return &atServer
}

// VisitServiceApi implements from cztctl.BaseCztctlParserVisitor
func (v *ApiVisitor) VisitServiceApi(ctx *cztctl.ServiceApiContext) any {
	var serviceApi ServiceApi
	serviceApi.ServiceToken = v.newExprWithToken(ctx.GetServiceToken())
	serviceName := ctx.ServiceName()
	serviceApi.Name = v.newExprWithText(serviceName.GetText(),
		serviceName.GetStart().GetLine(), serviceName.GetStart().GetColumn(),
		serviceName.GetStart().GetStart(), serviceName.GetStop().GetStop())
	serviceApi.Lbrace = v.newExprWithToken(ctx.GetLbrace())
	serviceApi.Rbrace = v.newExprWithToken(ctx.GetRbrace())
	for _, each := range ctx.AllServiceRoute() {
		serviceApi.ServiceRoute = append(serviceApi.ServiceRoute, each.Accept(v).(*ServiceRoute))
	}
	return &serviceApi
}

// VisitServiceRoute implements from cztctl.BaseCztctlParserVisitor
func (v *ApiVisitor) VisitServiceRoute(ctx *cztctl.ServiceRouteContext) any {
	var serviceRoute ServiceRoute
	if ctx.AtDoc() != nil {
		serviceRoute.AtDoc = ctx.AtDoc().Accept(v).(*AtDoc)
	}
	if ctx.AtCron() != nil {
		serviceRoute.AtCron = ctx.AtCron().Accept(v).(*AtCron)
	}
	if ctx.AtCronRetry() != nil {
		serviceRoute.AtCronRetry = ctx.AtCronRetry().Accept(v).(*AtCronRetry)
	}
	serviceRoute.AtHandler = ctx.AtHandler().Accept(v).(*AtHandler)
	serviceRoute.Route = ctx.Route().Accept(v).(*Route)
	return &serviceRoute
}

// VisitAtDoc implements from cztctl.BaseCztctlParserVisitor
func (v *ApiVisitor) VisitAtDoc(ctx *cztctl.AtDocContext) any {
	var atDoc AtDoc
	atDoc.AtDocToken = v.newExprWithTerminalNode(ctx.ATDOC())
	if ctx.STRING() != nil {
		atDoc.LineDoc = v.newExprWithTerminalNode(ctx.STRING())
	} else {
		for _, each := range ctx.AllKvLit() {
			atDoc.Kv = append(atDoc.Kv, each.Accept(v).(*KvExpr))
		}
	}
	atDoc.Lp = v.newExprWithToken(ctx.GetLp())
	atDoc.Rp = v.newExprWithToken(ctx.GetRp())
	if ctx.GetLp() != nil {
		if ctx.GetRp() == nil {
			v.panic(atDoc.Lp, "mismatched ')'")
		}
	}
	if ctx.GetRp() != nil {
		if ctx.GetLp() == nil {
			v.panic(atDoc.Rp, "mismatched '('")
		}
	}
	return &atDoc
}

// VisitAtHandler implements from cztctl.BaseCztctlParserVisitor
func (v *ApiVisitor) VisitAtHandler(ctx *cztctl.AtHandlerContext) any {
	var atHandler AtHandler
	atHandler.AtHandlerToken = v.newExprWithTerminalNode(ctx.ATHANDLER())
	atHandler.Name = v.newExprWithTerminalNode(ctx.ID())
	atHandler.DocExpr = v.getDoc(ctx)
	atHandler.CommentExpr = v.getComment(ctx)
	return &atHandler
}

// VisitAtCron implements from cztctl.BaseCztctlParserVisitor
func (v *ApiVisitor) VisitAtCron(ctx *cztctl.AtCronContext) any {
	return &AtCron{
		AtCronToken: v.newExprWithTerminalNode(ctx.ATCRON()),
		CronExpr:    v.newExprWithTerminalNode(ctx.STRING()),
	}
}

// VisitAtCronRetry implements from cztctl.BaseCztctlParserVisitor
func (v *ApiVisitor) VisitAtCronRetry(ctx *cztctl.AtCronRetryContext) any {
	return &AtCronRetry{
		AtCronRetryToken: v.newExprWithTerminalNode(ctx.ATCRONRETRY()),
		RetryCount:       v.newExprWithTerminalNode(ctx.INT()),
	}
}

// VisitRoute implements from cztctl.BaseCztctlParserVisitor
func (v *ApiVisitor) VisitRoute(ctx *cztctl.RouteContext) any {
	var route Route
	routeName := ctx.RouteName()
	route.RouteName = v.newExprWithText(routeName.GetText(),
		routeName.GetStart().GetLine(), routeName.GetStart().GetColumn(),
		routeName.GetStart().GetStart(), routeName.GetStop().GetStop())
	if ctx.GetRequest() != nil {
		req := ctx.GetRequest().Accept(v)
		if req != nil {
			route.Req = req.(*Body)
		}
	}
	route.DocExpr = v.getDoc(ctx)
	route.CommentExpr = v.getComment(ctx)
	return &route
}

// VisitRouteName implements from cztctl.BaseCztctlParserVisitor
func (v *ApiVisitor) VisitRouteName(ctx *cztctl.RouteNameContext) any {
	return v.newExprWithText(ctx.GetText(),
		ctx.GetStart().GetLine(), ctx.GetStart().GetColumn(),
		ctx.GetStart().GetStart(), ctx.GetStop().GetStop())
}

// VisitBody implements from cztctl.BaseCztctlParserVisitor
func (v *ApiVisitor) VisitBody(ctx *cztctl.BodyContext) any {
	if ctx.ID() == nil {
		if v.debug {
			fmt.Printf("%s line %d:  expr \"()\" is deprecated, if there has no request body, please omit it\n",
				v.prefix, ctx.GetStart().GetLine())
		}
		return nil
	}
	idExpr := v.newExprWithTerminalNode(ctx.ID())
	if cztctl.IsGolangKeyWord(idExpr.Text()) {
		v.panic(idExpr, fmt.Sprintf("expecting 'ID', but found golang keyword '%s'", idExpr.Text()))
	}
	return &Body{
		Lp:   v.newExprWithToken(ctx.GetLp()),
		Rp:   v.newExprWithToken(ctx.GetRp()),
		Name: &Literal{Literal: idExpr},
	}
}

// VisitKvValue implements from cztctl.BaseCztctlParserVisitor
func (v *ApiVisitor) VisitKvValue(ctx *cztctl.KvValueContext) any {
	return v.newExprWithText(ctx.GetText(),
		ctx.GetStart().GetLine(), ctx.GetStart().GetColumn(),
		ctx.GetStart().GetStart(), ctx.GetStop().GetStop())
}

// --- Format / Equal / Doc / Comment methods ---

func (b *Body) Format() error { return nil }
func (b *Body) Equal(v any) bool {
	if v == nil {
		return false
	}
	body, ok := v.(*Body)
	if !ok {
		return false
	}
	if !b.Lp.Equal(body.Lp) {
		return false
	}
	if !b.Rp.Equal(body.Rp) {
		return false
	}
	return b.Name.Equal(body.Name)
}

func (r *Route) Format() error { return nil }
func (r *Route) Doc() []Expr   { return r.DocExpr }
func (r *Route) Comment() Expr { return r.CommentExpr }
func (r *Route) Equal(v any) bool {
	if v == nil {
		return false
	}
	route, ok := v.(*Route)
	if !ok {
		return false
	}
	if !r.RouteName.Equal(route.RouteName) {
		return false
	}
	if r.Req != nil {
		if !r.Req.Equal(route.Req) {
			return false
		}
	}
	return EqualDoc(r, route)
}

func (a *AtHandler) Doc() []Expr   { return a.DocExpr }
func (a *AtHandler) Comment() Expr { return a.CommentExpr }
func (a *AtHandler) Format() error { return nil }
func (a *AtHandler) Equal(v any) bool {
	if v == nil {
		return false
	}
	h, ok := v.(*AtHandler)
	if !ok {
		return false
	}
	if !a.AtHandlerToken.Equal(h.AtHandlerToken) {
		return false
	}
	if !a.Name.Equal(h.Name) {
		return false
	}
	return EqualDoc(a, h)
}

func (a *AtDoc) Format() error { return nil }
func (a *AtDoc) Equal(v any) bool {
	if v == nil {
		return false
	}
	atDoc, ok := v.(*AtDoc)
	if !ok {
		return false
	}
	if !a.AtDocToken.Equal(atDoc.AtDocToken) {
		return false
	}
	if a.LineDoc != nil {
		if !a.LineDoc.Equal(atDoc.LineDoc) {
			return false
		}
	}
	var expecting, actual []*KvExpr
	expecting = append(expecting, a.Kv...)
	actual = append(actual, atDoc.Kv...)
	if len(expecting) != len(actual) {
		return false
	}
	for index, each := range expecting {
		ac := actual[index]
		if !each.Equal(ac) {
			return false
		}
	}
	return true
}

func (a *AtServer) Format() error { return nil }
func (a *AtServer) Equal(v any) bool {
	if v == nil {
		return false
	}
	atServer, ok := v.(*AtServer)
	if !ok {
		return false
	}
	if !a.AtServerToken.Equal(atServer.AtServerToken) {
		return false
	}
	if !a.Lp.Equal(atServer.Lp) {
		return false
	}
	if !a.Rp.Equal(atServer.Rp) {
		return false
	}
	var expecting, actual []*KvExpr
	expecting = append(expecting, a.Kv...)
	actual = append(actual, atServer.Kv...)
	if len(expecting) != len(actual) {
		return false
	}
	sort.Slice(expecting, func(i, j int) bool { return expecting[i].Key.Text() < expecting[j].Key.Text() })
	sort.Slice(actual, func(i, j int) bool { return actual[i].Key.Text() < actual[j].Key.Text() })
	for index, each := range expecting {
		ac := actual[index]
		if !each.Equal(ac) {
			return false
		}
	}
	return true
}

func (s *ServiceRoute) Equal(v any) bool {
	if v == nil {
		return false
	}
	sr, ok := v.(*ServiceRoute)
	if !ok {
		return false
	}
	if !s.AtDoc.Equal(sr.AtDoc) {
		return false
	}
	if s.AtHandler != nil {
		if !s.AtHandler.Equal(sr.AtHandler) {
			return false
		}
	}
	return s.Route.Equal(sr.Route)
}
func (s *ServiceRoute) Format() error { return nil }

// GetHandler returns handler name of api route
func (s *ServiceRoute) GetHandler() Expr {
	if s.AtHandler != nil {
		return s.AtHandler.Name
	}
	return nil
}

func (a *ServiceApi) Format() error { return nil }
func (a *ServiceApi) Equal(v any) bool {
	if v == nil {
		return false
	}
	api, ok := v.(*ServiceApi)
	if !ok {
		return false
	}
	if !a.ServiceToken.Equal(api.ServiceToken) {
		return false
	}
	if !a.Name.Equal(api.Name) {
		return false
	}
	if !a.Lbrace.Equal(api.Lbrace) {
		return false
	}
	if !a.Rbrace.Equal(api.Rbrace) {
		return false
	}
	var expecting, actual []*ServiceRoute
	expecting = append(expecting, a.ServiceRoute...)
	actual = append(actual, api.ServiceRoute...)
	if len(expecting) != len(actual) {
		return false
	}
	sort.Slice(expecting, func(i, j int) bool { return expecting[i].Route.RouteName.Text() < expecting[j].Route.RouteName.Text() })
	sort.Slice(actual, func(i, j int) bool { return actual[i].Route.RouteName.Text() < actual[j].Route.RouteName.Text() })
	for index, each := range expecting {
		ac := actual[index]
		if !each.Equal(ac) {
			return false
		}
	}
	return true
}

func (s *Service) Format() error { return nil }
func (s *Service) Equal(v any) bool {
	if v == nil {
		return false
	}
	service, ok := v.(*Service)
	if !ok {
		return false
	}
	if s.AtServer != nil {
		if !s.AtServer.Equal(service.AtServer) {
			return false
		}
	}
	return s.ServiceApi.Equal(service.ServiceApi)
}

// Get returns the target KV by specified key
func (kv KV) Get(key string) Expr {
	for _, each := range kv {
		if each.Key.Text() == key {
			return each.Value
		}
	}
	return nil
}
