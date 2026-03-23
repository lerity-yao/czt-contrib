package parser

import (
	"fmt"
	"log"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/lerity-yao/czt-contrib/cztctl/pkg/parser/extension/ast"
	"github.com/lerity-yao/czt-contrib/cztctl/pkg/parser/extension/scanner"
	"github.com/lerity-yao/czt-contrib/cztctl/pkg/parser/extension/token"
)

const (
	groupKeyText   = "group"
	infoTitleKey   = "Title"
	infoDescKey    = "Desc"
	infoVersionKey = "Version"
	infoAuthorKey  = "Author"
	infoEmailKey   = "Email"
)

// ParseMode controls cron vs rabbitmq specific parsing.
type ParseMode int

const (
	ParseModeCron     ParseMode = iota // allows @cron, @cronRetry; route = TaskType[(Param)]
	ParseModeRabbitMQ                  // route = queue.name (dot-separated)
)

// Parser is the parser for extension files (.cron / .rabbitmq).
type Parser struct {
	s      *scanner.Scanner
	errors []error
	mode   ParseMode

	curTok  token.Token
	peekTok token.Token

	headCommentGroup ast.CommentGroup
	api              *ast.AST
	node             map[token.Token]*ast.TokenNode
}

// New creates a new parser.
func New(filename string, src interface{}, mode ParseMode) *Parser {
	abs, err := filepath.Abs(filename)
	if err != nil {
		log.Fatalln(err)
	}

	p := &Parser{
		s:    scanner.MustNewScanner(abs, src),
		api:  &ast.AST{Filename: abs},
		node: make(map[token.Token]*ast.TokenNode),
		mode: mode,
	}

	return p
}

// Parse parses the extension file.
func (p *Parser) Parse() *ast.AST {
	if !p.init() {
		return nil
	}

	for p.curTokenIsNotEof() {
		stmt := p.parseStmt()
		if isNil(stmt) {
			return nil
		}
		p.appendStmt(stmt)
		if !p.nextToken() {
			return nil
		}
	}

	return p.api
}

func (p *Parser) parseStmt() ast.Stmt {
	switch p.curTok.Type {
	case token.IDENT:
		switch {
		case p.curTok.Is(token.Syntax):
			return p.parseSyntaxStmt()
		case p.curTok.Is(token.Info):
			return p.parseInfoStmt()
		case p.curTok.Is(token.Service):
			return p.parseService()
		case p.curTok.Is(token.TypeKeyword):
			return p.parseTypeStmt()
		case p.curTok.Is(token.ImportKeyword):
			return p.parseImportStmt()
		default:
			p.errors = append(p.errors, fmt.Errorf("%s syntax error: expected 'syntax' | 'info' | 'service' | 'type' | 'import', got '%s'", p.curTok.Position.String(), p.curTok.Text))
			return nil
		}
	case token.AT_SERVER:
		return p.parseService()
	default:
		p.errors = append(p.errors, fmt.Errorf("%s unexpected token '%s'", p.curTok.Position.String(), p.curTok.Text))
		return nil
	}
}

// ==================== Service Parsing ====================

func (p *Parser) parseService() *ast.ServiceStmt {
	var stmt = &ast.ServiceStmt{}
	if p.curTokenIs(token.AT_SERVER) {
		atServerStmt := p.parseAtServerStmt()
		if atServerStmt == nil {
			return nil
		}
		stmt.AtServerStmt = atServerStmt
		if !p.advanceIfPeekTokenIs(token.Service) {
			return nil
		}
	}
	stmt.Service = p.curTokenNode()

	if !p.advanceIfPeekTokenIs(token.IDENT) {
		return nil
	}

	nameExpr := p.parseServiceNameExpr()
	if nameExpr == nil {
		return nil
	}
	stmt.Name = nameExpr

	if !p.advanceIfPeekTokenIs(token.LBRACE) {
		return nil
	}
	stmt.LBrace = p.curTokenNode()

	routes := p.parseServiceItemsStmt()
	if routes == nil {
		return nil
	}
	stmt.Routes = routes

	if !p.advanceIfPeekTokenIs(token.RBRACE) {
		return nil
	}
	stmt.RBrace = p.curTokenNode()

	return stmt
}

func (p *Parser) parseServiceItemsStmt() []*ast.ServiceItemStmt {
	var stmt = make([]*ast.ServiceItemStmt, 0)
	expectedTokens := p.serviceItemStartTokens()
	expectedTokens = append(expectedTokens, token.RBRACE)

	for p.curTokenIsNotEof() && p.peekTokenIsNot(token.RBRACE) {
		item := p.parseServiceItemStmt()
		if item == nil {
			return nil
		}
		stmt = append(stmt, item)
		if p.peekTokenIs(token.RBRACE) {
			break
		}
		if p.notExpectPeekToken(expectedTokens...) {
			return nil
		}
	}

	return stmt
}

func (p *Parser) serviceItemStartTokens() []interface{} {
	tokens := []interface{}{token.AT_DOC, token.AT_HANDLER}
	if p.mode == ParseModeCron {
		tokens = append(tokens, token.AT_CRON, token.AT_CRON_RETRY)
	}
	return tokens
}

func (p *Parser) parseServiceItemStmt() *ast.ServiceItemStmt {
	var stmt = &ast.ServiceItemStmt{}

	// optional @doc
	if p.peekTokenIs(token.AT_DOC) {
		if !p.nextToken() {
			return nil
		}
		atDocStmt := p.parseAtDocStmt()
		if atDocStmt == nil {
			return nil
		}
		stmt.AtDoc = atDocStmt
	}

	// optional @cron (only in cron mode)
	if p.mode == ParseModeCron && p.peekTokenIs(token.AT_CRON) {
		if !p.nextToken() {
			return nil
		}
		atCronStmt := p.parseAtCronStmt()
		if atCronStmt == nil {
			return nil
		}
		stmt.AtCron = atCronStmt
	}

	// optional @cronRetry (only in cron mode)
	if p.mode == ParseModeCron && p.peekTokenIs(token.AT_CRON_RETRY) {
		if !p.nextToken() {
			return nil
		}
		atCronRetryStmt := p.parseAtCronRetryStmt()
		if atCronRetryStmt == nil {
			return nil
		}
		stmt.AtCronRetry = atCronRetryStmt
	}

	// required @handler
	if !p.advanceIfPeekTokenIs(token.AT_HANDLER, token.RBRACE) {
		return nil
	}
	if p.peekTokenIs(token.RBRACE) {
		return stmt
	}

	atHandlerStmt := p.parseAtHandlerStmt()
	if atHandlerStmt == nil {
		return nil
	}
	stmt.AtHandler = atHandlerStmt

	// route: task_type[(Param)] or queue.name
	route := p.parseExtensionRoute()
	if route == nil {
		return nil
	}
	stmt.Route = route

	return stmt
}

// parseExtensionRoute parses the extension route:
// cron: IDENT [(IDENT)]
// rabbitmq: IDENT [.IDENT]*
func (p *Parser) parseExtensionRoute() *ast.RouteStmt {
	var stmt = &ast.RouteStmt{}

	// expect IDENT as start
	if !p.advanceIfPeekTokenIs(token.IDENT) {
		return nil
	}

	nameText := p.curTok.Text
	namePos := p.curTok.Position

	if p.mode == ParseModeRabbitMQ {
		// rabbitmq: allow dot-separated queue names like order.created
		for p.peekTokenIs(token.DOT) {
			if !p.nextToken() { // consume DOT
				return nil
			}
			nameText += "."
			if !p.advanceIfPeekTokenIs(token.IDENT) {
				return nil
			}
			nameText += p.curTok.Text
		}
	}

	nameNode := ast.NewTokenNode(token.Token{
		Type:     token.IDENT,
		Text:     nameText,
		Position: namePos,
	})
	nameNode.SetLeadingCommentGroup(p.curTokenNode().LeadingCommentGroup)
	stmt.Name = nameNode

	// optional (ParamType) for both cron and rabbitmq
	if p.peekTokenIs(token.LPAREN) {
		bodyStmt := p.parseBodyStmt()
		if bodyStmt == nil {
			return nil
		}
		stmt.Request = bodyStmt
	}

	return stmt
}

// ==================== @doc / @handler / @cron / @cronRetry / @server ====================

func (p *Parser) parseAtDocStmt() ast.AtDocStmt {
	if p.notExpectPeekToken(token.LPAREN, token.STRING) {
		return nil
	}
	if p.peekTokenIs(token.LPAREN) {
		return p.parseAtDocGroupStmt()
	}
	return p.parseAtDocLiteralStmt()
}

func (p *Parser) parseAtDocGroupStmt() ast.AtDocStmt {
	var stmt = &ast.AtDocGroupStmt{}
	stmt.AtDoc = p.curTokenNode()

	if !p.advanceIfPeekTokenIs(token.LPAREN) {
		return nil
	}
	stmt.LParen = p.curTokenNode()

	for p.curTokenIsNotEof() && p.peekTokenIsNot(token.RPAREN) {
		expr := p.parseKVExpression()
		if expr == nil {
			return nil
		}
		stmt.Values = append(stmt.Values, expr)
		if p.notExpectPeekToken(token.RPAREN, token.IDENT) {
			return nil
		}
	}

	if !p.advanceIfPeekTokenIs(token.RPAREN) {
		return nil
	}
	stmt.RParen = p.curTokenNode()

	return stmt
}

func (p *Parser) parseAtDocLiteralStmt() ast.AtDocStmt {
	var stmt = &ast.AtDocLiteralStmt{}
	stmt.AtDoc = p.curTokenNode()

	if !p.advanceIfPeekTokenIs(token.STRING) {
		return nil
	}
	stmt.Value = p.curTokenNode()

	return stmt
}

func (p *Parser) parseAtHandlerStmt() *ast.AtHandlerStmt {
	var stmt = &ast.AtHandlerStmt{}
	stmt.AtHandler = p.curTokenNode()

	if !p.advanceIfPeekTokenIs(token.IDENT) {
		return nil
	}
	stmt.Name = p.curTokenNode()

	return stmt
}

func (p *Parser) parseAtCronStmt() *ast.AtCronStmt {
	var stmt = &ast.AtCronStmt{}
	stmt.AtCron = p.curTokenNode()

	if !p.advanceIfPeekTokenIs(token.STRING) {
		return nil
	}
	stmt.Value = p.curTokenNode()

	return stmt
}

func (p *Parser) parseAtCronRetryStmt() *ast.AtCronRetryStmt {
	var stmt = &ast.AtCronRetryStmt{}
	stmt.AtCronRetry = p.curTokenNode()

	if !p.advanceIfPeekTokenIs(token.INT) {
		return nil
	}
	stmt.Value = p.curTokenNode()

	return stmt
}

func (p *Parser) parseAtServerStmt() *ast.AtServerStmt {
	var stmt = &ast.AtServerStmt{}
	stmt.AtServer = p.curTokenNode()

	if !p.advanceIfPeekTokenIs(token.LPAREN) {
		return nil
	}
	stmt.LParen = p.curTokenNode()

	for p.curTokenIsNotEof() && p.peekTokenIsNot(token.RPAREN) {
		expr := p.parseAtServerKVExpression()
		if expr == nil {
			return nil
		}
		stmt.Values = append(stmt.Values, expr)
		if p.notExpectPeekToken(token.RPAREN, token.IDENT) {
			return nil
		}
	}

	if !p.advanceIfPeekTokenIs(token.RPAREN) {
		return nil
	}
	stmt.RParen = p.curTokenNode()

	return stmt
}

func (p *Parser) parseAtServerKVExpression() *ast.KVExpr {
	var expr = &ast.KVExpr{}

	if !p.advanceIfPeekTokenIs(token.IDENT, token.RPAREN) {
		return nil
	}
	expr.Key = p.curTokenNode()

	if !p.advanceIfPeekTokenIs(token.COLON) {
		return nil
	}
	expr.Colon = p.curTokenNode()

	if p.notExpectPeekToken(token.QUO, token.IDENT, token.INT, token.STRING) {
		return nil
	}

	// handle various value types
	if p.peekTokenIs(token.QUO) {
		if !p.nextToken() {
			return nil
		}
		slashTok := p.curTok
		var pathText = slashTok.Text
		if !p.advanceIfPeekTokenIs(token.IDENT) {
			return nil
		}
		pathText += p.curTok.Text
		valueTok := token.Token{Text: pathText, Position: slashTok.Position}
		node := ast.NewTokenNode(valueTok)
		node.SetLeadingCommentGroup(p.curTokenNode().LeadingCommentGroup)
		expr.Value = node
		return expr
	}

	if p.peekTokenIs(token.INT) {
		if !p.nextToken() {
			return nil
		}
		node := ast.NewTokenNode(p.curTok)
		node.SetLeadingCommentGroup(p.curTokenNode().LeadingCommentGroup)
		expr.Value = node
		return expr
	}

	if p.peekTokenIs(token.STRING) {
		if !p.nextToken() {
			return nil
		}
		node := ast.NewTokenNode(p.curTok)
		node.SetLeadingCommentGroup(p.curTokenNode().LeadingCommentGroup)
		expr.Value = node
		return expr
	}

	// IDENT value, possibly with commas or hyphens
	if !p.advanceIfPeekTokenIs(token.IDENT) {
		return nil
	}
	valueTok := p.curTok
	leadingCG := p.curTokenNode().LeadingCommentGroup

	if p.peekTokenIs(token.COMMA) {
		for p.peekTokenIs(token.COMMA) {
			if !p.nextToken() {
				return nil
			}
			commaTok := p.curTok
			if !p.advanceIfPeekTokenIs(token.IDENT) {
				return nil
			}
			idTok := p.curTok
			valueTok = token.Token{
				Text:     valueTok.Text + commaTok.Text + idTok.Text,
				Position: valueTok.Position,
			}
			leadingCG = p.curTokenNode().LeadingCommentGroup
		}
	} else if p.peekTokenIs(token.SUB) {
		for p.peekTokenIs(token.SUB) {
			if !p.nextToken() {
				return nil
			}
			subTok := p.curTok
			if !p.advanceIfPeekTokenIs(token.IDENT) {
				return nil
			}
			idTok := p.curTok
			valueTok = token.Token{
				Text:     valueTok.Text + subTok.Text + idTok.Text,
				Position: valueTok.Position,
			}
			leadingCG = p.curTokenNode().LeadingCommentGroup
		}
	}

	node := ast.NewTokenNode(valueTok)
	node.SetLeadingCommentGroup(leadingCG)
	expr.Value = node
	return expr
}

// ==================== Syntax / Info / Import / Type ====================

func (p *Parser) parseSyntaxStmt() *ast.SyntaxStmt {
	var stmt = &ast.SyntaxStmt{}
	stmt.Syntax = p.curTokenNode()

	if !p.advanceIfPeekTokenIs(token.ASSIGN) {
		return nil
	}
	stmt.Assign = p.curTokenNode()

	if !p.advanceIfPeekTokenIs(token.STRING) {
		return nil
	}
	stmt.Value = p.curTokenNode()

	return stmt
}

func (p *Parser) parseInfoStmt() *ast.InfoStmt {
	var stmt = &ast.InfoStmt{}
	stmt.Info = p.curTokenNode()

	if !p.advanceIfPeekTokenIs(token.LPAREN) {
		return nil
	}
	stmt.LParen = p.curTokenNode()

	for p.curTokenIsNotEof() && p.peekTokenIsNot(token.RPAREN) {
		expr := p.parseKVExpression()
		if expr == nil {
			return nil
		}
		stmt.Values = append(stmt.Values, expr)
		if p.notExpectPeekToken(token.RPAREN, token.IDENT) {
			return nil
		}
	}

	if !p.advanceIfPeekTokenIs(token.RPAREN) {
		return nil
	}
	stmt.RParen = p.curTokenNode()

	return stmt
}

func (p *Parser) parseKVExpression() *ast.KVExpr {
	var expr = &ast.KVExpr{}

	if !p.advanceIfPeekTokenIs(token.IDENT) {
		return nil
	}
	expr.Key = p.curTokenNode()

	if !p.advanceIfPeekTokenIs(token.COLON) {
		return nil
	}
	expr.Colon = p.curTokenNode()

	if !p.advanceIfPeekTokenIs(token.STRING, token.RAW_STRING, token.IDENT) {
		return nil
	}
	expr.Value = p.curTokenNode()

	return expr
}

func (p *Parser) parseImportStmt() ast.ImportStmt {
	if p.notExpectPeekToken(token.LPAREN, token.STRING) {
		return nil
	}
	if p.peekTokenIs(token.LPAREN) {
		return p.parseImportGroupStmt()
	}
	return p.parseImportLiteralStmt()
}

func (p *Parser) parseImportLiteralStmt() ast.ImportStmt {
	var stmt = &ast.ImportLiteralStmt{}
	stmt.Import = p.curTokenNode()

	if !p.advanceIfPeekTokenIs(token.STRING) {
		return nil
	}
	stmt.Value = p.curTokenNode()

	return stmt
}

func (p *Parser) parseImportGroupStmt() ast.ImportStmt {
	var stmt = &ast.ImportGroupStmt{}
	stmt.Import = p.curTokenNode()

	if !p.advanceIfPeekTokenIs(token.LPAREN) {
		return nil
	}
	stmt.LParen = p.curTokenNode()

	for p.curTokenIsNotEof() && p.peekTokenIsNot(token.RPAREN) {
		if !p.advanceIfPeekTokenIs(token.STRING) {
			return nil
		}
		stmt.Values = append(stmt.Values, p.curTokenNode())
		if p.notExpectPeekToken(token.RPAREN, token.STRING) {
			return nil
		}
	}

	if !p.advanceIfPeekTokenIs(token.RPAREN) {
		return nil
	}
	stmt.RParen = p.curTokenNode()

	return stmt
}

// ==================== Type Parsing ====================

func (p *Parser) parseTypeStmt() ast.TypeStmt {
	switch {
	case p.peekTokenIs(token.LPAREN):
		return p.parseTypeGroupStmt()
	case p.peekTokenIs(token.IDENT):
		return p.parseTypeLiteralStmt()
	default:
		p.expectPeekToken(token.LPAREN, token.IDENT)
		return nil
	}
}

func (p *Parser) parseTypeLiteralStmt() ast.TypeStmt {
	var stmt = &ast.TypeLiteralStmt{}
	stmt.Type = p.curTokenNode()

	expr := p.parseTypeExpr()
	if expr == nil {
		return nil
	}
	stmt.Expr = expr

	return stmt
}

func (p *Parser) parseTypeGroupStmt() ast.TypeStmt {
	var stmt = &ast.TypeGroupStmt{}
	stmt.Type = p.curTokenNode()

	if !p.nextToken() {
		return nil
	}
	stmt.LParen = p.curTokenNode()

	exprList := p.parseTypeExprList()
	if exprList == nil {
		return nil
	}
	stmt.ExprList = exprList

	if !p.advanceIfPeekTokenIs(token.RPAREN) {
		return nil
	}
	stmt.RParen = p.curTokenNode()

	return stmt
}

func (p *Parser) parseTypeExprList() []*ast.TypeExpr {
	if !p.expectPeekToken(token.IDENT, token.RPAREN) {
		return nil
	}

	var exprList = make([]*ast.TypeExpr, 0)
	for p.curTokenIsNotEof() && p.peekTokenIsNot(token.RPAREN, token.EOF) {
		expr := p.parseTypeExpr()
		if expr == nil {
			return nil
		}
		exprList = append(exprList, expr)
		if !p.expectPeekToken(token.IDENT, token.RPAREN) {
			return nil
		}
	}

	return exprList
}

func (p *Parser) parseTypeExpr() *ast.TypeExpr {
	var expr = &ast.TypeExpr{}
	if !p.advanceIfPeekTokenIs(token.IDENT) {
		return nil
	}
	if p.curTokenIsKeyword() {
		return nil
	}
	expr.Name = p.curTokenNode()

	if p.peekTokenIs(token.ASSIGN) {
		if !p.nextToken() {
			return nil
		}
		expr.Assign = p.curTokenNode()
	}

	dt := p.parseDataType()
	if isNil(dt) {
		return nil
	}
	expr.DataType = dt

	return expr
}

func (p *Parser) parseDataType() ast.DataType {
	switch {
	case p.peekTokenIs(token.Any):
		return p.parseAnyDataType()
	case p.peekTokenIs(token.LBRACE):
		return p.parseStructDataType()
	case p.peekTokenIs(token.IDENT):
		if p.peekTokenIs(token.MapKeyword) {
			return p.parseMapDataType()
		}
		if !p.nextToken() {
			return nil
		}
		if p.curTokenIsKeyword() {
			return nil
		}
		return &ast.BaseDataType{Base: p.curTokenNode()}
	case p.peekTokenIs(token.LBRACK):
		if !p.nextToken() {
			return nil
		}
		switch {
		case p.peekTokenIs(token.RBRACK):
			return p.parseSliceDataType()
		case p.peekTokenIs(token.INT, token.ELLIPSIS):
			return p.parseArrayDataType()
		default:
			p.expectPeekToken(token.RBRACK, token.INT, token.ELLIPSIS)
			return nil
		}
	case p.peekTokenIs(token.ANY):
		return p.parseInterfaceDataType()
	case p.peekTokenIs(token.MUL):
		return p.parsePointerDataType()
	default:
		p.expectPeekToken(token.IDENT, token.LBRACK, token.ANY, token.MUL, token.LBRACE)
		return nil
	}
}

func (p *Parser) parseStructDataType() *ast.StructDataType {
	var tp = &ast.StructDataType{}
	if !p.nextToken() {
		return nil
	}
	tp.LBrace = p.curTokenNode()

	if p.notExpectPeekToken(token.IDENT, token.MUL, token.RBRACE) {
		return nil
	}

	elems := p.parseElemExprList()
	if elems == nil {
		return nil
	}
	tp.Elements = elems

	if !p.advanceIfPeekTokenIs(token.RBRACE) {
		return nil
	}
	tp.RBrace = p.curTokenNode()

	return tp
}

func (p *Parser) parseElemExprList() ast.ElemExprList {
	var list = make(ast.ElemExprList, 0)
	for p.curTokenIsNotEof() && p.peekTokenIsNot(token.RBRACE, token.EOF) {
		if p.notExpectPeekToken(token.IDENT, token.MUL, token.RBRACE) {
			return nil
		}
		expr := p.parseElemExpr()
		if expr == nil {
			return nil
		}
		list = append(list, expr)
		if p.notExpectPeekToken(token.IDENT, token.MUL, token.RBRACE) {
			return nil
		}
	}
	return list
}

func (p *Parser) parseElemExpr() *ast.ElemExpr {
	var expr = &ast.ElemExpr{}
	if !p.advanceIfPeekTokenIs(token.IDENT, token.MUL) {
		return nil
	}
	if p.curTokenIsKeyword() {
		return nil
	}

	identNode := p.curTokenNode()
	if p.curTokenIs(token.MUL) {
		star := p.curTokenNode()
		if !p.advanceIfPeekTokenIs(token.IDENT) {
			return nil
		}
		var dt ast.DataType
		if p.curTokenIs(token.Any) {
			dt = &ast.AnyDataType{Any: p.curTokenNode()}
		} else {
			dt = &ast.BaseDataType{Base: p.curTokenNode()}
		}
		expr.DataType = &ast.PointerDataType{Star: star, DataType: dt}
	} else if p.peekTok.Line() > identNode.Token.Line() || p.peekTokenIs(token.RAW_STRING) {
		if p.curTokenIs(token.Any) {
			expr.DataType = &ast.AnyDataType{Any: identNode}
		} else {
			expr.DataType = &ast.BaseDataType{Base: identNode}
		}
	} else {
		expr.Name = append(expr.Name, identNode)
		if p.notExpectPeekToken(token.COMMA, token.IDENT, token.LBRACK, token.ANY, token.MUL, token.LBRACE) {
			return nil
		}

		for p.peekTokenIs(token.COMMA) {
			if !p.nextToken() {
				return nil
			}
			if !p.advanceIfPeekTokenIs(token.IDENT) {
				return nil
			}
			if p.curTokenIsKeyword() {
				return nil
			}
			expr.Name = append(expr.Name, p.curTokenNode())
		}

		dt := p.parseDataType()
		if isNil(dt) {
			return nil
		}
		expr.DataType = dt
	}

	if p.notExpectPeekToken(token.RAW_STRING, token.MUL, token.IDENT, token.RBRACE) {
		return nil
	}
	if p.peekTokenIs(token.RAW_STRING) {
		if !p.nextToken() {
			return nil
		}
		expr.Tag = p.curTokenNode()
	}

	return expr
}

func (p *Parser) parseAnyDataType() *ast.AnyDataType {
	if !p.nextToken() {
		return nil
	}
	return &ast.AnyDataType{Any: p.curTokenNode()}
}

func (p *Parser) parsePointerDataType() *ast.PointerDataType {
	var tp = &ast.PointerDataType{}
	if !p.nextToken() {
		return nil
	}
	tp.Star = p.curTokenNode()

	if p.notExpectPeekToken(token.IDENT, token.LBRACK, token.ANY, token.MUL) {
		return nil
	}

	dt := p.parseDataType()
	if isNil(dt) {
		return nil
	}
	tp.DataType = dt

	return tp
}

func (p *Parser) parseInterfaceDataType() *ast.InterfaceDataType {
	if !p.nextToken() {
		return nil
	}
	return &ast.InterfaceDataType{Interface: p.curTokenNode()}
}

func (p *Parser) parseMapDataType() *ast.MapDataType {
	var tp = &ast.MapDataType{}
	if !p.nextToken() {
		return nil
	}
	tp.Map = p.curTokenNode()

	if !p.advanceIfPeekTokenIs(token.LBRACK) {
		return nil
	}
	tp.LBrack = p.curTokenNode()

	dt := p.parseDataType()
	if isNil(dt) {
		return nil
	}
	tp.Key = dt

	if !p.advanceIfPeekTokenIs(token.RBRACK) {
		return nil
	}
	tp.RBrack = p.curTokenNode()

	dt = p.parseDataType()
	if isNil(dt) {
		return nil
	}
	tp.Value = dt

	return tp
}

func (p *Parser) parseArrayDataType() *ast.ArrayDataType {
	var tp = &ast.ArrayDataType{}
	tp.LBrack = p.curTokenNode()

	if !p.nextToken() {
		return nil
	}
	tp.Length = p.curTokenNode()

	if !p.advanceIfPeekTokenIs(token.RBRACK) {
		return nil
	}
	tp.RBrack = p.curTokenNode()

	dt := p.parseDataType()
	if isNil(dt) {
		return nil
	}
	tp.DataType = dt

	return tp
}

func (p *Parser) parseSliceDataType() *ast.SliceDataType {
	var tp = &ast.SliceDataType{}
	tp.LBrack = p.curTokenNode()

	if !p.advanceIfPeekTokenIs(token.RBRACK) {
		return nil
	}
	tp.RBrack = p.curTokenNode()

	dt := p.parseDataType()
	if isNil(dt) {
		return nil
	}
	tp.DataType = dt

	return tp
}

// ==================== Body (Request Param) ====================

func (p *Parser) parseBodyStmt() *ast.BodyStmt {
	var stmt = &ast.BodyStmt{}
	if !p.advanceIfPeekTokenIs(token.LPAREN) {
		return nil
	}
	stmt.LParen = p.curTokenNode()

	if p.peekTokenIs(token.RPAREN) {
		if !p.nextToken() {
			return nil
		}
		stmt.RParen = p.curTokenNode()
		return stmt
	}

	expr := p.parseBodyExpr()
	if expr == nil {
		return nil
	}
	stmt.Body = expr

	if !p.advanceIfPeekTokenIs(token.RPAREN) {
		return nil
	}
	stmt.RParen = p.curTokenNode()

	return stmt
}

func (p *Parser) parseBodyExpr() *ast.BodyExpr {
	var expr = &ast.BodyExpr{}
	switch {
	case p.peekTokenIs(token.LBRACK):
		if !p.nextToken() {
			return nil
		}
		expr.LBrack = p.curTokenNode()
		if !p.advanceIfPeekTokenIs(token.RBRACK) {
			return nil
		}
		expr.RBrack = p.curTokenNode()
		if p.peekTokenIs(token.MUL) {
			if !p.nextToken() {
				return nil
			}
			expr.Star = p.curTokenNode()
		}
		if !p.advanceIfPeekTokenIs(token.IDENT) {
			return nil
		}
		expr.Value = p.curTokenNode()
		return expr
	case p.peekTokenIs(token.MUL):
		if !p.nextToken() {
			return nil
		}
		expr.Star = p.curTokenNode()
		if !p.advanceIfPeekTokenIs(token.IDENT) {
			return nil
		}
		expr.Value = p.curTokenNode()
		return expr
	case p.peekTokenIs(token.IDENT):
		if !p.nextToken() {
			return nil
		}
		expr.Value = p.curTokenNode()
		return expr
	default:
		p.expectPeekToken(token.LBRACK, token.MUL, token.IDENT)
		return nil
	}
}

// ==================== Service Name ====================

func (p *Parser) parseServiceNameExpr() *ast.ServiceNameExpr {
	var expr = &ast.ServiceNameExpr{}
	var text = p.curTok.Text
	pos := p.curTok.Position

	// allow name-suffix (e.g., user-cron, order-mq)
	for p.peekTokenIs(token.SUB) {
		if !p.nextToken() {
			return nil
		}
		text += p.curTok.Text
		if !p.advanceIfPeekTokenIs(token.IDENT) {
			return nil
		}
		text += p.curTok.Text
	}

	node := ast.NewTokenNode(token.Token{
		Type:     token.IDENT,
		Text:     text,
		Position: pos,
	})
	node.SetLeadingCommentGroup(p.curTokenNode().LeadingCommentGroup)
	expr.Name = node

	return expr
}

// ==================== Token Helpers ====================

func (p *Parser) curTokenIsNotEof() bool {
	return p.curTok.Type != token.EOF
}

func (p *Parser) curTokenIsKeyword() bool {
	tp, ok := token.LookupKeyword(p.curTok.Text)
	if ok {
		p.errors = append(p.errors, fmt.Errorf("%s syntax error: expected 'IDENT', got '%s'", p.curTok.Position.String(), tp.String()))
		return true
	}
	return false
}

func (p *Parser) curTokenIs(expected ...interface{}) bool {
	for _, v := range expected {
		switch val := v.(type) {
		case token.Type:
			if p.curTok.Type == val {
				return true
			}
		case string:
			if p.curTok.Text == val {
				return true
			}
		}
	}
	return false
}

func (p *Parser) advanceIfPeekTokenIs(expected ...interface{}) bool {
	if p.expectPeekToken(expected...) {
		if !p.nextToken() {
			return false
		}
		return true
	}
	return false
}

func (p *Parser) peekTokenIs(expected ...interface{}) bool {
	for _, v := range expected {
		switch val := v.(type) {
		case token.Type:
			if p.peekTok.Type == val {
				return true
			}
		case string:
			if p.peekTok.Text == val {
				return true
			}
		}
	}
	return false
}

func (p *Parser) peekTokenIsNot(expected ...interface{}) bool {
	for _, v := range expected {
		switch val := v.(type) {
		case token.Type:
			if p.peekTok.Type == val {
				return false
			}
		case string:
			if p.peekTok.Text == val {
				return false
			}
		}
	}
	return true
}

func (p *Parser) notExpectPeekToken(expected ...interface{}) bool {
	if !p.peekTokenIsNot(expected...) {
		return false
	}

	var expectedString []string
	for _, v := range expected {
		expectedString = append(expectedString, fmt.Sprintf("'%s'", v))
	}

	var got string
	if p.peekTok.Type == token.ILLEGAL {
		got = p.peekTok.Text
	} else {
		got = p.peekTok.Type.String()
	}

	var err error
	if p.peekTok.Type == token.EOF {
		position := p.curTok.Position
		position.Column = position.Column + len(p.curTok.Text)
		err = fmt.Errorf("%s syntax error: expected %s, got '%s'", position, strings.Join(expectedString, " | "), got)
	} else {
		err = fmt.Errorf("%s syntax error: expected %s, got '%s'", p.peekTok.Position, strings.Join(expectedString, " | "), got)
	}
	p.errors = append(p.errors, err)
	return true
}

func (p *Parser) expectPeekToken(expected ...interface{}) bool {
	if p.peekTokenIs(expected...) {
		return true
	}

	var expectedString []string
	for _, v := range expected {
		expectedString = append(expectedString, fmt.Sprintf("'%s'", v))
	}

	var got string
	if p.peekTok.Type == token.ILLEGAL {
		got = p.peekTok.Text
	} else {
		got = p.peekTok.Type.String()
	}

	var err error
	if p.peekTok.Type == token.EOF {
		position := p.curTok.Position
		position.Column = position.Column + len(p.curTok.Text)
		err = fmt.Errorf("%s syntax error: expected %s, got '%s'", position, strings.Join(expectedString, " | "), got)
	} else {
		err = fmt.Errorf("%s syntax error: expected %s, got '%s'", p.peekTok.Position, strings.Join(expectedString, " | "), got)
	}
	p.errors = append(p.errors, err)
	return false
}

func (p *Parser) init() bool {
	if !p.nextToken() {
		return false
	}
	return p.nextToken()
}

func (p *Parser) nextToken() bool {
	var err error
	p.curTok = p.peekTok
	var line = -1
	if p.curTok.Valid() {
		if p.curTokenIs(token.EOF) {
			for _, v := range p.headCommentGroup {
				p.appendStmt(v)
			}
			p.headCommentGroup = ast.CommentGroup{}
			return true
		}

		node := ast.NewTokenNode(p.curTok)
		if p.headCommentGroup.Valid() {
			node.HeadCommentGroup = append(node.HeadCommentGroup, p.headCommentGroup...)
			p.headCommentGroup = ast.CommentGroup{}
		}
		p.node[p.curTok] = node
		line = p.curTok.Line()
	}

	p.peekTok, err = p.s.NextToken()
	if err != nil {
		p.errors = append(p.errors, err)
		return false
	}

	var leadingCommentGroup ast.CommentGroup
	for p.peekTok.Type == token.COMMENT || p.peekTok.Type == token.DOCUMENT {
		commentStmt := &ast.CommentStmt{Comment: p.peekTok}
		if p.peekTok.Line() == line && line > -1 {
			leadingCommentGroup = append(leadingCommentGroup, commentStmt)
		} else {
			p.headCommentGroup = append(p.headCommentGroup, commentStmt)
		}

		p.peekTok, err = p.s.NextToken()
		if err != nil {
			p.errors = append(p.errors, err)
			return false
		}
	}

	if len(leadingCommentGroup) > 0 {
		p.curTokenNode().SetLeadingCommentGroup(leadingCommentGroup)
	}

	return true
}

func (p *Parser) curTokenNode() *ast.TokenNode {
	return p.node[p.curTok]
}

func (p *Parser) appendStmt(stmt ...ast.Stmt) {
	p.api.Stmts = append(p.api.Stmts, stmt...)
}

// CheckErrors check parser errors.
func (p *Parser) CheckErrors() error {
	if len(p.errors) == 0 {
		return nil
	}
	var errors []string
	for _, e := range p.errors {
		errors = append(errors, e.Error())
	}
	return fmt.Errorf(strings.Join(errors, "\n"))
}

func isNil(v interface{}) bool {
	if v == nil {
		return true
	}
	vo := reflect.ValueOf(v)
	if vo.Kind() == reflect.Ptr {
		return vo.IsNil()
	}
	return false
}
