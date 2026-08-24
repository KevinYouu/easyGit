# TODO

## 已完成:Amend 修改上次提交(am)

- [x] 三模式单选:仅改消息(预填原消息 + ↑/↓ 历史)/ 追加暂存文件(M/A/D 状态列,--no-edit 保消息)/ 两者都做 → `internal/gitcmd/amend.go`
- [x] 历史改写安全:已推送前置警告;完成后基于「改写前是否在远程」提供 force-with-lease 安全覆盖询问(无上游提示跳过)
- [x] 命令注册:`amend (am)` cobra + root 描述 + 主菜单(cp 之后);i18n `amend.*` 全量键(zh/en)
- [x] 测试:真实仓库集成回归(isHeadPushed 三态/upstream 解析/三种 amend 模式状态断言/流程注入含警告取消与拒绝强推)→ `make all` 全绿
- [x] 文档:docs/features/Amend修改提交.md + README/README-ZH 双语命令表
- [ ] 提交:代码 + i18n + 测试 + docs + README + TODO.md 单一 commit,不推送

## 已完成:Stash 管理(st)

- [x] 主菜单三选:保存(可选消息,form 新增 InputOptional 可空输入)/ 列表管理 / 清空全部 → `internal/gitcmd/stash.go`
- [x] 列表管理:消息+`stash@{N} · 相对时间` 双列展示,条目操作 diff 预览/apply(保留)/pop(删除)/drop;diff 与 drop 后回列表循环
- [x] 冲突安全:apply/pop 冲突明确提示 git 保留条目不丢失;drop 单条与 clear 全部均需确认(clear 展示条目数)
- [x] 命令注册:`stash (st)` cobra + root 描述 + 主菜单高频位(sw 之后);i18n `stash.*` 全量键(zh/en)
- [x] 测试:真实仓库集成回归(引用解析表驱动/列表顺序/apply-pop 行为差异/pop 冲突保留/drop-clear-show 状态断言/流程注入含 Esc 路径)→ `make all` 全绿
- [x] 文档:docs/features/Stash管理.md + README/README-ZH 双语命令表
- [ ] 提交:代码 + i18n + 测试 + docs + README + TODO.md 单一 commit,不推送

## 已完成:分支切换与新建(sw / bc)

- [x] sw 分支切换:本地(排除当前)+ 远程分支列表(NameDescColumns 双列,`/` 过滤),远程项自动 `--track` 创建本地跟踪分支 → `internal/gitcmd/branch_switch.go`
- [x] 脏工作区三选:携带修改 / 自动 stash 切换后 pop(冲突时 git 保留条目并提示手动处理)/ 取消;干净仓库跳过询问
- [x] bc 分支新建:名称表单层校验 refname(空格/`..`/`~ ^ : ? * [ \`/`.lock` 结尾等),基点可选当前 HEAD/本地/远程引用,确认后并行推送 `-u` 设 upstream
- [x] 命令注册:`branch-switch (sw)` / `branch-create (bc)` cobra 命令 + 主菜单高频位插入 + root 描述动态化
- [x] i18n:`branch.switch.*` / `branch.create.*` / `branch.label.*` 全量键(zh/en)
- [x] 测试:真实仓库集成回归(名称校验表驱动/列表去重规则/本地与远程切换/upstream 断言/stash 往返与冲突保留/流程注入),推送段同步执行器注入规避测试沙箱 TTY → `make all` 全绿
- [x] 文档:docs/features/分支切换与新建.md + README/README-ZH 双语命令表
- [ ] 提交:代码 + i18n + 测试 + docs + README + TODO.md 单一 commit,不推送

## 已完成:交互与流程优化批次(15 项,逐项独立提交)

- [x] P0 提交消息空主题校验:form 层 InputWithValidate + validateCommitMessage 拦截 `fix: ` 空消息(commit fe264cb)
- [x] P0 ps 文件列表加 M/A/D 状态列(ListFormColumns 双列,零新增步骤)(35ae9b2)
- [x] P0 清理 3 处硬编码英文 → i18n(squash/merge/rebase)(97dc470)
- [x] P0 pa/ps 交互统一:循环导航 ListFormWrap + IncrementUsage 成功后计数(1f3310f)
- [x] P0 README 修正 `--theme light commit` 错误示例(1381332)
- [x] P1 统一提交数据源 GetCommitsOptions(rs/sq/d 共用,删三处重复解析)(138b13d)
- [x] P1 列表 `/` 过滤:输入即筛选,大小写不敏感,支持粘贴/退格/循环导航(7bd0fe5)
- [x] P1 记忆上次选择:reset 模式/merge 策略/cp 选项 → settings 表预选(0e3f4f9)
- [x] P1 推送前 pull 可配置:config 新增「推送前拉取」,未设 upstream 自动跳过(b7d6681)
- [x] P1 提取 SelectAndSaveRemotes 收敛 5 处远程选择+持久化逻辑(8d7330d)
- [x] P1 非 git 仓库统一检查:root PersistentPreRunE(init/version/update/config 排除)(6e2ab43)
- [x] P2 交互式主菜单:easyGit 无参数进入,按使用频率排序(03e2cbc)
- [x] P2 提交消息历史:独立 recent_messages 表(去重置顶/超限截断),输入 ↑/↓ 复用最近 10 条(3d589c2 + 33f9658)
- [x] P2 git status 免 spinner(ps 响应提速)(222085e)
- [x] 文档:README/README-ZH 双语同步(主菜单/过滤/历史/推送前拉取/状态列)

## 已完成:配置中心新增冲突编辑器

## 已完成:配置中心新增冲突编辑器

- [x] 新增 `config` 配置项「冲突编辑器」:settings 表 `conflict_editor` 键,GetConflictEditor/SaveConflictEditor(空串=清除恢复自动)→ `internal/config/conflictEditor.go`
- [x] 子流程:单选 自动检测(默认)/vim/vi/nano/自定义…(输入任意命令如 code -w,留空恢复自动),预选当前值 → `cmd/easygit/commands/config_conflict_editor.go` + 分派键 `ConfigKeyConflictEditor`
- [x] 主列表摘要:未设置显示「当前: 自动检测(EDITOR → vim/vi/nano)」,已设置显示编辑器命令 → `BuildConfigOptions`
- [x] 生效:变基冲突菜单「打开编辑器解决冲突」优先级改为 配置 > $EDITOR/$VISUAL > vim/vi/nano → `internal/gitcmd/rebase.go`
- [x] 编辑器解析:已知异步编辑器(code/subl/atom)自动补 -w 等待标志(已带不重复);引号感知拆分支持 Windows 带空格路径;Windows 回退链追加 notepad(经 start /wait 逐个打开)→ `resolveConflictEditor`/`splitCommand`
- [x] i18n:`config.option/summary.conflict.editor.*` + `conflict.editor.*` 键(zh/en)
- [x] 测试:未设置返回空/保存/覆盖/清空/主列表含新项 → `make all` 全绿
- [x] 文档:配置中心.md(表格/主列表示例/子流程/实现位置)+ 变基冲突闭环.md + README 双语 + 测试用例.md
- [ ] 提交:代码 + i18n + 测试 + docs + README + TODO.md 单一 commit,不推送

## 已完成:变基冲突解决闭环(多冲突自动循环)

- [x] 冲突后不再退出:新增 `handleRebaseConflict` 冲突解决闭环,循环展示未合并文件(`git diff --diff-filter=U`)与操作菜单,直到变基完成/跳过/中止/退出 → `internal/gitcmd/rebase.go`
- [x] 菜单五项:打开编辑器解决冲突(EDITOR/VISUAL,回退 vim/vi/nano,无编辑器展示清单提示手动)/已解决继续(自动 git add + --continue)/跳过/中止/退出保持挂起;菜单函数可注入(测试驱动) → `rebaseConflictMenu`
- [x] 三入口统一接入:标准变基 `handleStandardRebase`、进行中恢复菜单 `handleInProgressRebase`、交互式变基 `RunInternalRebase`(squash/drop 共用)
- [x] continue/skip 后检查 `isRebaseInProgress()`:仍有冲突自动回到菜单(多提交多冲突一次搞定);`handleRebaseError` 冲突判定改为以 rebase-merge/rebase-apply 目录为主
- [x] 编辑器防挂起:`runGitRebaseContinue` 注入 `GIT_EDITOR=true`(非 TTY 下 git 也会启动编辑器,pick 接受原始信息、reword 接受工具内已注入消息)
- [x] i18n:`rebase.conflict.menu.*`/`rebase.conflict.files`/`rebase.abort.message` 等键(zh/en)
- [x] 测试:冲突检测/未合并文件/多冲突自动循环(核心回归,两次冲突依次解决)/skip/abort/quit 路径,菜单注入驱动真实 git 仓库 → `make all` 全绿
- [x] 文档:docs/features/变基冲突闭环.md + README 双语命令注释 + 测试用例.md 同步
- [ ] 提交:代码 + i18n + 测试 + docs + README + TODO.md 单一 commit,不推送

## 已完成:死代码清理(colors/spinner 包及测试专用导出)

- [x] 删除零引用包 `internal/colors`(仅自测使用)与 `internal/spinner`(全仓库无引用,command 包自带 spinner 实现)→ 同步删除 theme 中仅被 spinner 包使用的 5 个帧/图标函数
- [x] 删除仅测试引用的导出函数:`theme`(RenderSelection/RenderMuted/RenderBadge/GetCustomTheme/GetProgressBarStyle + 孤儿变量 MutedStyle)、`i18n.GetSupportedLanguages`、`gitcmd.SelectRemote/SelectBranch`(config 命令已改用 SelectRemoteWithConfig)、`update.NewReleaseClientWithBaseURL`、`command.NewProgressModelWithoutSpinner`、`form.NewListModel/NewListModelWrap/NewListModelColumns`
- [x] 测试同步:theme_test 删 9 个对应测试并重写 TestSelectorIndicators;form 测试 24 处改用私有 newListModel;update 测试 8 处改用私有 newReleaseClient;command 测试改用 NewProgressModel + showSpinner=false
- [x] i18n 死键:`git.select.remote`/`git.select.branch`(zh/en)已无引用,删除
- [x] 排查确认活代码:NewInputForm/NewConfirmForm(被 Input/Confirm 调用)、GetAllBranches/GetLatestTag 等(包内间接调用)、GetSpinnerFrames/GetHorizontalRule 等(command 包使用)均保留
- [x] 文档:docs/features/测试用例.md、docs/TEST_STRATEGY.md 移除 colors/spinner 条目 → `make all` 全绿
- [ ] 提交:代码 + i18n + 测试 + docs + TODO.md 单一 commit,不推送

## 已完成:主题双色板(默认自动 + 手动切换)

- [x] `internal/theme` 双色板:Neutral Dark / Neutral Light 两套令牌(primary/muted/border/selection/diff 背景,新增 SelectionMuted),`ApplyMode(auto|dark|light)` 切换并重建包级样式;`DetectDarkBackground`(lipgloss HasDarkBackground,OSC 11 查询,失败/非 TTY 回退深色);huh 主题 `ThemeBase(current.isDark)` 动态适配
- [x] 启动优先级:`--theme` 标志(运行时) > 配置中心设置(settings 表) > 自动检测 → `cmd/easygit/main.go` / `root.go`(注册 `--theme` 持久标志)
- [x] 配置中心新增「界面主题」项(auto/dark/light),保存后立即应用 → `internal/config/theme.go`(GetTheme/SaveTheme,ErrInvalidTheme 哨兵) / `cmd/easygit/commands/config_theme.go`
- [x] 清理硬编码色:`multi_input.go` 两处 `#d4d4d4` → theme.SelectionMuted;捕获色值的包级样式函数化(helpbar keyStyle/helpActionStyle、multi_input 7 个样式),主题切换即时生效
- [x] i18n:config.option/summary.theme.* + theme.option.*(zh/en)
- [x] 测试:浅色板精确色值断言、样式重建(渲染 SGR 含浅色主色)、dark/light 往返、ResolveMode/ValidMode、config Get/SaveTheme(含非法值 errors.Is)、Mode 与 config 常量防漂移、BuildConfigOptions 5 项 → `make all` 全绿
- [x] 文档:README/README-ZH 主题章节、docs/ui-shadcn-tui-refactor.md 双色板与模式机制
- [ ] 提交:代码 + i18n + 测试 + docs + TODO.md 单一 commit,不推送

## 已完成:多远程并行推送(pa/ps/tc/td/bd)

- [x] 进度模型并行段支持:`ProgressModel` 新增 `parallelFrom`/`inFlight`/`pendingStart`/`failedSteps`,
      `NewProgressModelParallel` / `RunMultipleCommandsParallel`(parallelFrom 越界回退全串行);
      串行段逐条推进,到达 parallelFrom 后剩余步骤经 `tea.Batch` 一次性并发启动(框架 BatchMsg 并发执行);
      并行段任一失败等待其余在飞步骤结束,进度按 completedCount 计数(完成消息乱序不跳变),
      状态行显示并行文案 `progress.executing.parallel` → `internal/command/progress_model.go`
- [x] 错误摘要多失败:failedSteps 逐条输出「步骤失败 + 命令」,errorMessage 仅收集命令输出(原跳过首末行 hack 删除)→ `printExecutionSummary`
- [x] pa/ps 推送段并行:add → commit → pull 串行,所有远程 push 一次性并行启动 → `pushAll.go` / `pushSelected.go`
- [x] 遗漏修复:tc(tag-create)/td(tag-delete)/bd(branch-delete)不再硬编码 origin,统一走 SelectRemoteWithConfig(配置持久化+多选),推送段并行 → `tag.go` / `branch.go`;无远程时降级仅本地操作
- [x] i18n:`progress.executing.parallel`(zh/en)
- [x] 测试:并行状态机(串行推进→并行段一次性启动→乱序完成→全部结束)、并行段部分失败收集、串行段失败立即终止、并行渲染(多 running spinner/失败 ✗/进度含失败)、摘要多失败;真实 git 集成验证(/tmp 裸仓库×2 并行推送,两个远程 HEAD 一致) → `make all` 全绿
- [x] docs/features/推送配置持久化.md 实现位置同步(并发真实落地),TODO 登记
- [ ] 提交:代码 + i18n + 测试 + docs + TODO.md 单一 commit,不推送

## 已完成:配置中心(config)统一所有可配置项

- [x] 新增 `easyGit config` 子命令:主列表(单选)列出全部可配置项,每项 Description 实时显示当前值摘要;选中进入子流程,完成后返回主列表循环,Esc 退出;语言切换后列表按新语言重建即时生效
- [x] 四个配置项:界面语言(复用原 set-language 逻辑)、推送配置(子菜单「设置/清除/返回」,clear 参数并入菜单)、提交类型(新增 AddCommitType/DeleteCommitTypes,删除显示 usage 使用次数,不允许删空)、标签版本上限(自绘三列表单:标题列/输入框/行尾弱化简介,预览行实时组合,↑/↓/j/k 导航,数字校验非负整数,prefix/suffix 可空)
- [x] 自适应多列布局(列数不硬编码):ColumnSpec 声明任意列,ColumnAuto 按最长内容自适应(上限截断)/ColumnFlex 占满剩余;名称列完整不截断、摘要列占满;Option.Cells 支持逐列单元格;极窄终端降级单列 → `internal/form/layout.go` / `list.go`(ListFormColumns/NewListModelColumns)
- [x] 提交类型错误处理审查修复:重复校验收敛到 config.AddCommitType(sentinel ErrCommitTypeExists + errors.Is),新增 error.add.commit.type/error.delete.commit.type 带 err;删除死键 push.config.help/push.config.change.hint/config.tag.patch.input.title
- [x] `InputSpec` 新增 `AllowEmpty`(跳过非空校验)与 `Validate`(自定义校验)字段,向后兼容 → `internal/form/input.go`
- [x] 移除一级命令 `set-language` / `set-push-config`,收敛到 config 单一入口 → `cmd/easygit/root.go`
- [x] 修复 TERM=dumb 管道脚本输入源 bug:accessible 模式表单统一 `WithInput(stdinBuf)` + `lineReader` 行级包装(防 huh 内部 Scanner 预读吞掉 MultiInput 后续字段输入),顺带修复既有 pa 列表→输入流程 → `internal/form/form.go` / `list.go`
- [x] 测试:commit_type CRUD(含 errors.Is)+ BuildConfigOptions + FormatPatch;InputSpec.AllowEmpty/Validate(pumpInit 模拟 Init 链);lineReader 顺序/完整性;CalculateAdaptiveColumns 宽度计算;config 主列表/三列/删除多选渲染用例;紧凑多输入表单(预览实时刷新 + 8-10 行矮终端完整渲染)→ `make all` 全绿
- [x] TERM=dumb 全流程回归:语言切换(预选标记)、推送配置设置/清除、提交类型添加(重复报错文案)/删除、标签版本上限表单编辑(逐字段输入/空行保留默认/非法数字拒绝)与数字校验;expect 真实 TUI 三列渲染(行尾简介弱化不混入标题) + 矮终端(8 行)表单 + 预览实时刷新 + ↑/↓/j/k 导航 + 退出无残影(统一 AltScreen)
- [x] 多输入表单组件统一:自绘 multi_input.go(不经 huh 渲染,简介行尾弱化不占额外行),tc 命令(tag.go)与配置中心版本号上限共用 form.MultiInput(specs, preview);删除 huh 版卡片式/紧凑主题与构造器
- [x] 多输入表单体验优化:输入列统一宽度(简介紧贴且列对齐)、循环导航(末字段 ↓/j 不误触保存)、聚焦行整行背景(逐段携带 SelectionBg 无"打洞",含行尾补白;测试断言背景色码/行宽/焦点跟随)
- [x] docs/features/配置中心.md + README 双语命令表同步 + TODO 登记
- [ ] 提交:代码 + i18n + 测试 + docs + README + TODO.md 单一 commit,不推送

## 已完成:MultiInput 卡片化 UI 优化(焦点/模糊统一容器、序号标题、描述行)

- [x] 新增 `GetMultiInputTheme`:所有字段统一带边框卡片(边框 + Padding(0,1)),焦点边框 `PrimaryColor` 高亮 + 标题加粗,模糊边框 `BorderColor` + 标题弱化;两态同尺寸,Enter 切换焦点零布局跳变(修复旧版焦点无框/模糊有框割裂)→ `internal/theme/theme.go`
- [x] 提示符统一 `❯`(`newInputField` 加 `Prompt("❯ ")`,与列表选中指示符一致);`InputSpec` 新增 `Desc` 字段(标题下弱化描述行);标题自动加标准数字序号前缀(`stepTitle`,如 `1.` `2.`)→ `internal/form/input.go`
- [x] 错误提示红色加粗 + ✗ 图标(ErrorMessage/ErrorIndicator SetString)→ 多输入主题
- [x] 帮助栏键值对分隔符 `  ` → ` · `(全局统一)→ `internal/form/helpbar.go`
- [x] i18n:标题去冒号精简(版本号 / 提交信息、Version / Commit message)+ 新增 `tag.input.version.desc`、`tag.input.commit.message.desc` → `internal/i18n/zh.go` / `en.go`
- [x] 调用方 `tag.go` 传入描述(与测试 specs 一致);渲染测试阈值重测:自然高度 13 行,≥13 完整断言、≥8 标题齐全、≥4 仅首字段 → `make all` 全绿
- [x] docs/features/单页多输入表单.md 渲染/测试/注意事项同步 + TODO 登记
- [ ] 提交:代码 + i18n + 测试 + docs + TODO.md 单一 commit,不推送

## 已完成:tc 连续输入表单堆叠修复(单页多输入 MultiInput)

- [x] 新增 `InputSpec` / `NewMultiInputForm` / `MultiInput`:所有字段放入单个 `huh.NewGroup`,单页堆叠渲染,单 tea 程序单渲染器消除主屏残留堆叠;空 specs 前置守卫返回空 → `internal/form/input.go`
- [x] 输入字段统一 `newInputField` 构造闭包(占位符 + 非空校验),`NewInputForm` 与 `NewMultiInputForm` 共用防漂移
- [x] `CreateAndPushTag` 改用一次 `form.MultiInput` 收集版本号 + 提交消息 → `internal/gitcmd/tag.go`
- [x] i18n 新增 `form.help.next`(继续/提交 / Next/Submit)、`form.help.prev`(上一步 / Back) → `internal/i18n/zh.go` / `en.go`
- [x] 帮助栏键位 `multiInputHelpKeys`:Enter 继续/提交、Shift+Tab 上一步、Esc 取消 → `internal/form/helpbar.go`
- [x] 测试:构造断言(空 specs/空值校验阻挡推进/Enter 推进、shift+tab 回退、末字段提交)+ 单帧渲染双标题 + 帮助栏渲染,`pumpForm` 跳过 `cursor.BlinkMsg` 防命令链死循环 → `make all` 全绿
- [x] docs/features/单页多输入表单.md + TODO 登记
- [ ] 提交:代码 + i18n + 测试 + docs/features + TODO.md 单一 commit,不推送

## 已完成:列表消息完整显示(移除固定上限与仓库层预截断)

- [x] `CalculateMessageWidth` 删除 `messageColMax=160` 封顶:消息列 = `max(width-fixed, 20)`,终端越宽显示越完整 → `internal/form/layout.go`
- [x] `reset.go` 移除仓库层预截断:提交列表标签与确认行均用完整消息,删除 `resetMessageMaxWidth` 常量,截断统一由列表组件按实际列宽处理 → `internal/gitcmd/reset.go`
- [x] 宽屏居中失效:消息列占满剩余宽度后表总宽=终端宽,300 宽终端不再居中(完整显示优先于居中)
- [x] 测试同步:TestCalculateMessageWidth 新增 240/300 用例(无上限证明);TestShouldCenterTable 300 改不居中;TestTableListLongMessageSingleLine longMsg 加长至 >300 列;新增 TestTableListLongMessageFullDisplay(300 宽完整显示) → `make all` 全绿
- [x] TERM=dumb 管道回归 rs:提交列表与确认行输出完整提交消息
- [x] docs/features 文档同步 + TODO 登记,提交

## 已完成:TUI 布局优化(分隔线 + ❯ 选中指示符)

- [x] 选中指示符统一 `❯`(显示宽 2,与 huh 默认 `> ` 同宽零位移):huh Select/MultiSelect 主题换字(`SelectSelector`/`MultiSelectSelector` = `"❯ "` + PrimaryColor + Bold)→ `internal/theme/theme.go`
- [x] 表单包装模型:标题下顶部线 + 帮助栏上底部线(≥6 行终端),`dividerAfterTitle` 随 `Update` 重建保留 → `internal/form/form.go`、`select.go`、`multiSelect.go`
- [x] 布局公式更新:指示列 2 宽(单选最左列)、表单高度模型 +2 行(`min(n+4, 终端)`)、表格多选预留 4 行/单选 2 行 → `internal/form/layout.go`
- [x] 表格单选:最左 1 宽指示列,`rebuildRows(cursorRow)` 让 ❯ 跟随光标 → `table_select.go`
- [x] 表格多选:指示列 + 3 宽复选框列两列分离(❯ 与 `[ ]` 间距由列填充与内边距产生,零位移),View 结构「标题 + 顶部线 + 行 + 底部线 + 帮助」→ `table_multi_select.go`
- [x] 进度屏:状态行下空行改分隔线,执行中帮助栏前插底部线(完成态保留 exiting 提示)→ `progress_model.go`
- [x] 渲染/布局/主题测试全量更新(高度公式、列数、`❯` 断言)→ `make all` 全绿
- [x] 终端尺寸矩阵 40/60/80/120/300 宽 × 6/10/24 高:分隔线位置正确、❯ 跟随光标、无折行/溢出
- [x] docs/features 功能文档 + 测试用例.md 增补 + README 双语同步,提交

## 已完成:统一帮助栏 + 选项单行说明

- [x] 帮助栏键位改 `[Esc]` 前缀式(主色加粗无背景),主题 Zinc Dark → Neutral Dark(Muted 400 / Border 700) → `internal/form/helpbar.go`、`internal/theme/theme.go`
- [x] 修复选中态背景断裂:OptionLabel 内嵌段剥离段尾 reset,说明段前复位字重(`\x1b[22m`),背景连续贯穿整行且说明保持常规字重 → `internal/form/helpbar.go`
- [x] TERM=dumb(accessible)时 selectOptionLabel 剥离样式,防无 reset 序列泄漏;`isAccessibleMode` 统一判定(与 huh 构造期耦合注释钉住) → `internal/form/select.go`、`form.go`
- [x] 回归测试:选中态背景连续(名称与说明之间无 reset)、说明段生效字重常规(状态机断言)、TERM=dumb 纯文本(含 huh accessible 真实路径) → `helpbar_test.go`、`form_test.go`

- [x] helpbar 组件(RenderHelpBar/AppendHelpBar/OptionLabel/HelpBarMinTermHeight) → `internal/form/helpbar.go`
- [x] 表单 4 构造器接线帮助栏,Form 包装模型统一生产与测试路径 → `internal/form/form.go`
- [x] 表格单选/多选帮助栏接线,单选高度预留 1 行 → `internal/form/table_select.go`、`layout.go`
- [x] 进度屏硬编码英文 i18n 化 + 帮助栏(完成态保留 exiting 提示) → `progress_model.go`、`spinnerCommand.go`
- [x] reset 模式表格 → 列表式单选表单,4 项单行说明(default 不传参数) → `internal/gitcmd/reset.go`
- [x] 选项说明统一机制:config.Option.Description + SelectForm 单行组装(merge/cherry-pick/rebase/language/commit type 优雅降级)
- [x] 渲染测试全量更新(高度模型 min(n+2, 终端)/4 行隐藏/两段式单行用例) → `make all` 全绿
- [x] docs/features 功能文档 + 测试用例.md 增补 + README 双语同步

## 已完成:渲染测试覆盖所有组件与命令

- [x] Input 命令级用例(pushAll/pushSelected/squash/tag×2) → `render_input_confirm_test.go`
- [x] Confirm 命令级用例(branch×2/drop/merge/reset/tag,含动态消息) → `render_input_confirm_test.go`
- [x] Select 补 merge target、rebase target → `render_command_test.go`
- [x] MultiSelect 补 remote 远端多选 → `render_command_test.go`
- [x] progress_model_test.go 补 View 渲染测试(进度条自适应/状态图标) → 顺带修复窄屏越界
- [x] docs/features/测试用例.md 同步渲染测试矩阵
- [x] make all 全绿并提交
