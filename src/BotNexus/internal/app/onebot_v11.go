package app

import (
	"BotMatrix/common/types"
	"regexp"
	"strings"
)

// CQ 码正则表达式
var cqCodeRegex = regexp.MustCompile(`\[CQ:([a-zA-Z0-9]+)((?:,[a-zA-Z0-9-_.]+=[^,\]]*)*)\]`)

// parseV11Message 将 OneBot v11 的字符串消息解析为内部标准的分段格式
func parseV11Message(raw string) []types.MessageSegment {
	var segments []types.MessageSegment
	lastIndex := 0

	// 查找所有 CQ 码
	matches := cqCodeRegex.FindAllStringSubmatchIndex(raw, -1)

	for _, match := range matches {
		// 添加 CQ 码之前的文本
		if match[0] > lastIndex {
			text := raw[lastIndex:match[0]]
			segments = append(segments, types.MessageSegment{
				Type: "text",
				Data: types.TextSegmentData{Text: unescapeV11(text)},
			})
		}

		// 解析 CQ 码
		cqType := raw[match[2]:match[3]]
		paramsStr := raw[match[4]:match[5]]
		params := make(map[string]any)

		if paramsStr != "" {
			// 去掉开头的逗号
			paramsStr = paramsStr[1:]
			paramPairs := strings.Split(paramsStr, ",")
			for _, pair := range paramPairs {
				kv := strings.SplitN(pair, "=", 2)
				if len(kv) == 2 {
					params[kv[0]] = unescapeV11(kv[1])
				}
			}
		}

		segments = append(segments, types.MessageSegment{
			Type: cqType,
			Data: params,
		})

		lastIndex = match[1]
	}

	// 添加剩余的文本
	if lastIndex < len(raw) {
		segments = append(segments, types.MessageSegment{
			Type: "text",
			Data: types.TextSegmentData{Text: unescapeV11(raw[lastIndex:])},
		})
	}

	return segments
}

// unescapeV11 处理 v11 的转义字符
func unescapeV11(s string) string {
	s = strings.ReplaceAll(s, "&#44;", ",")
	s = strings.ReplaceAll(s, "&#91;", "[")
	s = strings.ReplaceAll(s, "&#93;", "]")
	s = strings.ReplaceAll(s, "&amp;", "&")
	return s
}
