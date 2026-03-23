plugins {
    id("org.jetbrains.intellij.platform") version "2.5.0"
    kotlin("jvm") version "2.2.0"
    antlr
}

group = "com.cztctl"
version = "0.0.1"

repositories {
    mavenCentral()
    intellijPlatform {
        defaultRepositories()
    }
}

dependencies {
    antlr("org.antlr:antlr4:4.13.2")
    implementation("org.antlr:antlr4-runtime:4.13.2")

    intellijPlatform {
        local("/home/yaox/app/goland")
    }
}

kotlin {
    jvmToolchain(21)
}

tasks {
    generateGrammarSource {
        val outDir = "${project.buildDir}/generated-src/antlr/main"
        arguments = arguments + listOf("-visitor", "-package", "com.cztctl.intellij.parser")
        outputDirectory = file(outDir)
    }

    compileKotlin {
        dependsOn(generateGrammarSource)
    }

    compileJava {
        dependsOn(generateGrammarSource)
    }

    patchPluginXml {
        sinceBuild.set("243")
        untilBuild.set("253.*")
    }

    prepareSandbox {
        from("src/main/textmate") {
            into(project.name)
        }
    }
}

sourceSets {
    main {
        java {
            srcDir("${project.buildDir}/generated-src/antlr/main")
        }
    }
}

// Exclude ANTLR4 tool from plugin distribution (only runtime needed)
configurations {
    runtimeClasspath {
        exclude(group = "org.antlr", module = "antlr4")
    }
}
