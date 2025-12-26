package i18n

// zhTranslations contains all Chinese translations
var zhTranslations = map[string]string{
	// Root command
	"root.short":       "fastGit 是一个帮助您快速提交代码的命令行工具。",
	"root.description": "快速高效的 Git 工作流工具",

	// Version command
	"version.short":   "显示 fastGit 版本信息",
	"version.version": "版本：",
	"version.github":  "项目地址：",
	"version.about":   "了解更多关于我的信息，请访问：",

	// Status command
	"status.short": "显示 git 状态",

	// Push commands
	"push.all.short":      "推送所有更改到远程仓库",
	"push.selected.short": "选择并推送特定更改",

	// Remote commands
	"remotes.short": "管理 git 远程仓库",

	// Reset command
	"reset.short": "重置 git 仓库到特定状态",

	// Tag commands
	"tag.short":        "管理 git 标签",
	"tag.delete.short": "删除 git 标签",

	// Merge command
	"merge.short": "合并分支",

	// Cherry-pick command
	"cherry.pick.short": "从其他分支选择提交",

	// Branch commands
	"branch.delete.short":        "删除 git 分支",
	"branch.no.branches":         "仓库中未找到分支",
	"branch.delete.select":       "选择要删除的分支",
	"branch.delete.confirm":      "您确定要删除分支 '%s' 吗？",
	"branch.delete.cancelled":    "🚫 分支删除已取消。",
	"branch.delete.local":        "正在删除本地分支",
	"branch.delete.local.loading": "正在删除本地分支 %s...",
	"branch.delete.local.success": "本地分支 %s 删除成功",
	"branch.delete.remote.confirm": "您是否也想删除远程分支？",
	"branch.delete.remote":       "正在删除远程分支",
	"branch.delete.remote.loading": "正在删除远程分支 %s...",
	"branch.delete.remote.success": "远程分支 %s 删除成功",
	"branch.no.deletable.branches": "没有可删除的分支 (您不能删除当前所在分支)。",

	// Update command
	"update.short": "更新 fastGit 到最新版本",

	// Init command
	"init.short": "初始化新的 git 仓库",

	// Language settings
	"language.set.short":       "设置默认语言",
	"language.set.error":       "设置语言失败",
	"language.set.success":     "✅ 语言设置成功",
	"language.current.title":   "📝 当前语言设置",
	"language.current.active":  "当前激活",
	"language.current.database": "数据库设置",
	"language.current.not.set": "未设置",
	"language.available":       "可用语言",
	"language.invalid":         "无效的语言代码,请使用 'en' 或 'zh'",

	// Common messages
	"error.general":      "发生错误：",
	"error.not.git.repo": "不是 git 仓库",
	"success.general":    "操作成功完成",
	"confirm.continue":   "您想要继续吗？",
	"select.option":      "请选择一个选项：",
	"input.required":     "此字段为必填项",

	// Git specific
	"git.branch":           "分支：",
	"git.commit":           "提交：",
	"git.status.clean":     "工作目录干净",
	"git.status.modified":  "已修改文件：",
	"git.status.untracked": "未跟踪文件：",
	"git.push.success":     "成功推送到远程仓库",
	"git.push.failed":      "推送到远程仓库失败",

	// File operations
	"file.select":   "选择文件：",
	"file.selected": "已选择文件：",
	"file.none":     "未选择文件",
	"file.all":      "所有文件",

	// Progress messages
	"progress.pushing":  "正在推送更改...",
	"progress.fetching": "正在获取更新...",
	"progress.merging":  "正在合并分支...",
	"progress.loading":  "加载中...",
	"progress.complete": "完成！",

	// Form components
	"form.input.placeholder": "请输入...",
	"form.input.empty.error": "输入不能为空",
	"form.confirm.title":     "确认",
	"form.select.title":      "请选择一个选项",
	"form.multiselect.title": "请选择选项",

	// Git commands and operations
	"git.remotes.title":     "远程仓库：",
	"git.status.no_changes": "没有文件更改。",
	"git.status.title":      "文件状态：",

	// Reset command
	"reset.title":           "🔄 Git 重置",
	"reset.select.commit":   "选择要重置到的提交",
	"reset.select.mode":     "选择重置模式",
	"reset.mode.soft":       "软重置（保留暂存更改）",
	"reset.mode.mixed":      "混合重置（保留未暂存更改）",
	"reset.mode.hard":       "硬重置（丢弃所有更改）",
	"reset.confirm.title":   "重置确认",
	"reset.confirm.message": "重置到：%s\n模式：%s",
	"reset.confirm.warning": "⚠️  警告：硬重置将永久删除所有未提交的更改！",
	"reset.cancelled":       "🚫 重置操作已取消。",
	"reset.executing":       "正在执行 git reset...",
	"reset.success":         "✅ Git 重置成功完成！",

	// Reset - additional keys for implementation
	"reset.error.select.commit": "选择提交错误：",
	"reset.error.select.mode":   "选择重置模式错误：",
	"reset.mode.soft.label":     "Soft - 保留工作目录和暂存区",
	"reset.mode.mixed.label":    "Mixed - 保留工作目录，清空暂存区",
	"reset.mode.hard.label":     "Hard - 丢弃所有未提交的更改",
	"reset.mode.soft.desc":      " (保留全部)",
	"reset.mode.mixed.desc":     " (默认)",
	"reset.mode.hard.desc":      " (危险)",
	"reset.confirm.to":          "确认重置到",
	"reset.confirm.mode":        "模式",
	"reset.hard.warning":        "⚠️ 将丢失所有未提交更改！",
	"reset.executing.mode":      "正在重置 (%s)...",
	"reset.completed.to":        "已重置到 %s (%s)",
	"reset.success.prefix":      "重置完成 (HEAD → %s)",
	"reset.hint.soft":           "💡 更改已保留在暂存区",
	"reset.hint.mixed":          "💡 更改已保留在工作区",
	"reset.hint.hard":           "💡 所有未提交更改已丢弃",
	"reset.cancelled.msg":       "已取消",
	"reset.error.git.reset":     "执行git reset命令时出错：",

	// Tag operations
	"tag.create.title":     "🏷️  创建并推送标签",
	"tag.input.name":       "输入标签名称：",
	"tag.input.message":    "输入标签信息（可选）：",
	"tag.confirm.create":   "创建并推送标签 '%s'？",
	"tag.creating":         "正在创建标签...",
	"tag.pushing":          "正在推送标签到远程...",
	"tag.success":          "✅ 标签创建并推送成功！",
	"tag.delete.title":     "🗑️  删除标签",
	"tag.delete.select":    "选择要删除的标签",
	"tag.delete.confirm":   "您确定要删除标签 '%s' 吗？\n这将从本地和远程仓库中删除标签。",
	"tag.delete.cancelled": "🚫 标签删除已取消。",
	"tag.delete.local":     "正在删除本地标签",
	"tag.delete.remote":    "正在删除远程标签",
	"tag.delete.success":   "标签删除成功",
	"tag.get.error":        "获取标签错误：",

	// Push operations
	"push.all.title":      "🚀 推送所有更改",
	"push.selected.title": "📋 推送选定更改",
	"push.select.files":   "选择要推送的文件：",
	"push.no.changes":     "没有更改需要推送。",
	"push.preparing":      "准备推送...",
	"push.success":        "✅ 推送成功完成！",

	// Merge operations
	"merge.title":         "🔀 合并分支",
	"merge.select.branch": "选择要合并的分支：",
	"merge.confirm":       "将 '%s' 合并到当前分支？",
	"merge.executing":     "正在合并分支...",
	"merge.success":       "✅ 合并成功完成！",

	// Update operations
	"update.checking":         "正在检查更新...",
	"update.downloading":      "正在下载更新...",
	"update.installing":       "正在安装更新...",
	"update.success":          "✅ 更新成功完成！",
	"update.restart.required": "请手动重启 fastGit。",
	"update.windows.script":   "🔄 正在运行 Windows 更新脚本...",
	"update.script.success":   "更新脚本执行成功",
	"update.failed.script":    "运行安装脚本失败：",
	"update.unsupported":      "不支持的平台：",

	// Update operations - detailed
	"update.checking.version":   "🔍 正在检查最新版本...",
	"update.latest.version":     "📦 最新版本：%s",
	"update.downloading.asset":  "正在下载 %s...",
	"update.download.success":   "下载成功完成",
	"update.download.failed":    "使用 curl 和 wget 下载均失败：",
	"update.curl.failed":        "⚠️  curl 失败，尝试使用 wget...",
	"update.wget.downloading":   "正在使用 wget 下载 %s...",
	"update.extracting":         "📂 正在解压下载的文件...",
	"update.extract.failed":     "解压 zip 文件失败：",
	"update.installing.to":      "📥 正在安装到 %s...",
	"update.sudo.required":      "⚠️  需要管理员权限进行安装",
	"update.password.prompt":    "💡 您可能需要输入密码...",
	"update.direct.install":     "✓ 直接安装（权限充足）",
	"update.no.sudo.required":   "✓ 直接安装（无需 sudo）",
	"update.install.failed":     "安装二进制文件失败：",
	"update.restart.terminal":   "💡 请重启终端或运行 'source ~/.bashrc'（或相应命令）以使用更新版本。",
	"update.sudo.installing":    "🔐 正在使用 sudo 安装...",
	"update.copy.failed":        "复制二进制文件失败：",
	"update.permissions.failed": "设置权限失败：",

	// Update - missing keys for new implementation
	"update.running_windows_script":                "🔄 正在运行 Windows 更新脚本...",
	"update.downloading_running_script":            "正在下载并运行更新脚本...",
	"update.script_executed_success":               "更新脚本执行成功",
	"update.failed_run_script":                     "运行安装脚本失败",
	"update.complete_restart_manual":               "✅ 更新完成。请手动重启 fastGit。",
	"update.unsupported_platform":                  "不支持的平台",
	"update.checking_latest_version":               "🔍 正在检查最新版本...",
	"update.failed_get_latest_version":             "获取最新版本失败",
	"update.latest_version":                        "📦 最新版本",
	"update.failed_create_temp_dir":                "创建临时目录失败",
	"update.downloading_latest_release":            "正在下载最新版本",
	"update.download_completed":                    "下载成功完成",
	"update.curl_failed_try_wget":                  "⚠️  curl 失败，尝试使用 wget...",
	"update.downloading_with_wget":                 "正在使用 wget 下载 %s...",
	"update.failed_download_both":                  "使用 curl 和 wget 下载均失败",
	"update.extracting_file":                       "📂 正在解压下载的文件...",
	"update.failed_extract_zip":                    "解压 zip 文件失败",
	"update.installing_to":                         "📥 正在安装到",
	"update.root_permissions_required":             "⚠️  需要管理员权限进行安装",
	"update.password_prompt_hint":                  "💡 您可能需要输入密码...",
	"update.direct_install_sufficient_permissions": "✓ 直接安装（权限充足）",
	"update.direct_install_no_sudo":                "✓ 直接安装（无需 sudo）",
	"update.failed_install_binary":                 "安装二进制文件失败",
	"update.completed_successfully":                "🎉 更新成功完成！",
	"update.restart_terminal_hint":                 "💡 请重启终端或运行 'source ~/.bashrc'（或相应命令）以使用更新版本。",
	"update.installing_binary_system":              "正在安装二进制文件到系统目录",
	"update.installing_binary":                     "正在安装 fastGit 二进制文件...",
	"update.binary_installed_success":              "二进制文件安装成功",
	"update.setting_executable_permissions":        "正在设置可执行权限",
	"update.setting_permissions":                   "正在设置权限...",
	"update.permissions_set_success":               "权限设置成功",
	"update.installing_with_sudo":                  "🔐 正在使用 sudo 安装...",
	"update.installing_binary_sudo":                "正在使用 sudo 安装二进制文件...",
	"update.failed_copy_binary":                    "复制二进制文件失败",
	"update.failed_set_permissions":                "设置权限失败",

	// Update command descriptions
	"update.cmd.download":            "下载最新版本",
	"update.cmd.install":             "安装二进制文件到系统目录",
	"update.cmd.install.loading":     "正在安装 fastGit 二进制文件...",
	"update.cmd.install.success":     "二进制文件安装成功",
	"update.cmd.permissions":         "设置可执行权限",
	"update.cmd.permissions.loading": "正在设置权限...",
	"update.cmd.permissions.success": "权限设置成功",
	"update.cmd.sudo.install":        "正在使用 sudo 安装二进制文件...",
	"update.cmd.sudo.permissions":    "正在设置可执行权限...",

	// Update error messages
	"update.error.version":  "获取最新版本失败：",
	"update.error.temp.dir": "创建临时目录失败：",

	// Error messages
	"error.git.log":           "执行 git log 命令出错：",
	"error.file.status":       "获取文件状态失败：",
	"error.select.form":       "选择表单错误：",
	"error.command.execution": "执行命令出错：",
	"error.permission.denied": "权限被拒绝：",
	"error.file.not.found":    "文件未找到：",

	// Success messages
	"success.operation.complete": "🎉 所有操作成功完成！",
	"success.step.complete":      "步骤完成：",
	"success.file.saved":         "文件保存成功：",

	// Command execution
	"cmd.failed.step": "第 %d 步失败：%s",
	"cmd.command":     "命令：",
	"cmd.executing":   "正在执行：",

	// Time and date
	"time.created":  "创建时间：",
	"time.modified": "修改时间：",
	"time.format":   "01-02 15:04",

	// File status
	"status.modified":  "已修改",
	"status.added":     "已添加",
	"status.deleted":   "已删除",
	"status.untracked": "未跟踪",
	"status.unknown":   "未知",

	// Push operations - detailed
	"push.select.commit.type":   "选择提交类型",
	"push.input.commit.message": "输入提交信息：",
	"push.input.tag.name":       "输入标签名称：",
	"push.input.tag.message":    "输入标签信息：",
	"push.no.files.selected":    "未选择要推送的文件。",
	"push.files.adding":         "正在添加文件...",
	"push.committing":           "正在提交更改...",
	"push.to.remote":            "正在推送到远程...",

	// Form validation
	"validation.required": "此字段为必填项",
	"validation.invalid":  "输入无效",

	// Common actions
	"action.continue": "继续",
	"action.cancel":   "取消",
	"action.confirm":  "确认",
	"action.select":   "选择",

	// UI Components
	"ui.executing.commands": "正在执行命令...",
	"ui.progress":           "进度：",
	"ui.status":             "状态：",
	"ui.error.details":      "错误详情：",
	"ui.operation.success":  "操作成功完成！",
	"ui.operation.failed":   "操作失败",
	"ui.exiting.error":      "💡 退出以显示错误详情...",
	"ui.exiting.success":    "💡 正在退出...",
	"ui.step":               "第 %d 步：%s",

	// Spinner messages
	"spinner.fastgit.operation":  "🚀 FastGit 操作进行中...",
	"spinner.operation.complete": "操作成功完成！",
	"spinner.operation.failed":   "操作失败",
	"spinner.error.details":      "错误详情：%v",
	"spinner.elapsed.time":       "⏱️ 耗时：%v",
	"spinner.step.progress":      "步骤进度：",
	"spinner.loading":            "加载中",
	"spinner.pending":            "等待中",
	"spinner.success":            "成功",

	// Table operations
	"table.user.aborted": "用户中止操作",
	"table.no.selection": "未进行选择",

	// Git command operations - pushAll
	"git.add.all.description": "将所有文件添加到暂存区",
	"git.add.all.loading":     "正在添加文件...",
	"git.add.all.success":     "文件添加成功",
	"git.commit.description":  "使用提交信息创建提交",
	"git.commit.loading":      "正在创建提交...",
	"git.commit.success":      "提交创建成功",
	"git.pull.description":    "从远程拉取最新更改",
	"git.pull.loading":        "正在拉取更改...",
	"git.pull.success":        "拉取成功完成",
	"git.push.description":    "推送更改到远程仓库",
	"git.push.loading":        "正在推送到远程...",

	// Git command operations - pushSelected
	"push.selected.no.files":       "没有文件需要推送。",
	"push.selected.no.selection":   "未选择文件。",
	"git.add.selected.description": "将选定文件添加到暂存区",
	"git.add.selected.loading":     "正在添加选定文件...",
	"git.add.selected.success":     "选定文件添加成功",

	// Tag operations - detailed
	"tag.input.version":         "输入版本号：",
	"tag.input.commit.message":  "输入提交信息：",
	"tag.create.description":    "创建注释标签",
	"tag.create.loading":        "正在创建标签...",
	"tag.create.success":        "标签 %s 创建成功",
	"tag.push.description":      "推送标签到远程仓库",
	"tag.push.loading":          "正在推送标签到远程...",
	"tag.push.success":          "标签 %s 推送成功",
	"tag.no.tags":               "仓库中未找到标签",
	"tag.delete.local.loading":  "正在删除本地标签 %s...",
	"tag.delete.local.success":  "本地标签 %s 删除成功",
	"tag.delete.remote.loading": "正在删除远程标签 %s...",
	"tag.delete.remote.success": "远程标签 %s 删除成功",

	// Merge operations - detailed
	"merge.no.branches":                     "没有分支可以合并。",
	"merge.select.target":                   "选择要合并到当前分支的分支：",
	"merge.select.strategy":                 "选择合并策略：",
	"merge.success.message":                 "合并成功完成。",
	"merge.failed":                          "合并失败",
	"merge.starting":                        "开始使用 %s 策略合并 '%s'...",
	"merge.warning.dirty.working.directory": "⚠️  警告：工作目录中有未提交的更改。",
	"merge.confirm.continue.with.changes":   "是否仍要继续合并？",
	"merge.conflict.detected":               "🔀 检测到合并冲突！",
	"merge.conflict.instructions":           "💡 请手动解决冲突，然后运行 'git add <文件>' 和 'git commit'",
	"merge.fast.forward.failed":             "❌ 无法进行快进合并",
	"merge.fast.forward.suggestion":         "💡 尝试使用'非快进'策略或解决任何冲突",
	"merge.uncommitted.changes":             "❌ 您有未提交的更改将被覆盖",

	// Merge strategies - 合并策略
	"merge.strategy.default.name":        "默认",
	"merge.strategy.default.description": "默认合并行为",
	"merge.strategy.ff.only.name":        "仅快进",
	"merge.strategy.ff.only.description": "仅在可以快进合并时进行合并",
	"merge.strategy.no.ff.name":          "非快进",
	"merge.strategy.no.ff.description":   "始终创建合并提交",
	"merge.strategy.squash.name":         "压缩",
	"merge.strategy.squash.description":  "将所有提交压缩为单个提交",

	// Cherry-pick messages - 选择提交相关信息
	"cherry.pick.select.commits":      "选择要 cherry-pick 的提交：",
	"cherry.pick.select.option":       "选择 cherry-pick 选项：",
	"cherry.pick.no.commits.selected": "未选择任何提交进行 cherry-pick",
	"cherry.pick.success.commit":      "成功 cherry-pick 提交",
	"cherry.pick.success.all":         "✅ 所有提交已成功 cherry-pick！",
	"cherry.pick.executing":           "正在执行",
	"cherry.pick.progress":            "正在 cherry-pick 提交...",
	"cherry.pick.error.get.commits":   "获取提交失败",
	"cherry.pick.error.no.commits":    "未找到提交",
	"cherry.pick.error.execute":       "执行 cherry-pick 失败",

	// Cherry-pick conflict and error handling - Cherry-pick 冲突和错误处理
	"cherry.pick.conflict.detected":          "🔀 检测到 cherry-pick 冲突！",
	"cherry.pick.conflict.instructions":      "💡 请手动解决冲突，然后运行 'git add <文件>' 和 'git cherry-pick --continue'",
	"cherry.pick.conflict.output":            "Git 输出",
	"cherry.pick.conflict.resolution.needed": "Cherry-pick 冲突需要手动解决",
	"cherry.pick.empty.commit":               "⚠️  检测到空提交",
	"cherry.pick.empty.commit.error":         "无法 cherry-pick 空提交",
	"cherry.pick.already.applied":            "✅ 提交已应用",
	"cherry.pick.failed.output":              "Cherry-pick 失败，输出",
	"cherry.pick.failed.generic":             "Cherry-pick 失败",

	// Cherry-pick options - Cherry-pick 选项
	"cherry.pick.option.default.name":          "默认",
	"cherry.pick.option.default.description":   "标准 cherry-pick",
	"cherry.pick.option.no.commit.name":        "不提交",
	"cherry.pick.option.no.commit.description": "应用更改但不提交",
	"cherry.pick.option.edit.name":             "编辑",
	"cherry.pick.option.edit.description":      "提交前编辑提交信息",
	"cherry.pick.option.signoff.name":          "签名",
	"cherry.pick.option.signoff.description":   "在提交信息中添加 Signed-off-by 行",

	// Error messages - detailed
	"error.get.options":        "获取选项失败：",
	"error.get.file.status":    "获取文件状态失败",
	"error.multiselect.form":   "获取文件状态失败：",
	"error.select.form.detail": "选择分支出错：",
	"error.current.branch":     "获取当前分支失败：",
}
