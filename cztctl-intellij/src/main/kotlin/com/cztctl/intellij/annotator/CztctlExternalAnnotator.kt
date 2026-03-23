package com.cztctl.intellij.annotator

import com.cztctl.intellij.highlight.CztctlSyntaxHighlighter
import com.cztctl.intellij.parser.CztctlLexer
import com.cztctl.intellij.parser.CztctlParser
import com.intellij.lang.annotation.AnnotationHolder
import com.intellij.lang.annotation.ExternalAnnotator
import com.intellij.lang.annotation.HighlightSeverity
import com.intellij.openapi.editor.Editor
import com.intellij.openapi.editor.markup.TextAttributes
import com.intellij.openapi.util.TextRange
import com.intellij.psi.PsiFile
import org.antlr.v4.runtime.*
import org.antlr.v4.runtime.tree.ParseTreeWalker

/**
 * Runs ANTLR4 full parse on file content:
 * 1. Syntax errors → red squiggly lines
 * 2. Semantic validation → file-type-specific errors (.rabbitmq forbids @cron/@cronRetry)
 * 3. Semantic highlighting → context-dependent coloring (handler name, service name, route)
 */
class CztctlExternalAnnotator : ExternalAnnotator<CztctlExternalAnnotator.AnnotationInput, CztctlExternalAnnotator.AnnotationResult>() {

    data class AnnotationInput(
        val source: String,
        val fileExtension: String
    )

    data class AnnotationResult(
        val errors: List<SyntaxError>,
        val highlights: List<SemanticHighlight>
    )

    data class SyntaxError(
        val line: Int,
        val charPositionInLine: Int,
        val length: Int,
        val message: String,
        val severity: HighlightSeverity = HighlightSeverity.ERROR
    )

    data class SemanticHighlight(
        val startOffset: Int,
        val endOffset: Int,
        val attrs: TextAttributes
    )

    override fun collectInformation(file: PsiFile, editor: Editor, hasErrors: Boolean): AnnotationInput {
        val ext = file.virtualFile?.extension ?: file.name.substringAfterLast('.', "")
        return AnnotationInput(file.text, ext)
    }

    override fun doAnnotate(input: AnnotationInput): AnnotationResult {
        val errors = mutableListOf<SyntaxError>()
        val highlights = mutableListOf<SemanticHighlight>()
        val source = input.source

        val lexer = CztctlLexer(CharStreams.fromString(source))
        lexer.removeErrorListeners()
        lexer.addErrorListener(object : BaseErrorListener() {
            override fun syntaxError(
                recognizer: Recognizer<*, *>?,
                offendingSymbol: Any?,
                line: Int,
                charPositionInLine: Int,
                msg: String?,
                e: RecognitionException?
            ) {
                errors.add(SyntaxError(line, charPositionInLine, 1, msg ?: "Lexer error"))
            }
        })

        val tokens = CommonTokenStream(lexer)
        val parser = CztctlParser(tokens)
        parser.removeErrorListeners()
        parser.addErrorListener(object : BaseErrorListener() {
            override fun syntaxError(
                recognizer: Recognizer<*, *>?,
                offendingSymbol: Any?,
                line: Int,
                charPositionInLine: Int,
                msg: String?,
                e: RecognitionException?
            ) {
                val length = if (offendingSymbol is Token) {
                    maxOf(1, offendingSymbol.stopIndex - offendingSymbol.startIndex + 1)
                } else 1
                errors.add(SyntaxError(line, charPositionInLine, length, msg ?: "Syntax error"))
            }
        })

        val tree = parser.api()

        // Walk parse tree for semantic validation + highlighting
        val walker = ParseTreeWalker()
        walker.walk(object : com.cztctl.intellij.parser.CztctlBaseListener() {

            // ── Semantic errors ──

            override fun enterAtCron(ctx: CztctlParser.AtCronContext) {
                // @cron keyword color: #A5D5F4
                val kwToken = ctx.ATCRON()?.symbol
                if (kwToken != null) {
                    highlights.add(SemanticHighlight(kwToken.startIndex, kwToken.stopIndex + 1, CztctlSyntaxHighlighter.SERVICE_ANNOTATION_ATTRS))
                }
                // semantic error: forbid @cron in .rabbitmq files
                if (input.fileExtension == "rabbitmq") {
                    val token = ctx.start
                    errors.add(
                        SyntaxError(
                            token.line,
                            token.charPositionInLine,
                            ctx.stop.stopIndex - ctx.start.startIndex + 1,
                            "@cron is not allowed in .rabbitmq files"
                        )
                    )
                }
            }

            override fun enterAtCronRetry(ctx: CztctlParser.AtCronRetryContext) {
                // @cronRetry keyword color: #A5D5F4
                val kwToken = ctx.ATCRONRETRY()?.symbol
                if (kwToken != null) {
                    highlights.add(SemanticHighlight(kwToken.startIndex, kwToken.stopIndex + 1, CztctlSyntaxHighlighter.SERVICE_ANNOTATION_ATTRS))
                }
                // semantic error: forbid @cronRetry in .rabbitmq files
                if (input.fileExtension == "rabbitmq") {
                    val token = ctx.start
                    errors.add(
                        SyntaxError(
                            token.line,
                            token.charPositionInLine,
                            ctx.stop.stopIndex - ctx.start.startIndex + 1,
                            "@cronRetry is not allowed in .rabbitmq files"
                        )
                    )
                }
            }

            // ── Semantic highlighting ──

            // ====== info module ======

            // info keyword: #519D9E, bold + italic
            override fun enterInfoSpec(ctx: CztctlParser.InfoSpecContext) {
                val t = ctx.INFO()?.symbol ?: return
                highlights.add(SemanticHighlight(t.startIndex, t.stopIndex + 1, CztctlSyntaxHighlighter.TEAL_KEYWORD_ATTRS))
            }

            // ====== type module ======

            // type keyword: #519D9E, bold + italic
            override fun enterTypeLit(ctx: CztctlParser.TypeLitContext) {
                val t = ctx.TYPE()?.symbol ?: return
                highlights.add(SemanticHighlight(t.startIndex, t.stopIndex + 1, CztctlSyntaxHighlighter.TEAL_KEYWORD_ATTRS))
            }

            override fun enterTypeBlock(ctx: CztctlParser.TypeBlockContext) {
                val t = ctx.TYPE()?.symbol ?: return
                highlights.add(SemanticHighlight(t.startIndex, t.stopIndex + 1, CztctlSyntaxHighlighter.TEAL_KEYWORD_ATTRS))
            }

            // struct name: #6AAFE6, bold
            override fun enterTypeStruct(ctx: CztctlParser.TypeStructContext) {
                val n = ctx.identifier() ?: return
                highlights.add(SemanticHighlight(n.start.startIndex, n.stop.stopIndex + 1, CztctlSyntaxHighlighter.STRUCT_NAME_ATTRS))
            }

            override fun enterTypeBlockStruct(ctx: CztctlParser.TypeBlockStructContext) {
                val n = ctx.identifier() ?: return
                highlights.add(SemanticHighlight(n.start.startIndex, n.stop.stopIndex + 1, CztctlSyntaxHighlighter.STRUCT_NAME_ATTRS))
            }

            // field name: #A5D5F4, data type: #a3daff, tag: #D4DFE6
            override fun enterNormalField(ctx: CztctlParser.NormalFieldContext) {
                val nameCtx = ctx.identifier() ?: return
                highlights.add(SemanticHighlight(nameCtx.start.startIndex, nameCtx.stop.stopIndex + 1, CztctlSyntaxHighlighter.FIELD_NAME_ATTRS))

                val typeCtx = ctx.dataType()
                if (typeCtx != null) {
                    highlights.add(SemanticHighlight(typeCtx.start.startIndex, typeCtx.stop.stopIndex + 1, CztctlSyntaxHighlighter.FIELD_TYPE_ATTRS))
                }

                val tagToken = ctx.RAW_STRING()?.symbol
                if (tagToken != null) {
                    highlights.add(SemanticHighlight(tagToken.startIndex, tagToken.stopIndex + 1, CztctlSyntaxHighlighter.FIELD_TAG_ATTRS))
                }
            }

            // ====== @server module ======

            // @server keyword: #519D9E, bold + italic
            override fun enterAtServer(ctx: CztctlParser.AtServerContext) {
                val t = ctx.ATSERVER()?.symbol ?: return
                highlights.add(SemanticHighlight(t.startIndex, t.stopIndex + 1, CztctlSyntaxHighlighter.TEAL_KEYWORD_ATTRS))
            }

            // KV keys in info/@server: #96D4D0; @server non-string values: #CEEBE7 italic
            override fun enterKvLit(ctx: CztctlParser.KvLitContext) {
                val isInfo = ctx.parent is CztctlParser.InfoSpecContext
                val isServer = ctx.parent is CztctlParser.AtServerContext
                if (!isInfo && !isServer) return

                val keyCtx = ctx.identifier() ?: return
                highlights.add(SemanticHighlight(keyCtx.start.startIndex, keyCtx.stop.stopIndex + 1, CztctlSyntaxHighlighter.KV_KEY_ATTRS))

                if (isServer) {
                    val valueCtx = ctx.kvValue() ?: return
                    val isString = valueCtx.STRING() != null || valueCtx.RAW_STRING() != null
                    if (!isString) {
                        highlights.add(SemanticHighlight(valueCtx.start.startIndex, valueCtx.stop.stopIndex + 1, CztctlSyntaxHighlighter.KV_VALUE_PLAIN_ATTRS))
                    }
                }
            }

            // ====== service module ======

            // service keyword: #519D9E, bold + italic
            override fun enterServiceApi(ctx: CztctlParser.ServiceApiContext) {
                val t = ctx.SERVICE()?.symbol ?: return
                highlights.add(SemanticHighlight(t.startIndex, t.stopIndex + 1, CztctlSyntaxHighlighter.TEAL_KEYWORD_ATTRS))
            }

            // service name: #6AAFE6
            override fun enterServiceName(ctx: CztctlParser.ServiceNameContext) {
                highlights.add(SemanticHighlight(ctx.start.startIndex, ctx.stop.stopIndex + 1, CztctlSyntaxHighlighter.SERVICE_NAME_ATTRS))
            }

            // @handler keyword: #A5D5F4, handler name: #6AAFE6
            override fun enterAtHandler(ctx: CztctlParser.AtHandlerContext) {
                val kwToken = ctx.ATHANDLER()?.symbol
                if (kwToken != null) {
                    highlights.add(SemanticHighlight(kwToken.startIndex, kwToken.stopIndex + 1, CztctlSyntaxHighlighter.SERVICE_ANNOTATION_ATTRS))
                }
                val idCtx = ctx.identifier() ?: return
                highlights.add(SemanticHighlight(idCtx.start.startIndex, idCtx.stop.stopIndex + 1, CztctlSyntaxHighlighter.HANDLER_NAME_ATTRS))
            }

            // @doc keyword: #A5D5F4
            override fun enterAtDoc(ctx: CztctlParser.AtDocContext) {
                val t = ctx.ATDOC()?.symbol ?: return
                highlights.add(SemanticHighlight(t.startIndex, t.stopIndex + 1, CztctlSyntaxHighlighter.SERVICE_ANNOTATION_ATTRS))
            }

            // Route name: .cron task type #8FBC94 italic, .rabbitmq queue #C5E99B italic
            override fun enterRoute(ctx: CztctlParser.RouteContext) {
                val routeNameCtx = ctx.routeName() ?: return
                val attrs = if (input.fileExtension == "cron")
                    CztctlSyntaxHighlighter.CRON_ROUTE_ATTRS
                else
                    CztctlSyntaxHighlighter.RABBITMQ_ROUTE_ATTRS
                highlights.add(SemanticHighlight(routeNameCtx.start.startIndex, routeNameCtx.stop.stopIndex + 1, attrs))
            }

            // Route body parameter: #6AAFE6, bold + italic
            override fun enterBody(ctx: CztctlParser.BodyContext) {
                val idCtx = ctx.identifier() ?: return
                highlights.add(SemanticHighlight(idCtx.start.startIndex, idCtx.stop.stopIndex + 1, CztctlSyntaxHighlighter.ROUTE_PARAM_ATTRS))
            }

        }, tree)

        return AnnotationResult(errors, highlights)
    }

    override fun apply(file: PsiFile, result: AnnotationResult, holder: AnnotationHolder) {
        val text = file.text
        val lineOffsets = buildLineOffsets(text)

        // Apply syntax/semantic errors
        for (error in result.errors) {
            val lineIndex = error.line - 1
            if (lineIndex < 0 || lineIndex >= lineOffsets.size) continue

            val lineStart = lineOffsets[lineIndex]
            val start = lineStart + error.charPositionInLine
            val end = minOf(start + error.length, text.length)

            if (start >= text.length) continue

            val range = TextRange(start, maxOf(start + 1, end))
            holder.newAnnotation(error.severity, error.message)
                .range(range)
                .create()
        }

        // Apply semantic highlights (enforced to override lexer coloring)
        for (highlight in result.highlights) {
            if (highlight.startOffset >= text.length) continue
            val end = minOf(highlight.endOffset, text.length)
            if (highlight.startOffset >= end) continue

            val range = TextRange(highlight.startOffset, end)
            holder.newSilentAnnotation(HighlightSeverity.INFORMATION)
                .range(range)
                .enforcedTextAttributes(highlight.attrs)
                .create()
        }
    }

    private fun buildLineOffsets(text: String): List<Int> {
        val offsets = mutableListOf(0)
        text.forEachIndexed { index, ch ->
            if (ch == '\n') offsets.add(index + 1)
        }
        return offsets
    }
}
