# cztctl-intellij 维护指南

## 1. 语法变更同步

当 `czt-contrib/cztctl` CLI 更新了 DSL 文件语法（如新增关键字、注解、语法结构），需要同步更新插件。

### 1.1 变更影响范围

| CLI 变更类型 | 需要修改的插件文件 |
|---|---|
| 新增关键字（如 `queue`） | `Cztctl.g4`（lexer token + parser rule + `identifier` 规则） |
| 新增注解（如 `@retry`） | `Cztctl.g4`（lexer token + parser rule） |
| 修改语法结构 | `Cztctl.g4`（parser rule） |
| 新增文件类型语义限制 | `CztctlExternalAnnotator.kt`（语义错误监听器） |
| 新增/修改语义着色 | `CztctlSyntaxHighlighter.kt`（色彩常量）+ `CztctlExternalAnnotator.kt`（着色监听器） |
| 新增 Token 高亮颜色 | `CztctlSyntaxHighlighter.kt`（`getTokenHighlights` 方法） |
| 新增/修改代码导航 | `CztctlGotoDeclarationHandler.kt` |
| 新增代码模板 | `src/main/resources/liveTemplates/cztctl.xml` |
| 修改新建文件模板 | `src/main/resources/fileTemplates/internal/*.ft` |

### 1.2 具体操作步骤

#### 新增关键字（以新增 `queue` 为例）

**Step 1**：修改 `src/main/antlr/com/cztctl/intellij/parser/Cztctl.g4`

```g4
// Lexer 部分：新增 token
QUEUE:              'queue';

// Parser 部分：在需要使用的规则中引用 QUEUE
// 如果关键字也能做标识符，还需加入 identifier 规则：
identifier:     ID | SYNTAX | IMPORT | INFO | TYPE | SERVICE | MAP | STRUCT | QUEUE;
```

**Step 2**：修改 `CztctlSyntaxHighlighter.kt`，在 `getTokenHighlights` 中添加映射：

```kotlin
CztctlLexer.QUEUE -> KEYWORD_KEYS
```

**Step 3**：构建验证

```bash
JAVA_HOME=/path/to/goland/jbr ./gradlew clean build
```

#### 新增注解（以新增 `@retry` 为例）

**Step 1**：修改 `Cztctl.g4`

```g4
// Lexer 部分
ATRETRY:            '@retry';

// Parser 部分：在 serviceRoute 规则中添加
serviceRoute:   atDoc? atCron? atCronRetry? atRetry? atHandler route;
atRetry:        ATRETRY INT;
```

**Step 2**：如果该注解有文件类型限制，修改 `CztctlExternalAnnotator.kt`：

```kotlin
override fun enterAtRetry(ctx: CztctlParser.AtRetryContext) {
    if (ext == "rabbitmq") {
        // 报红线
    }
}
```

**Step 3**：如需代码模板，修改 `src/main/resources/liveTemplates/cztctl.xml` 添加新模板。

**Step 4**：构建验证

```bash
JAVA_HOME=/path/to/goland/jbr ./gradlew clean build
```

#### 新增/修改语义着色

**Step 1**：在 `CztctlSyntaxHighlighter.kt` 的 `companion object` 中新增 `TextAttributes` 常量：

```kotlin
val MY_NEW_ATTRS = TextAttributes(
    JBColor(Color(0xABCDEF), Color(0xABCDEF)), null, null, null, Font.PLAIN
)
```

**Step 2**：在 `CztctlExternalAnnotator.kt` 的 ParseTreeWalker 中新增监听器：

```kotlin
override fun enterMyRule(ctx: CztctlParser.MyRuleContext) {
    val token = ctx.start
    highlights.add(SemanticHighlight(token.startIndex, token.stopIndex + 1, CztctlSyntaxHighlighter.MY_NEW_ATTRS))
}
```

注意：语义着色使用 `enforcedTextAttributes` 强制覆盖 Token 层颜色。

**Step 3**：构建验证。

#### 新增代码导航规则

修改 `CztctlGotoDeclarationHandler.kt`，在 `getGotoDeclarationTargets` 方法中新增判断逻辑。当前支持两种跳转：

- **路由参数** → 查找当前文件和 import 文件中的 type 定义
- **import 路径** → 打开对应文件

#### 修改新建文件模板

编辑 `src/main/resources/fileTemplates/internal/` 下的 `.ft` 文件即可修改新建文件的默认内容。模板使用 FreeMarker 语法，`${NAME}` 变量为用户输入的文件名。

### 1.3 注意事项

- 插件的 `Cztctl.g4` 是独立的 Combined Grammar（Java 目标），**不共享** CLI 的 .g4 文件
- CLI 使用 Go 手写解析器，插件使用 ANTLR4。两端行为需手动保持一致
- 修改 `.g4` 后 Gradle 构建会自动重新生成 `CztctlLexer.java` / `CztctlParser.java`，无需手动运行 ANTLR4
- `CztctlSyntaxHighlighter.kt` 中的 Token 常量（如 `CztctlLexer.QUEUE`）是 ANTLR4 生成的，`.g4` 改了后会自动更新

## 2. 版本发布

### 2.1 更新版本号

1. 修改 `build.gradle.kts`：

```kotlin
version = "0.0.2"  // 新版本号
```

2. 修改 `src/main/resources/META-INF/plugin.xml` 的 `<change-notes>`：

```xml
<change-notes><![CDATA[
    <h3>0.0.2</h3>
    <ul>
        <li>新功能描述</li>
    </ul>
    <h3>0.0.1</h3>
    <ul>
        <li>初始版本</li>
    </ul>
]]></change-notes>
```

### 2.2 构建插件包

```bash
cd czt-contrib/cztctl-intellij
JAVA_HOME=/path/to/goland/jbr ./gradlew clean buildPlugin
```

产物：`build/distributions/cztctl-intellij-<version>.zip`

### 2.3 本地验证

1. GoLand → Settings → Plugins → Install Plugin from Disk...
2. 选择新构建的 zip 文件
3. 重启 GoLand
4. 打开 `.cron` / `.rabbitmq` 文件，检查：
   - 语法高亮是否正常
   - 语义着色是否生效（info/type/@server/service 各模块独立配色）
   - 故意写错语法，红线是否正确标注
   - Ctrl+Click 路由参数是否能跳转到 type 定义
   - Ctrl+Click import 路径是否能打开对应文件
   - Live Templates 是否可用
   - `.rabbitmq` 中写 `@cron` 是否报红线
   - 右键 New → Cron File / RabbitMQ File 是否能正常创建文件

### 2.4 发布到 JetBrains Marketplace

#### 前置条件

1. 在 [JetBrains Marketplace](https://plugins.jetbrains.com/) 注册账号
2. 获取 Permanent Token：
   - 登录 → 右上角头像 → **My Tokens**
   - 点击 **Generate Token**
   - 保存生成的 Token

#### 方式一：网页手动上传

1. 登录 [JetBrains Marketplace](https://plugins.jetbrains.com/)
2. 点击右上角头像 → **Upload plugin**
3. 首次上传选择 **Upload new plugin**，后续更新选择已有插件的 **Upload update**
4. 上传 `build/distributions/cztctl-intellij-<version>.zip`
5. 填写插件信息：
   - **License**：与项目一致
   - **Tags**：`go-zero`, `DSL`, `cron`, `rabbitmq`, `code-generation`
   - **Category**：Custom Languages
6. 提交后等待 JetBrains 审核（通常 1-2 个工作日）

#### 方式二：Gradle 命令行发布

1. 设置环境变量：

```bash
export PUBLISH_TOKEN="你的 Marketplace Token"
```

2. 在 `build.gradle.kts` 中添加发布配置（如果还没有）：

```kotlin
tasks {
    publishPlugin {
        token.set(System.getenv("PUBLISH_TOKEN"))
    }
}
```

3. 执行发布：

```bash
JAVA_HOME=/path/to/goland/jbr ./gradlew publishPlugin
```

#### 审核要点

JetBrains 审核会检查：

- `plugin.xml` 中的 `<description>` 是否清晰描述功能
- `<change-notes>` 是否有版本更新说明
- 插件是否能正常安装和卸载
- 是否有安全问题（如网络请求、文件系统操作）

#### 更新已发布的插件

1. 修改代码 + 递增版本号
2. 构建新的 zip
3. 用上述任一方式上传新版本
4. JetBrains 审核通过后，用户的 GoLand 会提示更新

### 2.5 GoLand 兼容版本管理

当新版本 GoLand 发布后，需要更新 `build.gradle.kts`：

```kotlin
patchPluginXml {
    sinceBuild.set("243")    // 最低支持版本
    untilBuild.set("253.*")  // 最高支持版本，改为新版本号
}
```

同时需要将 `dependencies` 中的 `local` 路径指向对应版本的 GoLand：

```kotlin
intellijPlatform {
    local("/path/to/goland")  // 指向本地安装的 GoLand
}
```

## 3. 常见问题

### 构建报错 `Cannot resolve JBR`

确保 `JAVA_HOME` 指向 GoLand 自带的 JBR：

```bash
JAVA_HOME=/path/to/goland/jbr ./gradlew build
```

### ANTLR4 语法错误

修改 `Cztctl.g4` 后构建失败，Gradle 会输出 ANTLR4 的错误信息（如 rule conflict、token 未定义等），根据提示修复 `.g4` 文件即可。

### 插件安装后无高亮

1. 确认文件扩展名是 `.cron` 或 `.rabbitmq`
2. 检查 GoLand → Settings → Plugins 中插件是否已启用
3. 尝试 File → Invalidate Caches and Restart

### Live Templates 不生效

1. 确认文件类型被识别为 cztctl DSL（底部状态栏应显示 `cztctl DSL`）
2. 检查 Settings → Editor → Live Templates → cztctl 分组是否存在
