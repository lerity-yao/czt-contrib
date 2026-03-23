package com.cztctl.intellij.parser

import com.cztctl.intellij.CztctlFile
import com.cztctl.intellij.CztctlFileType
import com.intellij.lang.ASTNode
import com.intellij.lang.ParserDefinition
import com.intellij.lang.PsiParser
import com.intellij.lexer.Lexer
import com.intellij.openapi.project.Project
import com.intellij.psi.FileViewProvider
import com.intellij.psi.PsiElement
import com.intellij.psi.PsiFile
import com.intellij.psi.tree.IFileElementType
import com.intellij.psi.tree.TokenSet
import com.intellij.lang.PsiBuilder
import com.intellij.psi.impl.source.tree.LeafPsiElement

class CztctlParserDefinition : ParserDefinition {

    override fun createLexer(project: Project?): Lexer = CztctlLexerAdapter()

    override fun createParser(project: Project?): PsiParser = CztctlPsiParser()

    override fun getFileNodeType(): IFileElementType = CztctlElementTypes.FILE

    override fun getCommentTokens(): TokenSet = CztctlElementTypes.COMMENTS

    override fun getStringLiteralElements(): TokenSet = CztctlElementTypes.STRINGS

    override fun getWhitespaceTokens(): TokenSet = CztctlElementTypes.WHITE_SPACES

    override fun createElement(node: ASTNode): PsiElement = LeafPsiElement(node.elementType, node.text)

    override fun createFile(viewProvider: FileViewProvider): PsiFile = CztctlFile(viewProvider)
}

/**
 * Lightweight PsiParser — creates a flat file-level PSI tree.
 * Real syntax checking is done by CztctlExternalAnnotator using ANTLR4 full parse.
 */
private class CztctlPsiParser : PsiParser {
    override fun parse(root: com.intellij.psi.tree.IElementType, builder: PsiBuilder): ASTNode {
        val marker = builder.mark()
        while (!builder.eof()) {
            builder.advanceLexer()
        }
        marker.done(root)
        return builder.treeBuilt
    }
}
