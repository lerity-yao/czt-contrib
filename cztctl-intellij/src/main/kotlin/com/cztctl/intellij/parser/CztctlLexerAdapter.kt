package com.cztctl.intellij.parser

import com.intellij.lexer.LexerBase
import com.intellij.psi.tree.IElementType
import org.antlr.v4.runtime.CharStreams
import org.antlr.v4.runtime.Token

/**
 * Wraps ANTLR4 CztctlLexer as an IntelliJ Lexer.
 * Produces IElementType tokens mapped from ANTLR4 token types.
 */
class CztctlLexerAdapter : LexerBase() {

    private var buffer: CharSequence = ""
    private var startOffset: Int = 0
    private var endOffset: Int = 0
    private var lexer: CztctlLexer? = null
    private var currentToken: Token? = null

    override fun start(buffer: CharSequence, startOffset: Int, endOffset: Int, initialState: Int) {
        this.buffer = buffer
        this.startOffset = startOffset
        this.endOffset = endOffset
        val text = buffer.subSequence(startOffset, endOffset).toString()
        lexer = CztctlLexer(CharStreams.fromString(text))
        lexer!!.removeErrorListeners()
        advance()
    }

    override fun getState(): Int = 0

    override fun getTokenType(): IElementType? {
        val token = currentToken ?: return null
        if (token.type == Token.EOF) return null
        return CztctlElementTypes.tokenType(token.type)
    }

    override fun getTokenStart(): Int {
        val token = currentToken ?: return endOffset
        return startOffset + token.startIndex
    }

    override fun getTokenEnd(): Int {
        val token = currentToken ?: return endOffset
        return startOffset + token.stopIndex + 1
    }

    override fun advance() {
        currentToken = lexer?.nextToken()
    }

    override fun getBufferSequence(): CharSequence = buffer

    override fun getBufferEnd(): Int = endOffset
}
