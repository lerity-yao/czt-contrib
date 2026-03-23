// Code generated from CztctlParser.g4 by ANTLR 4.7.2. DO NOT EDIT.

package cztctl // CztctlParser
import "github.com/zeromicro/antlr"

type BaseCztctlParserVisitor struct {
	*antlr.BaseParseTreeVisitor
}

func (v *BaseCztctlParserVisitor) VisitApi(ctx *ApiContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCztctlParserVisitor) VisitSpec(ctx *SpecContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCztctlParserVisitor) VisitSyntaxLit(ctx *SyntaxLitContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCztctlParserVisitor) VisitImportSpec(ctx *ImportSpecContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCztctlParserVisitor) VisitImportLit(ctx *ImportLitContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCztctlParserVisitor) VisitImportBlock(ctx *ImportBlockContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCztctlParserVisitor) VisitImportBlockValue(ctx *ImportBlockValueContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCztctlParserVisitor) VisitImportValue(ctx *ImportValueContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCztctlParserVisitor) VisitInfoSpec(ctx *InfoSpecContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCztctlParserVisitor) VisitTypeSpec(ctx *TypeSpecContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCztctlParserVisitor) VisitTypeLit(ctx *TypeLitContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCztctlParserVisitor) VisitTypeBlock(ctx *TypeBlockContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCztctlParserVisitor) VisitTypeLitBody(ctx *TypeLitBodyContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCztctlParserVisitor) VisitTypeBlockBody(ctx *TypeBlockBodyContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCztctlParserVisitor) VisitTypeStruct(ctx *TypeStructContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCztctlParserVisitor) VisitTypeAlias(ctx *TypeAliasContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCztctlParserVisitor) VisitTypeBlockStruct(ctx *TypeBlockStructContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCztctlParserVisitor) VisitTypeBlockAlias(ctx *TypeBlockAliasContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCztctlParserVisitor) VisitField(ctx *FieldContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCztctlParserVisitor) VisitNormalField(ctx *NormalFieldContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCztctlParserVisitor) VisitAnonymousFiled(ctx *AnonymousFiledContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCztctlParserVisitor) VisitDataType(ctx *DataTypeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCztctlParserVisitor) VisitPointerType(ctx *PointerTypeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCztctlParserVisitor) VisitMapType(ctx *MapTypeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCztctlParserVisitor) VisitArrayType(ctx *ArrayTypeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCztctlParserVisitor) VisitServiceSpec(ctx *ServiceSpecContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCztctlParserVisitor) VisitAtServer(ctx *AtServerContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCztctlParserVisitor) VisitServiceApi(ctx *ServiceApiContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCztctlParserVisitor) VisitServiceName(ctx *ServiceNameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCztctlParserVisitor) VisitServiceRoute(ctx *ServiceRouteContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCztctlParserVisitor) VisitAtDoc(ctx *AtDocContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCztctlParserVisitor) VisitAtHandler(ctx *AtHandlerContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCztctlParserVisitor) VisitAtCron(ctx *AtCronContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCztctlParserVisitor) VisitAtCronRetry(ctx *AtCronRetryContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCztctlParserVisitor) VisitRoute(ctx *RouteContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCztctlParserVisitor) VisitRouteName(ctx *RouteNameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCztctlParserVisitor) VisitBody(ctx *BodyContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCztctlParserVisitor) VisitKvLit(ctx *KvLitContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseCztctlParserVisitor) VisitKvValue(ctx *KvValueContext) interface{} {
	return v.VisitChildren(ctx)
}
