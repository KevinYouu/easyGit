package update

import (
	"regexp"
	"strconv"
	"strings"
)

// gitDescribeSuffix 匹配 git describe 产生的提交计数后缀（如 30-ge693221、30-ge693221-dirty），
// 这类后缀表示同版本号的构建产物，不属于预发布
var gitDescribeSuffix = regexp.MustCompile(`^\d+-g[0-9a-f]+(-dirty)?$`)

// compareVersions 比较两个版本号大小（忽略 v 前缀与 -dirty、-g<commit> 等构建后缀）。
// 数字段相同且一方为预发布（rc/beta/alpha 等）时，预发布视为旧版本，
// 避免本地预发布构建被误判为已最新而跳过升级；同为预发布时按后缀逐段比较（rc.2 > rc.1）。
// 任一版本无法解析为纯数字段时返回 ok=false，调用方应保守处理（继续更新）。
func compareVersions(local, remote string) (result int, ok bool) {
	localParts, localOK := versionParts(local)
	remoteParts, remoteOK := versionParts(remote)
	if !localOK || !remoteOK {
		return 0, false
	}

	maxLen := max(len(remoteParts), len(localParts))
	for i := range maxLen {
		var localNum, remoteNum int
		if i < len(localParts) {
			localNum = localParts[i]
		}
		if i < len(remoteParts) {
			remoteNum = remoteParts[i]
		}
		if localNum < remoteNum {
			return -1, true
		}
		if localNum > remoteNum {
			return 1, true
		}
	}

	// 数字段相同：比较预发布状态与后缀
	// 如本地 0.2.5-rc.1 遇到远程 v0.2.5 时应判定为旧版本，允许升级到稳定版
	switch {
	case isPreRelease(local) && isPreRelease(remote):
		// 同为预发布时按后缀逐段比较（rc.1 < rc.2），相等才视为已最新
		return comparePreReleaseSuffixes(preReleaseSuffix(local), preReleaseSuffix(remote)), true
	case isPreRelease(local) != isPreRelease(remote):
		if isPreRelease(local) {
			return -1, true
		}
		return 1, true
	default:
		return 0, true
	}
}

// versionParts 将版本号解析为数字段数组，含非数字段或为空时返回 ok=false
func versionParts(v string) (parts []int, ok bool) {
	v = strings.TrimPrefix(v, "v")
	// 截断预发布/构建元数据与 git describe 后缀
	if idx := strings.IndexAny(v, "-+"); idx >= 0 {
		v = v[:idx]
	}

	rawParts := strings.Split(v, ".")
	result := make([]int, 0, len(rawParts))
	for _, rawPart := range rawParts {
		if rawPart == "" {
			return nil, false
		}
		num, err := strconv.Atoi(rawPart)
		if err != nil {
			return nil, false
		}
		result = append(result, num)
	}
	if len(result) == 0 {
		return nil, false
	}
	return result, true
}

// preReleaseSuffix 提取版本号 - 之后的预发布后缀，无则返回空串。
// 必须先剥离 + 构建元数据再按 - 切分：元数据可能含连字符（如 1.2.3+build-1），
// 若先按 - 切分会把构建元数据误判为预发布。
func preReleaseSuffix(v string) string {
	v = strings.TrimPrefix(v, "v")
	if before, _, ok := strings.Cut(v, "+"); ok {
		v = before
	}
	_, after, ok := strings.Cut(v, "-")
	if !ok {
		return ""
	}
	return after
}

// comparePreReleaseSuffixes 按 semver 规则比较两个预发布后缀：
// 点分标识符逐段比较，数字段按数值、字母段按字典序，数字标识符低于字母标识符，
// 前缀相同的短后缀低于长后缀（如 beta < beta.1、rc.1 < rc.2）。
func comparePreReleaseSuffixes(localSuffix, remoteSuffix string) int {
	localParts := strings.Split(localSuffix, ".")
	remoteParts := strings.Split(remoteSuffix, ".")

	maxLen := max(len(localParts), len(remoteParts))
	for i := range maxLen {
		var left, right string
		if i < len(localParts) {
			left = localParts[i]
		}
		if i < len(remoteParts) {
			right = remoteParts[i]
		}
		// 一方已无更多段：短后缀优先级更低（semver: 1.0.0-beta < 1.0.0-beta.1）
		if left == "" {
			return -1
		}
		if right == "" {
			return 1
		}

		leftNum, leftErr := strconv.Atoi(left)
		rightNum, rightErr := strconv.Atoi(right)
		switch {
		case leftErr == nil && rightErr == nil:
			if leftNum < rightNum {
				return -1
			}
			if leftNum > rightNum {
				return 1
			}
		case leftErr == nil:
			// 数字标识符优先级低于字母标识符（semver: 1.0.0-1 < 1.0.0-a）
			return -1
		case rightErr == nil:
			return 1
		default:
			if left < right {
				return -1
			}
			if left > right {
				return 1
			}
		}
	}
	return 0
}

// isPreRelease 判断版本是否带预发布后缀（rc/beta/alpha 等）。
// git describe 的提交计数后缀与 -dirty 视为同版本构建产物，不算预发布。
func isPreRelease(v string) bool {
	suffix := preReleaseSuffix(v)
	if suffix == "" || suffix == "dirty" {
		return false
	}
	return !gitDescribeSuffix.MatchString(suffix)
}
