package com.cztctl.intellij.action

import com.cztctl.intellij.CztctlFileType
import com.intellij.ide.actions.CreateFileFromTemplateAction
import com.intellij.ide.actions.CreateFileFromTemplateDialog
import com.intellij.openapi.project.Project
import com.intellij.psi.PsiDirectory

class CztctlNewRabbitmqFileAction : CreateFileFromTemplateAction(
    TITLE, DESCRIPTION, CztctlFileType.icon
) {
    companion object {
        private const val TITLE = "RabbitMQ File (.rabbitmq)"
        private const val DESCRIPTION = "Create a new cztctl .rabbitmq file"
        private const val TEMPLATE_NAME = "Cztctl RabbitMQ File"
    }

    override fun buildDialog(
        project: Project,
        directory: PsiDirectory,
        builder: CreateFileFromTemplateDialog.Builder
    ) {
        builder.setTitle("New RabbitMQ File")
            .addKind(TITLE, CztctlFileType.icon, TEMPLATE_NAME)
    }

    override fun getActionName(
        directory: PsiDirectory,
        newName: String,
        templateName: String
    ): String = "Create RabbitMQ File: $newName"
}
