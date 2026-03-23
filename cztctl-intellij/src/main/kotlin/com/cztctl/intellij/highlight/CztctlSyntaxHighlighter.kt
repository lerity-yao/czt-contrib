package com.cztctl.intellij.highlight

import com.cztctl.intellij.parser.CztctlElementTypes
import com.cztctl.intellij.parser.CztctlLexer
import com.cztctl.intellij.parser.CztctlLexerAdapter
import com.intellij.lexer.Lexer
import com.intellij.openapi.editor.DefaultLanguageHighlighterColors as C
import com.intellij.openapi.editor.HighlighterColors
import com.intellij.openapi.editor.colors.TextAttributesKey
import com.intellij.openapi.editor.colors.TextAttributesKey.createTextAttributesKey
import com.intellij.openapi.editor.markup.TextAttributes
import com.intellij.openapi.fileTypes.SyntaxHighlighterBase
import com.intellij.psi.tree.IElementType
import com.intellij.ui.JBColor
import java.awt.Color
import java.awt.Font

class CztctlSyntaxHighlighter : SyntaxHighlighterBase() {

    companion object {
        // Token-level attribute keys
        val KEYWORD = createTextAttributesKey("CZTCTL_KEYWORD", C.KEYWORD)
        val ANNOTATION = createTextAttributesKey("CZTCTL_ANNOTATION", C.METADATA)
        val HANDLER_KEYWORD = createTextAttributesKey("CZTCTL_HANDLER_KEYWORD", C.KEYWORD)
        val STRING = createTextAttributesKey("CZTCTL_STRING", C.STRING)
        val NUMBER = createTextAttributesKey("CZTCTL_NUMBER", C.NUMBER)
        val COMMENT = createTextAttributesKey("CZTCTL_COMMENT", C.LINE_COMMENT)
        val BLOCK_COMMENT = createTextAttributesKey("CZTCTL_BLOCK_COMMENT", C.BLOCK_COMMENT)
        val IDENTIFIER = createTextAttributesKey("CZTCTL_IDENTIFIER", C.IDENTIFIER)
        val BRACES = createTextAttributesKey("CZTCTL_BRACES", C.BRACES)
        val BRACKETS = createTextAttributesKey("CZTCTL_BRACKETS", C.BRACKETS)
        val PARENS = createTextAttributesKey("CZTCTL_PARENS", C.PARENTHESES)
        val OPERATOR = createTextAttributesKey("CZTCTL_OPERATOR", C.OPERATION_SIGN)
        val DOT_KEY = createTextAttributesKey("CZTCTL_DOT", C.DOT)
        val COMMA_KEY = createTextAttributesKey("CZTCTL_COMMA", C.COMMA)
        val BAD_CHAR = createTextAttributesKey("CZTCTL_BAD_CHARACTER", HighlighterColors.BAD_CHARACTER)

        // Enforced TextAttributes for semantic highlighting (ExternalAnnotator)

        // ── info / @server module keywords: #519D9E, bold + italic ──
        val TEAL_KEYWORD_ATTRS = TextAttributes(
            JBColor(Color(0x519D9E), Color(0x519D9E)), null, null, null, Font.BOLD or Font.ITALIC
        )
        // ── type / service module keywords: #21486F, bold + italic ──
        val BLUE_KEYWORD_ATTRS = TextAttributes(
            JBColor(Color(0x21486F), Color(0x21486F)), null, null, null, Font.BOLD or Font.ITALIC
        )
        // ── KV keys in info / @server: #96D4D0 ──
        val KV_KEY_ATTRS = TextAttributes(
            JBColor(Color(0x96D4D0), Color(0x96D4D0)), null, null, null, Font.PLAIN
        )
        // ── @server KV non-string values: #CEEBE7, italic ──
        val KV_VALUE_PLAIN_ATTRS = TextAttributes(
            JBColor(Color(0xCEEBE7), Color(0xCEEBE7)), null, null, null, Font.ITALIC
        )
        // ── struct name: #6AAFE6, bold ──
        val STRUCT_NAME_ATTRS = TextAttributes(
            JBColor(Color(0x6AAFE6), Color(0x6AAFE6)), null, null, null, Font.BOLD
        )
        // ── field name: #A5D5F4 ──
        val FIELD_NAME_ATTRS = TextAttributes(
            JBColor(Color(0xA5D5F4), Color(0xA5D5F4)), null, null, null, Font.PLAIN
        )
        // ── field data type: #a3daff ──
        val FIELD_TYPE_ATTRS = TextAttributes(
            JBColor(Color(0xa3daff), Color(0xa3daff)), null, null, null, Font.PLAIN
        )
        // ── field tag (RAW_STRING): #D4DFE6 ──
        val FIELD_TAG_ATTRS = TextAttributes(
            JBColor(Color(0xD4DFE6), Color(0xD4DFE6)), null, null, null, Font.PLAIN
        )
        // ── service name: #6AAFE6 ──
        val SERVICE_NAME_ATTRS = TextAttributes(
            JBColor(Color(0x6AAFE6), Color(0x6AAFE6)), null, null, null, Font.PLAIN
        )
        // ── handler name: #6AAFE6 ──
        val HANDLER_NAME_ATTRS = TextAttributes(
            JBColor(Color(0x6AAFE6), Color(0x6AAFE6)), null, null, null, Font.PLAIN
        )
        // ── cron route (task type): #8FBC94, italic ──
        val CRON_ROUTE_ATTRS = TextAttributes(
            JBColor(Color(0x8FBC94), Color(0x8FBC94)), null, null, null, Font.ITALIC
        )
        // ── rabbitmq route (queue): #C5E99B, italic ──
        val RABBITMQ_ROUTE_ATTRS = TextAttributes(
            JBColor(Color(0xC5E99B), Color(0xC5E99B)), null, null, null, Font.ITALIC
        )
        // ── route body parameter: #6AAFE6, bold + italic ──
        val ROUTE_PARAM_ATTRS = TextAttributes(
            JBColor(Color(0x6AAFE6), Color(0x6AAFE6)), null, null, null, Font.BOLD or Font.ITALIC
        )
        // ── service annotation keywords (@handler/@doc/@cron/@cronRetry): #A5D5F4 ──
        val SERVICE_ANNOTATION_ATTRS = TextAttributes(
            JBColor(Color(0xA5D5F4), Color(0xA5D5F4)), null, null, null, Font.PLAIN
        )

        private val EMPTY = arrayOf<TextAttributesKey>()
    }

    override fun getHighlightingLexer(): Lexer = CztctlLexerAdapter()

    override fun getTokenHighlights(tokenType: IElementType): Array<TextAttributesKey> {
        return when (tokenType) {
            // Language keywords
            CztctlElementTypes.tokenType(CztctlLexer.SYNTAX),
            CztctlElementTypes.tokenType(CztctlLexer.IMPORT),
            CztctlElementTypes.tokenType(CztctlLexer.INFO),
            CztctlElementTypes.tokenType(CztctlLexer.TYPE),
            CztctlElementTypes.tokenType(CztctlLexer.SERVICE),
            CztctlElementTypes.tokenType(CztctlLexer.MAP),
            CztctlElementTypes.tokenType(CztctlLexer.STRUCT) -> arrayOf(KEYWORD)

            // Annotation keywords
            CztctlElementTypes.tokenType(CztctlLexer.ATDOC),
            CztctlElementTypes.tokenType(CztctlLexer.ATSERVER),
            CztctlElementTypes.tokenType(CztctlLexer.ATCRON),
            CztctlElementTypes.tokenType(CztctlLexer.ATCRONRETRY) -> arrayOf(ANNOTATION)

            // @handler keyword (separate color from other annotations)
            CztctlElementTypes.tokenType(CztctlLexer.ATHANDLER) -> arrayOf(HANDLER_KEYWORD)

            // Literals
            CztctlElementTypes.tokenType(CztctlLexer.STRING),
            CztctlElementTypes.tokenType(CztctlLexer.RAW_STRING) -> arrayOf(STRING)

            CztctlElementTypes.tokenType(CztctlLexer.INT) -> arrayOf(NUMBER)

            // Comments
            CztctlElementTypes.tokenType(CztctlLexer.LINE_COMMENT) -> arrayOf(COMMENT)
            CztctlElementTypes.tokenType(CztctlLexer.COMMENT) -> arrayOf(BLOCK_COMMENT)

            // Punctuation
            CztctlElementTypes.tokenType(CztctlLexer.LBRACE),
            CztctlElementTypes.tokenType(CztctlLexer.RBRACE) -> arrayOf(BRACES)

            CztctlElementTypes.tokenType(CztctlLexer.LBRACK),
            CztctlElementTypes.tokenType(CztctlLexer.RBRACK) -> arrayOf(BRACKETS)

            CztctlElementTypes.tokenType(CztctlLexer.LPAREN),
            CztctlElementTypes.tokenType(CztctlLexer.RPAREN) -> arrayOf(PARENS)

            CztctlElementTypes.tokenType(CztctlLexer.ASSIGN),
            CztctlElementTypes.tokenType(CztctlLexer.COLON),
            CztctlElementTypes.tokenType(CztctlLexer.STAR) -> arrayOf(OPERATOR)

            CztctlElementTypes.tokenType(CztctlLexer.DOT) -> arrayOf(DOT_KEY)
            CztctlElementTypes.tokenType(CztctlLexer.COMMA) -> arrayOf(COMMA_KEY)

            // Identifier
            CztctlElementTypes.tokenType(CztctlLexer.ID) -> arrayOf(IDENTIFIER)

            // Bad character
            CztctlElementTypes.BAD_CHARACTER -> arrayOf(BAD_CHAR)

            else -> EMPTY
        }
    }
}
