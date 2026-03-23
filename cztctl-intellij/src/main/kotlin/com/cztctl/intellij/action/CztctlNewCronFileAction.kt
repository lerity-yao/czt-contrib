package com.cztctl.intellij.action

import com.cztctl.intellij.CztctlFileType
import com.intellij.ide.actions.CreateFileFromTemplateAction
import com.intellij.ide.actions.CreateFileFromTemplateDialog
import com.intellij.openapi.project.Project
import com.intellij.psi.PsiDirectory

class CztctlNewCronFileAction : CreateFileFromTemplateAction(
    TITLE, DESCRIPTION, CztctlFileType.icon
) {
    companion object {
        private const val TITLE = "Cron File (.cron)"
        private const val DESCRIPTION = "Create a new cztctl .cron file"
        private const val TEMPLATE_NAME = "Cztctl Cron File"
    }

    override fun buildDialog(
        project: Project,
        directory: PsiDirectory,
        builder: CreateFileFromTemplateDialog.Builder
    ) {
        builder.setTitle("New Cron File")
            .addKind(TITLE, CztctlFileType.icon, TEMPLATE_NAME)
    }

    override fun getActionName(
        directory: PsiDirectory,
        newName: String,
        templateName: String
    ): String = "Create Cron File: $newName"
}
