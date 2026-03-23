package com.cztctl.intellij

import com.intellij.lang.Language

object CztctlLanguage : Language("cztctl") {
    override fun getDisplayName(): String = "cztctl DSL"
    override fun isCaseSensitive(): Boolean = true
}
