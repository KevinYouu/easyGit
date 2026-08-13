# update 更新功能优化

## 背景 / 问题

> 原 `easyGit update` 存在严重缺陷：安装路径硬编码导致更新无效、非原子安装可能丢失旧版本、无版本比较、HTTP 请求无超时无状态码检查、zip 解压存在路径穿越漏洞、下载无完整性校验、Windows 远程管道执行脚本、i18n 死键堆积。

## 设计决策

| 决策 | 方案 | 原因 |
|------|------|------|
| 安装目标 | `os.Executable()` + `EvalSymlinks` 解析当前二进制真实路径，失败回退默认目录 | 兼容 `make install` 到 GOPATH/bin 等自定义位置，保证新版真正生效 |
| 安装方式 | 目标目录内创建暂存文件 → 复制 → `chmod` → `os.Rename` 原子替换 | rename 成功前旧版本始终可用，失败可清理暂存文件不留残留 |
| 权限探测 | 直接尝试在目标目录 `CreateTemp`，失败自动切 sudo（cp→chmod→mv） | 消除原 `hasWritePermission` 的 TOCTOU 竞态 |
| 权限回退 | 目录可写但旧文件不可替换（粘滞位目录/root 拥有）时 rename 返回 EPERM，同样回退 sudo | 避免更新在最后一步失败且无法自动补救 |
| 版本比较 | 自实现数字段比较（忽略 v 前缀、`-dirty`、`-g<commit>` 后缀），预发布后缀（rc/beta）低于同数字段稳定版，同为预发布时按 semver 规则逐段比较（rc.2 > rc.1） | 无外部依赖；本地已是最新时跳过下载；本地预发布构建可正常升级到稳定版或更新预发布版 |
| HTTP 客户端 | 原生 `http.Client`（响应头超时 10s、整体传输超时 60s、User-Agent、状态码检查），替代 curl/wget 外部命令 | 可注入 baseURL 便于测试，避免外部工具依赖，慢速网络不中断下载 |
| 完整性校验 | 下载 `checksums.txt` 并比对 SHA256，不匹配立即中止；错误信息含实际与期望哈希 | goreleaser 已生成校验文件，防篡改/截断，便于诊断 |
| zip 解压 | `filepath.Clean` 后校验绝对路径、Windows 盘符路径与 `..` 前缀，拒绝非法条目；限制条目数量（1024）、单条目大小（200MB）、解压总大小（500MB） | 防 Zip Slip 路径穿越与 zip bomb 磁盘耗尽 |
| Windows 更新 | 下载 `install.ps1` 到临时文件后本地执行（`-ExecutionPolicy RemoteSigned`） | 替代 `iwr \| iex` 远程管道，脚本可审计；Go 下载器不写 MOTW，RemoteSigned 仍可执行本地文件 |
| 下载 UI | `command.RunFuncWithSpinnerOptions`（原生函数 + spinner），失败时由调用方输出具体原因 | 统一 UI 规范，避免出现无原因的错误行 |

## 接口 / 使用方式

```bash
easyGit update
```

流程（Unix）：检查最新版本 → 版本比较（已最新则退出）→ 下载 zip → 校验 SHA256 → 安全解压 → 解析安装路径 → 原子替换（无权限自动 sudo）。

错误均通过 `logs.Error` 输出（命令层），更新包内不直接打印裸错误。

## 已知限制

- 下载的校验文件与安装包同源（GitHub release），信任 GitHub 为前提
- Windows 更新仍依赖 PowerShell 环境；脚本从 `main` 分支拉取且不做 SHA256 校验（脚本与其下载的安装包均信任 GitHub），完整性校验仅覆盖 Unix 流程
- 无法解析的本地版本号（如开发版 `untracked`）视为已过期，每次 `update` 都会检查下载
- 权限不足时依赖 `sudo`（无密码交互式终端提示），无 sudo 环境下更新会失败并提示

## 变更历史

| 日期 | 描述 |
|------|------|
| 2026-08-13 | 初始版本：全部问题修复 |
| 2026-08-13 | 审查修复：预发布版本比较、EPERM 权限回退、HTTP 超时拆分、校验错误详情、zip bomb 防护、PowerShell 策略收紧 |
| 2026-08-13 | 二轮审查修复：Windows 盘符路径显式拒绝、sudo 失败路径清理 root 属主暂存文件、预发布后缀逐段比较、构建元数据误判修复、spinner panic 防护、updateUnixTo 失败路径测试 |
