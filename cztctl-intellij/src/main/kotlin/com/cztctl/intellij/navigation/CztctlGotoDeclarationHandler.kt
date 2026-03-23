package com.cztctl.intellij.navigation

import com.cztctl.intellij.CztctlLanguage
import com.cztctl.intellij.parser.CztctlBaseListener
import com.cztctl.intellij.parser.CztctlLexer
import com.cztctl.intellij.parser.CztctlParser
import com.intellij.codeInsight.navigation.actions.GotoDeclarationHandler
import com.intellij.openapi.editor.Editor
import com.intellij.openapi.vfs.LocalFileSystem
import com.intellij.psi.PsiElement
import com.intellij.psi.PsiManager
import org.antlr.v4.runtime.CharStreams
import org.antlr.v4.runtime.CommonTokenStream
import org.antlr.v4.runtime.tree.ParseTreeWalker
import java.io.File

/**
 * Ctrl+Click navigation for cztctl DSL files:
 * 1. Route body parameter (type reference) → jump to type definition (current file or imported files)
 * 2. Import path string → jump to the imported file
 */
class CztctlGotoDeclarationHandler : GotoDeclarationHandler {

    override fun getGotoDeclarationTargets(
        sourceElement: PsiElement?,
        offset: Int,
        editor: Editor?
    ): Array<PsiElement>? {
        val element = sourceElement ?: return null
        val file = element.containingFile ?: return null
        if (file.language != CztctlLanguage) return null

        val text = file.text
        val project = file.project

        // Parse with ANTLR4
        val lexer = CztctlLexer(CharStreams.fromString(text))
        lexer.removeErrorListeners()
        val tokens = CommonTokenStream(lexer)
        val parser = CztctlParser(tokens)
        parser.removeErrorListeners()
        val tree = parser.api()

        // Collect navigation targets from parse tree
        val importPaths = mutableListOf<ImportPathInfo>()  // import path → file
        val typeRefs = mutableListOf<TypeRefInfo>()         // body identifier → type name
        val typeDefs = mutableListOf<TypeDefInfo>()         // type definitions

        val walker = ParseTreeWalker()
        walker.walk(object : CztctlBaseListener() {

            // Collect import paths
            override fun enterImportValue(ctx: CztctlParser.ImportValueContext) {
                val token = ctx.STRING() ?: return
                // STRING includes quotes, so inner range is [start+1, stop]
                val pathStart = token.symbol.startIndex + 1
                val pathEnd = token.symbol.stopIndex  // stop is last char (the closing quote)
                val rawPath = token.text.removeSurrounding("\"")
                importPaths.add(ImportPathInfo(pathStart, pathEnd, rawPath))
            }

            // Collect route body parameters (type references)
            override fun enterBody(ctx: CztctlParser.BodyContext) {
                val idCtx = ctx.identifier() ?: return
                val name = idCtx.text
                val start = idCtx.start.startIndex
                val end = idCtx.stop.stopIndex + 1
                typeRefs.add(TypeRefInfo(start, end, name))
            }

            // Collect type definitions: typeStruct
            override fun enterTypeStruct(ctx: CztctlParser.TypeStructContext) {
                val nameCtx = ctx.identifier() ?: return
                typeDefs.add(TypeDefInfo(nameCtx.text, nameCtx.start.startIndex))
            }

            // Collect type definitions: typeBlockStruct
            override fun enterTypeBlockStruct(ctx: CztctlParser.TypeBlockStructContext) {
                val nameCtx = ctx.identifier() ?: return
                typeDefs.add(TypeDefInfo(nameCtx.text, nameCtx.start.startIndex))
            }

            // Collect type definitions: typeAlias
            override fun enterTypeAlias(ctx: CztctlParser.TypeAliasContext) {
                val nameCtx = ctx.identifier() ?: return
                typeDefs.add(TypeDefInfo(nameCtx.text, nameCtx.start.startIndex))
            }

            // Collect type definitions: typeBlockAlias
            override fun enterTypeBlockAlias(ctx: CztctlParser.TypeBlockAliasContext) {
                val nameCtx = ctx.identifier() ?: return
                typeDefs.add(TypeDefInfo(nameCtx.text, nameCtx.start.startIndex))
            }

        }, tree)

        // Check if cursor is on an import path string
        for (imp in importPaths) {
            if (offset in imp.startOffset until imp.endOffset) {
                // Resolve import path relative to current file
                val currentDir = file.virtualFile?.parent?.path ?: return null
                val targetPath = File(currentDir, imp.path).canonicalPath
                val vf = LocalFileSystem.getInstance().findFileByPath(targetPath) ?: return null
                val psiFile = PsiManager.getInstance(project).findFile(vf) ?: return null
                return arrayOf(psiFile)
            }
        }

        // Check if cursor is on a route body parameter (type reference)
        for (ref in typeRefs) {
            if (offset in ref.startOffset until ref.endOffset) {
                // Search current file for type definition
                val localDef = typeDefs.find { it.name == ref.typeName }
                if (localDef != null) {
                    val targetElement = file.findElementAt(localDef.offset)
                    if (targetElement != null) return arrayOf(targetElement)
                }

                // Search imported files for type definition
                val currentDir = file.virtualFile?.parent?.path ?: return null
                for (imp in importPaths) {
                    val targetPath = File(currentDir, imp.path).canonicalPath
                    val vf = LocalFileSystem.getInstance().findFileByPath(targetPath) ?: continue
                    val importedPsiFile = PsiManager.getInstance(project).findFile(vf) ?: continue
                    val importedText = importedPsiFile.text

                    // Parse imported file to find type definitions
                    val impLexer = CztctlLexer(CharStreams.fromString(importedText))
                    impLexer.removeErrorListeners()
                    val impTokens = CommonTokenStream(impLexer)
                    val impParser = CztctlParser(impTokens)
                    impParser.removeErrorListeners()
                    val impTree = impParser.api()

                    val impTypeDefs = mutableListOf<TypeDefInfo>()
                    val impWalker = ParseTreeWalker()
                    impWalker.walk(object : CztctlBaseListener() {
                        override fun enterTypeStruct(ctx: CztctlParser.TypeStructContext) {
                            val nameCtx = ctx.identifier() ?: return
                            impTypeDefs.add(TypeDefInfo(nameCtx.text, nameCtx.start.startIndex))
                        }
                        override fun enterTypeBlockStruct(ctx: CztctlParser.TypeBlockStructContext) {
                            val nameCtx = ctx.identifier() ?: return
                            impTypeDefs.add(TypeDefInfo(nameCtx.text, nameCtx.start.startIndex))
                        }
                        override fun enterTypeAlias(ctx: CztctlParser.TypeAliasContext) {
                            val nameCtx = ctx.identifier() ?: return
                            impTypeDefs.add(TypeDefInfo(nameCtx.text, nameCtx.start.startIndex))
                        }
                        override fun enterTypeBlockAlias(ctx: CztctlParser.TypeBlockAliasContext) {
                            val nameCtx = ctx.identifier() ?: return
                            impTypeDefs.add(TypeDefInfo(nameCtx.text, nameCtx.start.startIndex))
                        }
                    }, impTree)

                    val match = impTypeDefs.find { it.name == ref.typeName }
                    if (match != null) {
                        val targetElement = importedPsiFile.findElementAt(match.offset)
                        if (targetElement != null) return arrayOf(targetElement)
                    }
                }

                return null
            }
        }

        return null
    }

    private data class ImportPathInfo(val startOffset: Int, val endOffset: Int, val path: String)
    private data class TypeRefInfo(val startOffset: Int, val endOffset: Int, val typeName: String)
    private data class TypeDefInfo(val name: String, val offset: Int)
}
