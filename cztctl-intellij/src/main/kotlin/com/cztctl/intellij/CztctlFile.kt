package com.cztctl.intellij

import com.intellij.extapi.psi.PsiFileBase
import com.intellij.openapi.fileTypes.FileType
import com.intellij.psi.FileViewProvider

class CztctlFile(viewProvider: FileViewProvider) : PsiFileBase(viewProvider, CztctlLanguage) {
    override fun getFileType(): FileType = CztctlFileType
    override fun toString(): String = "cztctl DSL File"
}
