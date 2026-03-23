// Code generated from CztctlParser.g4 by ANTLR 4.7.2. DO NOT EDIT.

package cztctl // CztctlParser
import "github.com/zeromicro/antlr"

// BaseCztctlParserListener is a complete listener for a parse tree produced by CztctlParserParser.
type BaseCztctlParserListener struct{}

var _ CztctlParserListener = &BaseCztctlParserListener{}

// VisitTerminal is called when a terminal node is visited.
func (s *BaseCztctlParserListener) VisitTerminal(node antlr.TerminalNode) {}

// VisitErrorNode is called when an error node is visited.
func (s *BaseCztctlParserListener) VisitErrorNode(node antlr.ErrorNode) {}

// EnterEveryRule is called when any rule is entered.
func (s *BaseCztctlParserListener) EnterEveryRule(ctx antlr.ParserRuleContext) {}

// ExitEveryRule is called when any rule is exited.
func (s *BaseCztctlParserListener) ExitEveryRule(ctx antlr.ParserRuleContext) {}

// EnterApi is called when production api is entered.
func (s *BaseCztctlParserListener) EnterApi(ctx *ApiContext) {}

// ExitApi is called when production api is exited.
func (s *BaseCztctlParserListener) ExitApi(ctx *ApiContext) {}

// EnterSpec is called when production spec is entered.
func (s *BaseCztctlParserListener) EnterSpec(ctx *SpecContext) {}

// ExitSpec is called when production spec is exited.
func (s *BaseCztctlParserListener) ExitSpec(ctx *SpecContext) {}

// EnterSyntaxLit is called when production syntaxLit is entered.
func (s *BaseCztctlParserListener) EnterSyntaxLit(ctx *SyntaxLitContext) {}

// ExitSyntaxLit is called when production syntaxLit is exited.
func (s *BaseCztctlParserListener) ExitSyntaxLit(ctx *SyntaxLitContext) {}

// EnterImportSpec is called when production importSpec is entered.
func (s *BaseCztctlParserListener) EnterImportSpec(ctx *ImportSpecContext) {}

// ExitImportSpec is called when production importSpec is exited.
func (s *BaseCztctlParserListener) ExitImportSpec(ctx *ImportSpecContext) {}

// EnterImportLit is called when production importLit is entered.
func (s *BaseCztctlParserListener) EnterImportLit(ctx *ImportLitContext) {}

// ExitImportLit is called when production importLit is exited.
func (s *BaseCztctlParserListener) ExitImportLit(ctx *ImportLitContext) {}

// EnterImportBlock is called when production importBlock is entered.
func (s *BaseCztctlParserListener) EnterImportBlock(ctx *ImportBlockContext) {}

// ExitImportBlock is called when production importBlock is exited.
func (s *BaseCztctlParserListener) ExitImportBlock(ctx *ImportBlockContext) {}

// EnterImportBlockValue is called when production importBlockValue is entered.
func (s *BaseCztctlParserListener) EnterImportBlockValue(ctx *ImportBlockValueContext) {}

// ExitImportBlockValue is called when production importBlockValue is exited.
func (s *BaseCztctlParserListener) ExitImportBlockValue(ctx *ImportBlockValueContext) {}

// EnterImportValue is called when production importValue is entered.
func (s *BaseCztctlParserListener) EnterImportValue(ctx *ImportValueContext) {}

// ExitImportValue is called when production importValue is exited.
func (s *BaseCztctlParserListener) ExitImportValue(ctx *ImportValueContext) {}

// EnterInfoSpec is called when production infoSpec is entered.
func (s *BaseCztctlParserListener) EnterInfoSpec(ctx *InfoSpecContext) {}

// ExitInfoSpec is called when production infoSpec is exited.
func (s *BaseCztctlParserListener) ExitInfoSpec(ctx *InfoSpecContext) {}

// EnterTypeSpec is called when production typeSpec is entered.
func (s *BaseCztctlParserListener) EnterTypeSpec(ctx *TypeSpecContext) {}

// ExitTypeSpec is called when production typeSpec is exited.
func (s *BaseCztctlParserListener) ExitTypeSpec(ctx *TypeSpecContext) {}

// EnterTypeLit is called when production typeLit is entered.
func (s *BaseCztctlParserListener) EnterTypeLit(ctx *TypeLitContext) {}

// ExitTypeLit is called when production typeLit is exited.
func (s *BaseCztctlParserListener) ExitTypeLit(ctx *TypeLitContext) {}

// EnterTypeBlock is called when production typeBlock is entered.
func (s *BaseCztctlParserListener) EnterTypeBlock(ctx *TypeBlockContext) {}

// ExitTypeBlock is called when production typeBlock is exited.
func (s *BaseCztctlParserListener) ExitTypeBlock(ctx *TypeBlockContext) {}

// EnterTypeLitBody is called when production typeLitBody is entered.
func (s *BaseCztctlParserListener) EnterTypeLitBody(ctx *TypeLitBodyContext) {}

// ExitTypeLitBody is called when production typeLitBody is exited.
func (s *BaseCztctlParserListener) ExitTypeLitBody(ctx *TypeLitBodyContext) {}

// EnterTypeBlockBody is called when production typeBlockBody is entered.
func (s *BaseCztctlParserListener) EnterTypeBlockBody(ctx *TypeBlockBodyContext) {}

// ExitTypeBlockBody is called when production typeBlockBody is exited.
func (s *BaseCztctlParserListener) ExitTypeBlockBody(ctx *TypeBlockBodyContext) {}

// EnterTypeStruct is called when production typeStruct is entered.
func (s *BaseCztctlParserListener) EnterTypeStruct(ctx *TypeStructContext) {}

// ExitTypeStruct is called when production typeStruct is exited.
func (s *BaseCztctlParserListener) ExitTypeStruct(ctx *TypeStructContext) {}

// EnterTypeAlias is called when production typeAlias is entered.
func (s *BaseCztctlParserListener) EnterTypeAlias(ctx *TypeAliasContext) {}

// ExitTypeAlias is called when production typeAlias is exited.
func (s *BaseCztctlParserListener) ExitTypeAlias(ctx *TypeAliasContext) {}

// EnterTypeBlockStruct is called when production typeBlockStruct is entered.
func (s *BaseCztctlParserListener) EnterTypeBlockStruct(ctx *TypeBlockStructContext) {}

// ExitTypeBlockStruct is called when production typeBlockStruct is exited.
func (s *BaseCztctlParserListener) ExitTypeBlockStruct(ctx *TypeBlockStructContext) {}

// EnterTypeBlockAlias is called when production typeBlockAlias is entered.
func (s *BaseCztctlParserListener) EnterTypeBlockAlias(ctx *TypeBlockAliasContext) {}

// ExitTypeBlockAlias is called when production typeBlockAlias is exited.
func (s *BaseCztctlParserListener) ExitTypeBlockAlias(ctx *TypeBlockAliasContext) {}

// EnterField is called when production field is entered.
func (s *BaseCztctlParserListener) EnterField(ctx *FieldContext) {}

// ExitField is called when production field is exited.
func (s *BaseCztctlParserListener) ExitField(ctx *FieldContext) {}

// EnterNormalField is called when production normalField is entered.
func (s *BaseCztctlParserListener) EnterNormalField(ctx *NormalFieldContext) {}

// ExitNormalField is called when production normalField is exited.
func (s *BaseCztctlParserListener) ExitNormalField(ctx *NormalFieldContext) {}

// EnterAnonymousFiled is called when production anonymousFiled is entered.
func (s *BaseCztctlParserListener) EnterAnonymousFiled(ctx *AnonymousFiledContext) {}

// ExitAnonymousFiled is called when production anonymousFiled is exited.
func (s *BaseCztctlParserListener) ExitAnonymousFiled(ctx *AnonymousFiledContext) {}

// EnterDataType is called when production dataType is entered.
func (s *BaseCztctlParserListener) EnterDataType(ctx *DataTypeContext) {}

// ExitDataType is called when production dataType is exited.
func (s *BaseCztctlParserListener) ExitDataType(ctx *DataTypeContext) {}

// EnterPointerType is called when production pointerType is entered.
func (s *BaseCztctlParserListener) EnterPointerType(ctx *PointerTypeContext) {}

// ExitPointerType is called when production pointerType is exited.
func (s *BaseCztctlParserListener) ExitPointerType(ctx *PointerTypeContext) {}

// EnterMapType is called when production mapType is entered.
func (s *BaseCztctlParserListener) EnterMapType(ctx *MapTypeContext) {}

// ExitMapType is called when production mapType is exited.
func (s *BaseCztctlParserListener) ExitMapType(ctx *MapTypeContext) {}

// EnterArrayType is called when production arrayType is entered.
func (s *BaseCztctlParserListener) EnterArrayType(ctx *ArrayTypeContext) {}

// ExitArrayType is called when production arrayType is exited.
func (s *BaseCztctlParserListener) ExitArrayType(ctx *ArrayTypeContext) {}

// EnterServiceSpec is called when production serviceSpec is entered.
func (s *BaseCztctlParserListener) EnterServiceSpec(ctx *ServiceSpecContext) {}

// ExitServiceSpec is called when production serviceSpec is exited.
func (s *BaseCztctlParserListener) ExitServiceSpec(ctx *ServiceSpecContext) {}

// EnterAtServer is called when production atServer is entered.
func (s *BaseCztctlParserListener) EnterAtServer(ctx *AtServerContext) {}

// ExitAtServer is called when production atServer is exited.
func (s *BaseCztctlParserListener) ExitAtServer(ctx *AtServerContext) {}

// EnterServiceApi is called when production serviceApi is entered.
func (s *BaseCztctlParserListener) EnterServiceApi(ctx *ServiceApiContext) {}

// ExitServiceApi is called when production serviceApi is exited.
func (s *BaseCztctlParserListener) ExitServiceApi(ctx *ServiceApiContext) {}

// EnterServiceName is called when production serviceName is entered.
func (s *BaseCztctlParserListener) EnterServiceName(ctx *ServiceNameContext) {}

// ExitServiceName is called when production serviceName is exited.
func (s *BaseCztctlParserListener) ExitServiceName(ctx *ServiceNameContext) {}

// EnterServiceRoute is called when production serviceRoute is entered.
func (s *BaseCztctlParserListener) EnterServiceRoute(ctx *ServiceRouteContext) {}

// ExitServiceRoute is called when production serviceRoute is exited.
func (s *BaseCztctlParserListener) ExitServiceRoute(ctx *ServiceRouteContext) {}

// EnterAtDoc is called when production atDoc is entered.
func (s *BaseCztctlParserListener) EnterAtDoc(ctx *AtDocContext) {}

// ExitAtDoc is called when production atDoc is exited.
func (s *BaseCztctlParserListener) ExitAtDoc(ctx *AtDocContext) {}

// EnterAtHandler is called when production atHandler is entered.
func (s *BaseCztctlParserListener) EnterAtHandler(ctx *AtHandlerContext) {}

// ExitAtHandler is called when production atHandler is exited.
func (s *BaseCztctlParserListener) ExitAtHandler(ctx *AtHandlerContext) {}

// EnterAtCron is called when production atCron is entered.
func (s *BaseCztctlParserListener) EnterAtCron(ctx *AtCronContext) {}

// ExitAtCron is called when production atCron is exited.
func (s *BaseCztctlParserListener) ExitAtCron(ctx *AtCronContext) {}

// EnterAtCronRetry is called when production atCronRetry is entered.
func (s *BaseCztctlParserListener) EnterAtCronRetry(ctx *AtCronRetryContext) {}

// ExitAtCronRetry is called when production atCronRetry is exited.
func (s *BaseCztctlParserListener) ExitAtCronRetry(ctx *AtCronRetryContext) {}

// EnterRoute is called when production route is entered.
func (s *BaseCztctlParserListener) EnterRoute(ctx *RouteContext) {}

// ExitRoute is called when production route is exited.
func (s *BaseCztctlParserListener) ExitRoute(ctx *RouteContext) {}

// EnterRouteName is called when production routeName is entered.
func (s *BaseCztctlParserListener) EnterRouteName(ctx *RouteNameContext) {}

// ExitRouteName is called when production routeName is exited.
func (s *BaseCztctlParserListener) ExitRouteName(ctx *RouteNameContext) {}

// EnterBody is called when production body is entered.
func (s *BaseCztctlParserListener) EnterBody(ctx *BodyContext) {}

// ExitBody is called when production body is exited.
func (s *BaseCztctlParserListener) ExitBody(ctx *BodyContext) {}

// EnterKvLit is called when production kvLit is entered.
func (s *BaseCztctlParserListener) EnterKvLit(ctx *KvLitContext) {}

// ExitKvLit is called when production kvLit is exited.
func (s *BaseCztctlParserListener) ExitKvLit(ctx *KvLitContext) {}

// EnterKvValue is called when production kvValue is entered.
func (s *BaseCztctlParserListener) EnterKvValue(ctx *KvValueContext) {}

// ExitKvValue is called when production kvValue is exited.
func (s *BaseCztctlParserListener) ExitKvValue(ctx *KvValueContext) {}
