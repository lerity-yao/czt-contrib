// Generated from com/cztctl/intellij/parser/Cztctl.g4 by ANTLR 4.13.2
package com.cztctl.intellij.parser;
import org.antlr.v4.runtime.tree.ParseTreeListener;

/**
 * This interface defines a complete listener for a parse tree produced by
 * {@link CztctlParser}.
 */
public interface CztctlListener extends ParseTreeListener {
	/**
	 * Enter a parse tree produced by {@link CztctlParser#api}.
	 * @param ctx the parse tree
	 */
	void enterApi(CztctlParser.ApiContext ctx);
	/**
	 * Exit a parse tree produced by {@link CztctlParser#api}.
	 * @param ctx the parse tree
	 */
	void exitApi(CztctlParser.ApiContext ctx);
	/**
	 * Enter a parse tree produced by {@link CztctlParser#spec}.
	 * @param ctx the parse tree
	 */
	void enterSpec(CztctlParser.SpecContext ctx);
	/**
	 * Exit a parse tree produced by {@link CztctlParser#spec}.
	 * @param ctx the parse tree
	 */
	void exitSpec(CztctlParser.SpecContext ctx);
	/**
	 * Enter a parse tree produced by {@link CztctlParser#syntaxLit}.
	 * @param ctx the parse tree
	 */
	void enterSyntaxLit(CztctlParser.SyntaxLitContext ctx);
	/**
	 * Exit a parse tree produced by {@link CztctlParser#syntaxLit}.
	 * @param ctx the parse tree
	 */
	void exitSyntaxLit(CztctlParser.SyntaxLitContext ctx);
	/**
	 * Enter a parse tree produced by {@link CztctlParser#importSpec}.
	 * @param ctx the parse tree
	 */
	void enterImportSpec(CztctlParser.ImportSpecContext ctx);
	/**
	 * Exit a parse tree produced by {@link CztctlParser#importSpec}.
	 * @param ctx the parse tree
	 */
	void exitImportSpec(CztctlParser.ImportSpecContext ctx);
	/**
	 * Enter a parse tree produced by {@link CztctlParser#importLit}.
	 * @param ctx the parse tree
	 */
	void enterImportLit(CztctlParser.ImportLitContext ctx);
	/**
	 * Exit a parse tree produced by {@link CztctlParser#importLit}.
	 * @param ctx the parse tree
	 */
	void exitImportLit(CztctlParser.ImportLitContext ctx);
	/**
	 * Enter a parse tree produced by {@link CztctlParser#importBlock}.
	 * @param ctx the parse tree
	 */
	void enterImportBlock(CztctlParser.ImportBlockContext ctx);
	/**
	 * Exit a parse tree produced by {@link CztctlParser#importBlock}.
	 * @param ctx the parse tree
	 */
	void exitImportBlock(CztctlParser.ImportBlockContext ctx);
	/**
	 * Enter a parse tree produced by {@link CztctlParser#importBlockValue}.
	 * @param ctx the parse tree
	 */
	void enterImportBlockValue(CztctlParser.ImportBlockValueContext ctx);
	/**
	 * Exit a parse tree produced by {@link CztctlParser#importBlockValue}.
	 * @param ctx the parse tree
	 */
	void exitImportBlockValue(CztctlParser.ImportBlockValueContext ctx);
	/**
	 * Enter a parse tree produced by {@link CztctlParser#importValue}.
	 * @param ctx the parse tree
	 */
	void enterImportValue(CztctlParser.ImportValueContext ctx);
	/**
	 * Exit a parse tree produced by {@link CztctlParser#importValue}.
	 * @param ctx the parse tree
	 */
	void exitImportValue(CztctlParser.ImportValueContext ctx);
	/**
	 * Enter a parse tree produced by {@link CztctlParser#infoSpec}.
	 * @param ctx the parse tree
	 */
	void enterInfoSpec(CztctlParser.InfoSpecContext ctx);
	/**
	 * Exit a parse tree produced by {@link CztctlParser#infoSpec}.
	 * @param ctx the parse tree
	 */
	void exitInfoSpec(CztctlParser.InfoSpecContext ctx);
	/**
	 * Enter a parse tree produced by {@link CztctlParser#typeSpec}.
	 * @param ctx the parse tree
	 */
	void enterTypeSpec(CztctlParser.TypeSpecContext ctx);
	/**
	 * Exit a parse tree produced by {@link CztctlParser#typeSpec}.
	 * @param ctx the parse tree
	 */
	void exitTypeSpec(CztctlParser.TypeSpecContext ctx);
	/**
	 * Enter a parse tree produced by {@link CztctlParser#typeLit}.
	 * @param ctx the parse tree
	 */
	void enterTypeLit(CztctlParser.TypeLitContext ctx);
	/**
	 * Exit a parse tree produced by {@link CztctlParser#typeLit}.
	 * @param ctx the parse tree
	 */
	void exitTypeLit(CztctlParser.TypeLitContext ctx);
	/**
	 * Enter a parse tree produced by {@link CztctlParser#typeBlock}.
	 * @param ctx the parse tree
	 */
	void enterTypeBlock(CztctlParser.TypeBlockContext ctx);
	/**
	 * Exit a parse tree produced by {@link CztctlParser#typeBlock}.
	 * @param ctx the parse tree
	 */
	void exitTypeBlock(CztctlParser.TypeBlockContext ctx);
	/**
	 * Enter a parse tree produced by {@link CztctlParser#typeLitBody}.
	 * @param ctx the parse tree
	 */
	void enterTypeLitBody(CztctlParser.TypeLitBodyContext ctx);
	/**
	 * Exit a parse tree produced by {@link CztctlParser#typeLitBody}.
	 * @param ctx the parse tree
	 */
	void exitTypeLitBody(CztctlParser.TypeLitBodyContext ctx);
	/**
	 * Enter a parse tree produced by {@link CztctlParser#typeBlockBody}.
	 * @param ctx the parse tree
	 */
	void enterTypeBlockBody(CztctlParser.TypeBlockBodyContext ctx);
	/**
	 * Exit a parse tree produced by {@link CztctlParser#typeBlockBody}.
	 * @param ctx the parse tree
	 */
	void exitTypeBlockBody(CztctlParser.TypeBlockBodyContext ctx);
	/**
	 * Enter a parse tree produced by {@link CztctlParser#typeStruct}.
	 * @param ctx the parse tree
	 */
	void enterTypeStruct(CztctlParser.TypeStructContext ctx);
	/**
	 * Exit a parse tree produced by {@link CztctlParser#typeStruct}.
	 * @param ctx the parse tree
	 */
	void exitTypeStruct(CztctlParser.TypeStructContext ctx);
	/**
	 * Enter a parse tree produced by {@link CztctlParser#typeAlias}.
	 * @param ctx the parse tree
	 */
	void enterTypeAlias(CztctlParser.TypeAliasContext ctx);
	/**
	 * Exit a parse tree produced by {@link CztctlParser#typeAlias}.
	 * @param ctx the parse tree
	 */
	void exitTypeAlias(CztctlParser.TypeAliasContext ctx);
	/**
	 * Enter a parse tree produced by {@link CztctlParser#typeBlockStruct}.
	 * @param ctx the parse tree
	 */
	void enterTypeBlockStruct(CztctlParser.TypeBlockStructContext ctx);
	/**
	 * Exit a parse tree produced by {@link CztctlParser#typeBlockStruct}.
	 * @param ctx the parse tree
	 */
	void exitTypeBlockStruct(CztctlParser.TypeBlockStructContext ctx);
	/**
	 * Enter a parse tree produced by {@link CztctlParser#typeBlockAlias}.
	 * @param ctx the parse tree
	 */
	void enterTypeBlockAlias(CztctlParser.TypeBlockAliasContext ctx);
	/**
	 * Exit a parse tree produced by {@link CztctlParser#typeBlockAlias}.
	 * @param ctx the parse tree
	 */
	void exitTypeBlockAlias(CztctlParser.TypeBlockAliasContext ctx);
	/**
	 * Enter a parse tree produced by {@link CztctlParser#field}.
	 * @param ctx the parse tree
	 */
	void enterField(CztctlParser.FieldContext ctx);
	/**
	 * Exit a parse tree produced by {@link CztctlParser#field}.
	 * @param ctx the parse tree
	 */
	void exitField(CztctlParser.FieldContext ctx);
	/**
	 * Enter a parse tree produced by {@link CztctlParser#normalField}.
	 * @param ctx the parse tree
	 */
	void enterNormalField(CztctlParser.NormalFieldContext ctx);
	/**
	 * Exit a parse tree produced by {@link CztctlParser#normalField}.
	 * @param ctx the parse tree
	 */
	void exitNormalField(CztctlParser.NormalFieldContext ctx);
	/**
	 * Enter a parse tree produced by {@link CztctlParser#anonymousField}.
	 * @param ctx the parse tree
	 */
	void enterAnonymousField(CztctlParser.AnonymousFieldContext ctx);
	/**
	 * Exit a parse tree produced by {@link CztctlParser#anonymousField}.
	 * @param ctx the parse tree
	 */
	void exitAnonymousField(CztctlParser.AnonymousFieldContext ctx);
	/**
	 * Enter a parse tree produced by {@link CztctlParser#dataType}.
	 * @param ctx the parse tree
	 */
	void enterDataType(CztctlParser.DataTypeContext ctx);
	/**
	 * Exit a parse tree produced by {@link CztctlParser#dataType}.
	 * @param ctx the parse tree
	 */
	void exitDataType(CztctlParser.DataTypeContext ctx);
	/**
	 * Enter a parse tree produced by {@link CztctlParser#pointerType}.
	 * @param ctx the parse tree
	 */
	void enterPointerType(CztctlParser.PointerTypeContext ctx);
	/**
	 * Exit a parse tree produced by {@link CztctlParser#pointerType}.
	 * @param ctx the parse tree
	 */
	void exitPointerType(CztctlParser.PointerTypeContext ctx);
	/**
	 * Enter a parse tree produced by {@link CztctlParser#mapType}.
	 * @param ctx the parse tree
	 */
	void enterMapType(CztctlParser.MapTypeContext ctx);
	/**
	 * Exit a parse tree produced by {@link CztctlParser#mapType}.
	 * @param ctx the parse tree
	 */
	void exitMapType(CztctlParser.MapTypeContext ctx);
	/**
	 * Enter a parse tree produced by {@link CztctlParser#arrayType}.
	 * @param ctx the parse tree
	 */
	void enterArrayType(CztctlParser.ArrayTypeContext ctx);
	/**
	 * Exit a parse tree produced by {@link CztctlParser#arrayType}.
	 * @param ctx the parse tree
	 */
	void exitArrayType(CztctlParser.ArrayTypeContext ctx);
	/**
	 * Enter a parse tree produced by {@link CztctlParser#serviceSpec}.
	 * @param ctx the parse tree
	 */
	void enterServiceSpec(CztctlParser.ServiceSpecContext ctx);
	/**
	 * Exit a parse tree produced by {@link CztctlParser#serviceSpec}.
	 * @param ctx the parse tree
	 */
	void exitServiceSpec(CztctlParser.ServiceSpecContext ctx);
	/**
	 * Enter a parse tree produced by {@link CztctlParser#atServer}.
	 * @param ctx the parse tree
	 */
	void enterAtServer(CztctlParser.AtServerContext ctx);
	/**
	 * Exit a parse tree produced by {@link CztctlParser#atServer}.
	 * @param ctx the parse tree
	 */
	void exitAtServer(CztctlParser.AtServerContext ctx);
	/**
	 * Enter a parse tree produced by {@link CztctlParser#serviceApi}.
	 * @param ctx the parse tree
	 */
	void enterServiceApi(CztctlParser.ServiceApiContext ctx);
	/**
	 * Exit a parse tree produced by {@link CztctlParser#serviceApi}.
	 * @param ctx the parse tree
	 */
	void exitServiceApi(CztctlParser.ServiceApiContext ctx);
	/**
	 * Enter a parse tree produced by {@link CztctlParser#serviceName}.
	 * @param ctx the parse tree
	 */
	void enterServiceName(CztctlParser.ServiceNameContext ctx);
	/**
	 * Exit a parse tree produced by {@link CztctlParser#serviceName}.
	 * @param ctx the parse tree
	 */
	void exitServiceName(CztctlParser.ServiceNameContext ctx);
	/**
	 * Enter a parse tree produced by {@link CztctlParser#serviceRoute}.
	 * @param ctx the parse tree
	 */
	void enterServiceRoute(CztctlParser.ServiceRouteContext ctx);
	/**
	 * Exit a parse tree produced by {@link CztctlParser#serviceRoute}.
	 * @param ctx the parse tree
	 */
	void exitServiceRoute(CztctlParser.ServiceRouteContext ctx);
	/**
	 * Enter a parse tree produced by {@link CztctlParser#atDoc}.
	 * @param ctx the parse tree
	 */
	void enterAtDoc(CztctlParser.AtDocContext ctx);
	/**
	 * Exit a parse tree produced by {@link CztctlParser#atDoc}.
	 * @param ctx the parse tree
	 */
	void exitAtDoc(CztctlParser.AtDocContext ctx);
	/**
	 * Enter a parse tree produced by {@link CztctlParser#atHandler}.
	 * @param ctx the parse tree
	 */
	void enterAtHandler(CztctlParser.AtHandlerContext ctx);
	/**
	 * Exit a parse tree produced by {@link CztctlParser#atHandler}.
	 * @param ctx the parse tree
	 */
	void exitAtHandler(CztctlParser.AtHandlerContext ctx);
	/**
	 * Enter a parse tree produced by {@link CztctlParser#atCron}.
	 * @param ctx the parse tree
	 */
	void enterAtCron(CztctlParser.AtCronContext ctx);
	/**
	 * Exit a parse tree produced by {@link CztctlParser#atCron}.
	 * @param ctx the parse tree
	 */
	void exitAtCron(CztctlParser.AtCronContext ctx);
	/**
	 * Enter a parse tree produced by {@link CztctlParser#atCronRetry}.
	 * @param ctx the parse tree
	 */
	void enterAtCronRetry(CztctlParser.AtCronRetryContext ctx);
	/**
	 * Exit a parse tree produced by {@link CztctlParser#atCronRetry}.
	 * @param ctx the parse tree
	 */
	void exitAtCronRetry(CztctlParser.AtCronRetryContext ctx);
	/**
	 * Enter a parse tree produced by {@link CztctlParser#route}.
	 * @param ctx the parse tree
	 */
	void enterRoute(CztctlParser.RouteContext ctx);
	/**
	 * Exit a parse tree produced by {@link CztctlParser#route}.
	 * @param ctx the parse tree
	 */
	void exitRoute(CztctlParser.RouteContext ctx);
	/**
	 * Enter a parse tree produced by {@link CztctlParser#routeName}.
	 * @param ctx the parse tree
	 */
	void enterRouteName(CztctlParser.RouteNameContext ctx);
	/**
	 * Exit a parse tree produced by {@link CztctlParser#routeName}.
	 * @param ctx the parse tree
	 */
	void exitRouteName(CztctlParser.RouteNameContext ctx);
	/**
	 * Enter a parse tree produced by {@link CztctlParser#body}.
	 * @param ctx the parse tree
	 */
	void enterBody(CztctlParser.BodyContext ctx);
	/**
	 * Exit a parse tree produced by {@link CztctlParser#body}.
	 * @param ctx the parse tree
	 */
	void exitBody(CztctlParser.BodyContext ctx);
	/**
	 * Enter a parse tree produced by {@link CztctlParser#kvLit}.
	 * @param ctx the parse tree
	 */
	void enterKvLit(CztctlParser.KvLitContext ctx);
	/**
	 * Exit a parse tree produced by {@link CztctlParser#kvLit}.
	 * @param ctx the parse tree
	 */
	void exitKvLit(CztctlParser.KvLitContext ctx);
	/**
	 * Enter a parse tree produced by {@link CztctlParser#kvValue}.
	 * @param ctx the parse tree
	 */
	void enterKvValue(CztctlParser.KvValueContext ctx);
	/**
	 * Exit a parse tree produced by {@link CztctlParser#kvValue}.
	 * @param ctx the parse tree
	 */
	void exitKvValue(CztctlParser.KvValueContext ctx);
	/**
	 * Enter a parse tree produced by {@link CztctlParser#identifier}.
	 * @param ctx the parse tree
	 */
	void enterIdentifier(CztctlParser.IdentifierContext ctx);
	/**
	 * Exit a parse tree produced by {@link CztctlParser#identifier}.
	 * @param ctx the parse tree
	 */
	void exitIdentifier(CztctlParser.IdentifierContext ctx);
}