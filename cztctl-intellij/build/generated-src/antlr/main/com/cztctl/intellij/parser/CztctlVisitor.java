// Generated from com/cztctl/intellij/parser/Cztctl.g4 by ANTLR 4.13.2
package com.cztctl.intellij.parser;
import org.antlr.v4.runtime.tree.ParseTreeVisitor;

/**
 * This interface defines a complete generic visitor for a parse tree produced
 * by {@link CztctlParser}.
 *
 * @param <T> The return type of the visit operation. Use {@link Void} for
 * operations with no return type.
 */
public interface CztctlVisitor<T> extends ParseTreeVisitor<T> {
	/**
	 * Visit a parse tree produced by {@link CztctlParser#api}.
	 * @param ctx the parse tree
	 * @return the visitor result
	 */
	T visitApi(CztctlParser.ApiContext ctx);
	/**
	 * Visit a parse tree produced by {@link CztctlParser#spec}.
	 * @param ctx the parse tree
	 * @return the visitor result
	 */
	T visitSpec(CztctlParser.SpecContext ctx);
	/**
	 * Visit a parse tree produced by {@link CztctlParser#syntaxLit}.
	 * @param ctx the parse tree
	 * @return the visitor result
	 */
	T visitSyntaxLit(CztctlParser.SyntaxLitContext ctx);
	/**
	 * Visit a parse tree produced by {@link CztctlParser#importSpec}.
	 * @param ctx the parse tree
	 * @return the visitor result
	 */
	T visitImportSpec(CztctlParser.ImportSpecContext ctx);
	/**
	 * Visit a parse tree produced by {@link CztctlParser#importLit}.
	 * @param ctx the parse tree
	 * @return the visitor result
	 */
	T visitImportLit(CztctlParser.ImportLitContext ctx);
	/**
	 * Visit a parse tree produced by {@link CztctlParser#importBlock}.
	 * @param ctx the parse tree
	 * @return the visitor result
	 */
	T visitImportBlock(CztctlParser.ImportBlockContext ctx);
	/**
	 * Visit a parse tree produced by {@link CztctlParser#importBlockValue}.
	 * @param ctx the parse tree
	 * @return the visitor result
	 */
	T visitImportBlockValue(CztctlParser.ImportBlockValueContext ctx);
	/**
	 * Visit a parse tree produced by {@link CztctlParser#importValue}.
	 * @param ctx the parse tree
	 * @return the visitor result
	 */
	T visitImportValue(CztctlParser.ImportValueContext ctx);
	/**
	 * Visit a parse tree produced by {@link CztctlParser#infoSpec}.
	 * @param ctx the parse tree
	 * @return the visitor result
	 */
	T visitInfoSpec(CztctlParser.InfoSpecContext ctx);
	/**
	 * Visit a parse tree produced by {@link CztctlParser#typeSpec}.
	 * @param ctx the parse tree
	 * @return the visitor result
	 */
	T visitTypeSpec(CztctlParser.TypeSpecContext ctx);
	/**
	 * Visit a parse tree produced by {@link CztctlParser#typeLit}.
	 * @param ctx the parse tree
	 * @return the visitor result
	 */
	T visitTypeLit(CztctlParser.TypeLitContext ctx);
	/**
	 * Visit a parse tree produced by {@link CztctlParser#typeBlock}.
	 * @param ctx the parse tree
	 * @return the visitor result
	 */
	T visitTypeBlock(CztctlParser.TypeBlockContext ctx);
	/**
	 * Visit a parse tree produced by {@link CztctlParser#typeLitBody}.
	 * @param ctx the parse tree
	 * @return the visitor result
	 */
	T visitTypeLitBody(CztctlParser.TypeLitBodyContext ctx);
	/**
	 * Visit a parse tree produced by {@link CztctlParser#typeBlockBody}.
	 * @param ctx the parse tree
	 * @return the visitor result
	 */
	T visitTypeBlockBody(CztctlParser.TypeBlockBodyContext ctx);
	/**
	 * Visit a parse tree produced by {@link CztctlParser#typeStruct}.
	 * @param ctx the parse tree
	 * @return the visitor result
	 */
	T visitTypeStruct(CztctlParser.TypeStructContext ctx);
	/**
	 * Visit a parse tree produced by {@link CztctlParser#typeAlias}.
	 * @param ctx the parse tree
	 * @return the visitor result
	 */
	T visitTypeAlias(CztctlParser.TypeAliasContext ctx);
	/**
	 * Visit a parse tree produced by {@link CztctlParser#typeBlockStruct}.
	 * @param ctx the parse tree
	 * @return the visitor result
	 */
	T visitTypeBlockStruct(CztctlParser.TypeBlockStructContext ctx);
	/**
	 * Visit a parse tree produced by {@link CztctlParser#typeBlockAlias}.
	 * @param ctx the parse tree
	 * @return the visitor result
	 */
	T visitTypeBlockAlias(CztctlParser.TypeBlockAliasContext ctx);
	/**
	 * Visit a parse tree produced by {@link CztctlParser#field}.
	 * @param ctx the parse tree
	 * @return the visitor result
	 */
	T visitField(CztctlParser.FieldContext ctx);
	/**
	 * Visit a parse tree produced by {@link CztctlParser#normalField}.
	 * @param ctx the parse tree
	 * @return the visitor result
	 */
	T visitNormalField(CztctlParser.NormalFieldContext ctx);
	/**
	 * Visit a parse tree produced by {@link CztctlParser#anonymousField}.
	 * @param ctx the parse tree
	 * @return the visitor result
	 */
	T visitAnonymousField(CztctlParser.AnonymousFieldContext ctx);
	/**
	 * Visit a parse tree produced by {@link CztctlParser#dataType}.
	 * @param ctx the parse tree
	 * @return the visitor result
	 */
	T visitDataType(CztctlParser.DataTypeContext ctx);
	/**
	 * Visit a parse tree produced by {@link CztctlParser#pointerType}.
	 * @param ctx the parse tree
	 * @return the visitor result
	 */
	T visitPointerType(CztctlParser.PointerTypeContext ctx);
	/**
	 * Visit a parse tree produced by {@link CztctlParser#mapType}.
	 * @param ctx the parse tree
	 * @return the visitor result
	 */
	T visitMapType(CztctlParser.MapTypeContext ctx);
	/**
	 * Visit a parse tree produced by {@link CztctlParser#arrayType}.
	 * @param ctx the parse tree
	 * @return the visitor result
	 */
	T visitArrayType(CztctlParser.ArrayTypeContext ctx);
	/**
	 * Visit a parse tree produced by {@link CztctlParser#serviceSpec}.
	 * @param ctx the parse tree
	 * @return the visitor result
	 */
	T visitServiceSpec(CztctlParser.ServiceSpecContext ctx);
	/**
	 * Visit a parse tree produced by {@link CztctlParser#atServer}.
	 * @param ctx the parse tree
	 * @return the visitor result
	 */
	T visitAtServer(CztctlParser.AtServerContext ctx);
	/**
	 * Visit a parse tree produced by {@link CztctlParser#serviceApi}.
	 * @param ctx the parse tree
	 * @return the visitor result
	 */
	T visitServiceApi(CztctlParser.ServiceApiContext ctx);
	/**
	 * Visit a parse tree produced by {@link CztctlParser#serviceName}.
	 * @param ctx the parse tree
	 * @return the visitor result
	 */
	T visitServiceName(CztctlParser.ServiceNameContext ctx);
	/**
	 * Visit a parse tree produced by {@link CztctlParser#serviceRoute}.
	 * @param ctx the parse tree
	 * @return the visitor result
	 */
	T visitServiceRoute(CztctlParser.ServiceRouteContext ctx);
	/**
	 * Visit a parse tree produced by {@link CztctlParser#atDoc}.
	 * @param ctx the parse tree
	 * @return the visitor result
	 */
	T visitAtDoc(CztctlParser.AtDocContext ctx);
	/**
	 * Visit a parse tree produced by {@link CztctlParser#atHandler}.
	 * @param ctx the parse tree
	 * @return the visitor result
	 */
	T visitAtHandler(CztctlParser.AtHandlerContext ctx);
	/**
	 * Visit a parse tree produced by {@link CztctlParser#atCron}.
	 * @param ctx the parse tree
	 * @return the visitor result
	 */
	T visitAtCron(CztctlParser.AtCronContext ctx);
	/**
	 * Visit a parse tree produced by {@link CztctlParser#atCronRetry}.
	 * @param ctx the parse tree
	 * @return the visitor result
	 */
	T visitAtCronRetry(CztctlParser.AtCronRetryContext ctx);
	/**
	 * Visit a parse tree produced by {@link CztctlParser#route}.
	 * @param ctx the parse tree
	 * @return the visitor result
	 */
	T visitRoute(CztctlParser.RouteContext ctx);
	/**
	 * Visit a parse tree produced by {@link CztctlParser#routeName}.
	 * @param ctx the parse tree
	 * @return the visitor result
	 */
	T visitRouteName(CztctlParser.RouteNameContext ctx);
	/**
	 * Visit a parse tree produced by {@link CztctlParser#body}.
	 * @param ctx the parse tree
	 * @return the visitor result
	 */
	T visitBody(CztctlParser.BodyContext ctx);
	/**
	 * Visit a parse tree produced by {@link CztctlParser#kvLit}.
	 * @param ctx the parse tree
	 * @return the visitor result
	 */
	T visitKvLit(CztctlParser.KvLitContext ctx);
	/**
	 * Visit a parse tree produced by {@link CztctlParser#kvValue}.
	 * @param ctx the parse tree
	 * @return the visitor result
	 */
	T visitKvValue(CztctlParser.KvValueContext ctx);
	/**
	 * Visit a parse tree produced by {@link CztctlParser#identifier}.
	 * @param ctx the parse tree
	 * @return the visitor result
	 */
	T visitIdentifier(CztctlParser.IdentifierContext ctx);
}