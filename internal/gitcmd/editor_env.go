package gitcmd

import "strings"

// QuoteForEditorEnv 将可执行文件/脚本路径编码为 git 可解析的编辑器环境变量值
// (GIT_SEQUENCE_EDITOR / GIT_EDITOR)。git 把这两个变量的取值交给 shell 解释,
// 含空格或反斜杠的路径(如 C:\Program Files\easyGit\easyGit.exe)必须用双引号
// 包裹,否则路径会被截断或反斜杠被吞掉;已含引号的输入原样返回避免二次包裹。
func QuoteForEditorEnv(path string) string {
	if strings.ContainsAny(path, " \\") {
		if strings.HasPrefix(path, `"`) && strings.HasSuffix(path, `"`) {
			return path
		}
		return `"` + path + `"`
	}
	return path
}
