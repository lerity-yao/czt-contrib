package com.cztctl.intellij.parser

import com.cztctl.intellij.CztctlLanguage
import com.intellij.psi.tree.IElementType
import com.intellij.psi.tree.IFileElementType
import com.intellij.psi.tree.TokenSet

object CztctlElementTypes {

    // File root
    val FILE = IFileElementType(CztctlLanguage)

    // Token types — one per ANTLR4 lexer token, lazily built from CztctlLexer vocabulary
    private val tokenTypes: Array<IElementType?> by lazy {
        val vocab = CztctlLexer.VOCABULARY
        Array(vocab.maxTokenType + 1) { i ->
            val name = vocab.getSymbolicName(i)
            if (name != null) CztctlTokenType(name) else null
        }
    }

    fun tokenType(antlrType: Int): IElementType =
        tokenTypes.getOrNull(antlrType) ?: BAD_CHARACTER

    val BAD_CHARACTER = CztctlTokenType("BAD_CHARACTER")

    // Rule element type (for PSI tree internal nodes)
    class CztctlRuleType(name: String) : IElementType(name, CztctlLanguage)

    // Token element type
    class CztctlTokenType(name: String) : IElementType(name, CztctlLanguage) {
        override fun toString(): String = "CztctlTokenType.${super.toString()}"
    }

    // Token sets for brace matching, comments, strings
    val COMMENTS: TokenSet by lazy {
        TokenSet.create(
            tokenType(CztctlLexer.COMMENT),
            tokenType(CztctlLexer.LINE_COMMENT)
        )
    }

    val STRINGS: TokenSet by lazy {
        TokenSet.create(
            tokenType(CztctlLexer.STRING),
            tokenType(CztctlLexer.RAW_STRING)
        )
    }

    val WHITE_SPACES: TokenSet by lazy {
        TokenSet.create(tokenType(CztctlLexer.WS))
    }
}
