package com.cztctl.intellij

import com.intellij.icons.AllIcons
import com.intellij.openapi.fileTypes.LanguageFileType
import javax.swing.Icon

object CztctlFileType : LanguageFileType(CztctlLanguage) {
    override fun getName(): String = "cztctl DSL"
    override fun getDescription(): String = "cztctl DSL file (.cron / .rabbitmq)"
    override fun getDefaultExtension(): String = "cron"
    override fun getIcon(): Icon = AllIcons.FileTypes.Custom
}
