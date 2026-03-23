// Code generated from CztctlParser.g4 by ANTLR 4.7.2. DO NOT EDIT.

package cztctl // CztctlParser
import "github.com/zeromicro/antlr"

// CztctlParserListener is a complete listener for a parse tree produced by CztctlParserParser.
type CztctlParserListener interface {
	antlr.ParseTreeListener

	// EnterApi is called when entering the api production.
	EnterApi(c *ApiContext)

	// EnterSpec is called when entering the spec production.
	EnterSpec(c *SpecContext)

	// EnterSyntaxLit is called when entering the syntaxLit production.
	EnterSyntaxLit(c *SyntaxLitContext)

	// EnterImportSpec is called when entering the importSpec production.
	EnterImportSpec(c *ImportSpecContext)

	// EnterImportLit is called when entering the importLit production.
	EnterImportLit(c *ImportLitContext)

	// EnterImportBlock is called when entering the importBlock production.
	EnterImportBlock(c *ImportBlockContext)

	// EnterImportBlockValue is called when entering the importBlockValue production.
	EnterImportBlockValue(c *ImportBlockValueContext)

	// EnterImportValue is called when entering the importValue production.
	EnterImportValue(c *ImportValueContext)

	// EnterInfoSpec is called when entering the infoSpec production.
	EnterInfoSpec(c *InfoSpecContext)

	// EnterTypeSpec is called when entering the typeSpec production.
	EnterTypeSpec(c *TypeSpecContext)

	// EnterTypeLit is called when entering the typeLit production.
	EnterTypeLit(c *TypeLitContext)

	// EnterTypeBlock is called when entering the typeBlock production.
	EnterTypeBlock(c *TypeBlockContext)

	// EnterTypeLitBody is called when entering the typeLitBody production.
	EnterTypeLitBody(c *TypeLitBodyContext)

	// EnterTypeBlockBody is called when entering the typeBlockBody production.
	EnterTypeBlockBody(c *TypeBlockBodyContext)

	// EnterTypeStruct is called when entering the typeStruct production.
	EnterTypeStruct(c *TypeStructContext)

	// EnterTypeAlias is called when entering the typeAlias production.
	EnterTypeAlias(c *TypeAliasContext)

	// EnterTypeBlockStruct is called when entering the typeBlockStruct production.
	EnterTypeBlockStruct(c *TypeBlockStructContext)

	// EnterTypeBlockAlias is called when entering the typeBlockAlias production.
	EnterTypeBlockAlias(c *TypeBlockAliasContext)

	// EnterField is called when entering the field production.
	EnterField(c *FieldContext)

	// EnterNormalField is called when entering the normalField production.
	EnterNormalField(c *NormalFieldContext)

	// EnterAnonymousFiled is called when entering the anonymousFiled production.
	EnterAnonymousFiled(c *AnonymousFiledContext)

	// EnterDataType is called when entering the dataType production.
	EnterDataType(c *DataTypeContext)

	// EnterPointerType is called when entering the pointerType production.
	EnterPointerType(c *PointerTypeContext)

	// EnterMapType is called when entering the mapType production.
	EnterMapType(c *MapTypeContext)

	// EnterArrayType is called when entering the arrayType production.
	EnterArrayType(c *ArrayTypeContext)

	// EnterServiceSpec is called when entering the serviceSpec production.
	EnterServiceSpec(c *ServiceSpecContext)

	// EnterAtServer is called when entering the atServer production.
	EnterAtServer(c *AtServerContext)

	// EnterServiceApi is called when entering the serviceApi production.
	EnterServiceApi(c *ServiceApiContext)

	// EnterServiceName is called when entering the serviceName production.
	EnterServiceName(c *ServiceNameContext)

	// EnterServiceRoute is called when entering the serviceRoute production.
	EnterServiceRoute(c *ServiceRouteContext)

	// EnterAtDoc is called when entering the atDoc production.
	EnterAtDoc(c *AtDocContext)

	// EnterAtHandler is called when entering the atHandler production.
	EnterAtHandler(c *AtHandlerContext)

	// EnterAtCron is called when entering the atCron production.
	EnterAtCron(c *AtCronContext)

	// EnterAtCronRetry is called when entering the atCronRetry production.
	EnterAtCronRetry(c *AtCronRetryContext)

	// EnterRoute is called when entering the route production.
	EnterRoute(c *RouteContext)

	// EnterRouteName is called when entering the routeName production.
	EnterRouteName(c *RouteNameContext)

	// EnterBody is called when entering the body production.
	EnterBody(c *BodyContext)

	// EnterKvLit is called when entering the kvLit production.
	EnterKvLit(c *KvLitContext)

	// EnterKvValue is called when entering the kvValue production.
	EnterKvValue(c *KvValueContext)

	// ExitApi is called when exiting the api production.
	ExitApi(c *ApiContext)

	// ExitSpec is called when exiting the spec production.
	ExitSpec(c *SpecContext)

	// ExitSyntaxLit is called when exiting the syntaxLit production.
	ExitSyntaxLit(c *SyntaxLitContext)

	// ExitImportSpec is called when exiting the importSpec production.
	ExitImportSpec(c *ImportSpecContext)

	// ExitImportLit is called when exiting the importLit production.
	ExitImportLit(c *ImportLitContext)

	// ExitImportBlock is called when exiting the importBlock production.
	ExitImportBlock(c *ImportBlockContext)

	// ExitImportBlockValue is called when exiting the importBlockValue production.
	ExitImportBlockValue(c *ImportBlockValueContext)

	// ExitImportValue is called when exiting the importValue production.
	ExitImportValue(c *ImportValueContext)

	// ExitInfoSpec is called when exiting the infoSpec production.
	ExitInfoSpec(c *InfoSpecContext)

	// ExitTypeSpec is called when exiting the typeSpec production.
	ExitTypeSpec(c *TypeSpecContext)

	// ExitTypeLit is called when exiting the typeLit production.
	ExitTypeLit(c *TypeLitContext)

	// ExitTypeBlock is called when exiting the typeBlock production.
	ExitTypeBlock(c *TypeBlockContext)

	// ExitTypeLitBody is called when exiting the typeLitBody production.
	ExitTypeLitBody(c *TypeLitBodyContext)

	// ExitTypeBlockBody is called when exiting the typeBlockBody production.
	ExitTypeBlockBody(c *TypeBlockBodyContext)

	// ExitTypeStruct is called when exiting the typeStruct production.
	ExitTypeStruct(c *TypeStructContext)

	// ExitTypeAlias is called when exiting the typeAlias production.
	ExitTypeAlias(c *TypeAliasContext)

	// ExitTypeBlockStruct is called when exiting the typeBlockStruct production.
	ExitTypeBlockStruct(c *TypeBlockStructContext)

	// ExitTypeBlockAlias is called when exiting the typeBlockAlias production.
	ExitTypeBlockAlias(c *TypeBlockAliasContext)

	// ExitField is called when exiting the field production.
	ExitField(c *FieldContext)

	// ExitNormalField is called when exiting the normalField production.
	ExitNormalField(c *NormalFieldContext)

	// ExitAnonymousFiled is called when exiting the anonymousFiled production.
	ExitAnonymousFiled(c *AnonymousFiledContext)

	// ExitDataType is called when exiting the dataType production.
	ExitDataType(c *DataTypeContext)

	// ExitPointerType is called when exiting the pointerType production.
	ExitPointerType(c *PointerTypeContext)

	// ExitMapType is called when exiting the mapType production.
	ExitMapType(c *MapTypeContext)

	// ExitArrayType is called when exiting the arrayType production.
	ExitArrayType(c *ArrayTypeContext)

	// ExitServiceSpec is called when exiting the serviceSpec production.
	ExitServiceSpec(c *ServiceSpecContext)

	// ExitAtServer is called when exiting the atServer production.
	ExitAtServer(c *AtServerContext)

	// ExitServiceApi is called when exiting the serviceApi production.
	ExitServiceApi(c *ServiceApiContext)

	// ExitServiceName is called when exiting the serviceName production.
	ExitServiceName(c *ServiceNameContext)

	// ExitServiceRoute is called when exiting the serviceRoute production.
	ExitServiceRoute(c *ServiceRouteContext)

	// ExitAtDoc is called when exiting the atDoc production.
	ExitAtDoc(c *AtDocContext)

	// ExitAtHandler is called when exiting the atHandler production.
	ExitAtHandler(c *AtHandlerContext)

	// ExitAtCron is called when exiting the atCron production.
	ExitAtCron(c *AtCronContext)

	// ExitAtCronRetry is called when exiting the atCronRetry production.
	ExitAtCronRetry(c *AtCronRetryContext)

	// ExitRoute is called when exiting the route production.
	ExitRoute(c *RouteContext)

	// ExitRouteName is called when exiting the routeName production.
	ExitRouteName(c *RouteNameContext)

	// ExitBody is called when exiting the body production.
	ExitBody(c *BodyContext)

	// ExitKvLit is called when exiting the kvLit production.
	ExitKvLit(c *KvLitContext)

	// ExitKvValue is called when exiting the kvValue production.
	ExitKvValue(c *KvValueContext)
}
