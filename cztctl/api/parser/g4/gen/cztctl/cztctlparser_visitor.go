// Code generated from CztctlParser.g4 by ANTLR 4.7.2. DO NOT EDIT.

package cztctl // CztctlParser
import "github.com/zeromicro/antlr"

// A complete Visitor for a parse tree produced by CztctlParserParser.
type CztctlParserVisitor interface {
	antlr.ParseTreeVisitor

	// Visit a parse tree produced by CztctlParserParser#api.
	VisitApi(ctx *ApiContext) interface{}

	// Visit a parse tree produced by CztctlParserParser#spec.
	VisitSpec(ctx *SpecContext) interface{}

	// Visit a parse tree produced by CztctlParserParser#syntaxLit.
	VisitSyntaxLit(ctx *SyntaxLitContext) interface{}

	// Visit a parse tree produced by CztctlParserParser#importSpec.
	VisitImportSpec(ctx *ImportSpecContext) interface{}

	// Visit a parse tree produced by CztctlParserParser#importLit.
	VisitImportLit(ctx *ImportLitContext) interface{}

	// Visit a parse tree produced by CztctlParserParser#importBlock.
	VisitImportBlock(ctx *ImportBlockContext) interface{}

	// Visit a parse tree produced by CztctlParserParser#importBlockValue.
	VisitImportBlockValue(ctx *ImportBlockValueContext) interface{}

	// Visit a parse tree produced by CztctlParserParser#importValue.
	VisitImportValue(ctx *ImportValueContext) interface{}

	// Visit a parse tree produced by CztctlParserParser#infoSpec.
	VisitInfoSpec(ctx *InfoSpecContext) interface{}

	// Visit a parse tree produced by CztctlParserParser#typeSpec.
	VisitTypeSpec(ctx *TypeSpecContext) interface{}

	// Visit a parse tree produced by CztctlParserParser#typeLit.
	VisitTypeLit(ctx *TypeLitContext) interface{}

	// Visit a parse tree produced by CztctlParserParser#typeBlock.
	VisitTypeBlock(ctx *TypeBlockContext) interface{}

	// Visit a parse tree produced by CztctlParserParser#typeLitBody.
	VisitTypeLitBody(ctx *TypeLitBodyContext) interface{}

	// Visit a parse tree produced by CztctlParserParser#typeBlockBody.
	VisitTypeBlockBody(ctx *TypeBlockBodyContext) interface{}

	// Visit a parse tree produced by CztctlParserParser#typeStruct.
	VisitTypeStruct(ctx *TypeStructContext) interface{}

	// Visit a parse tree produced by CztctlParserParser#typeAlias.
	VisitTypeAlias(ctx *TypeAliasContext) interface{}

	// Visit a parse tree produced by CztctlParserParser#typeBlockStruct.
	VisitTypeBlockStruct(ctx *TypeBlockStructContext) interface{}

	// Visit a parse tree produced by CztctlParserParser#typeBlockAlias.
	VisitTypeBlockAlias(ctx *TypeBlockAliasContext) interface{}

	// Visit a parse tree produced by CztctlParserParser#field.
	VisitField(ctx *FieldContext) interface{}

	// Visit a parse tree produced by CztctlParserParser#normalField.
	VisitNormalField(ctx *NormalFieldContext) interface{}

	// Visit a parse tree produced by CztctlParserParser#anonymousFiled.
	VisitAnonymousFiled(ctx *AnonymousFiledContext) interface{}

	// Visit a parse tree produced by CztctlParserParser#dataType.
	VisitDataType(ctx *DataTypeContext) interface{}

	// Visit a parse tree produced by CztctlParserParser#pointerType.
	VisitPointerType(ctx *PointerTypeContext) interface{}

	// Visit a parse tree produced by CztctlParserParser#mapType.
	VisitMapType(ctx *MapTypeContext) interface{}

	// Visit a parse tree produced by CztctlParserParser#arrayType.
	VisitArrayType(ctx *ArrayTypeContext) interface{}

	// Visit a parse tree produced by CztctlParserParser#serviceSpec.
	VisitServiceSpec(ctx *ServiceSpecContext) interface{}

	// Visit a parse tree produced by CztctlParserParser#atServer.
	VisitAtServer(ctx *AtServerContext) interface{}

	// Visit a parse tree produced by CztctlParserParser#serviceApi.
	VisitServiceApi(ctx *ServiceApiContext) interface{}

	// Visit a parse tree produced by CztctlParserParser#serviceName.
	VisitServiceName(ctx *ServiceNameContext) interface{}

	// Visit a parse tree produced by CztctlParserParser#serviceRoute.
	VisitServiceRoute(ctx *ServiceRouteContext) interface{}

	// Visit a parse tree produced by CztctlParserParser#atDoc.
	VisitAtDoc(ctx *AtDocContext) interface{}

	// Visit a parse tree produced by CztctlParserParser#atHandler.
	VisitAtHandler(ctx *AtHandlerContext) interface{}

	// Visit a parse tree produced by CztctlParserParser#atCron.
	VisitAtCron(ctx *AtCronContext) interface{}

	// Visit a parse tree produced by CztctlParserParser#atCronRetry.
	VisitAtCronRetry(ctx *AtCronRetryContext) interface{}

	// Visit a parse tree produced by CztctlParserParser#route.
	VisitRoute(ctx *RouteContext) interface{}

	// Visit a parse tree produced by CztctlParserParser#routeName.
	VisitRouteName(ctx *RouteNameContext) interface{}

	// Visit a parse tree produced by CztctlParserParser#body.
	VisitBody(ctx *BodyContext) interface{}

	// Visit a parse tree produced by CztctlParserParser#kvLit.
	VisitKvLit(ctx *KvLitContext) interface{}

	// Visit a parse tree produced by CztctlParserParser#kvValue.
	VisitKvValue(ctx *KvValueContext) interface{}
}
