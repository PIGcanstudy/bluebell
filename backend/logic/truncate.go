package logic

import (
	"unicode"
	"unicode/utf8"
)

// 截断字符串，以单词为单位，最多保留maxWords个单词
func TruncateByWords(s string, maxWords int) string {
	processedWords := 0
	wordStarted := false
	for i := 0; i < len(s); {
		// 从字符串 s 的第 i 个字节开始，解码出第一个 Unicode 字符（rune）
		// 及其占用的字节数（width）
		r, width := utf8.DecodeRuneInString(s[i:])
		if !isSeparator(r) { // 不是分隔符
			i += width
			wordStarted = true
			continue
		}

		if !wordStarted { // 没有处理过单词
			i += width
			continue
		}

		// 是分隔符，并且前面已经处理过了单词
		wordStarted = false
		processedWords++
		if processedWords == maxWords {
			const ending = "..."
			if (i + len(ending)) >= len(s) {
				// 没超过字数直接返回
				return s
			}
			// 超过了就返回指定长度字符 + 3个省略号
			return s[:i] + ending
		}

		i += width
	}

	// Source string contains less words count than maxWords.
	return s
}

// 判断一个字符是否是分隔符
func isSeparator(r rune) bool {
	// ASCII alphanumerics and underscore are not separators
	if r <= 0x7F {
		switch {
		case '0' <= r && r <= '9':
			return false
		case 'a' <= r && r <= 'z':
			return false
		case 'A' <= r && r <= 'Z':
			return false
		case r == '_':
			return false
		}
		return true
	}
	// Letters and digits are not separators
	if unicode.IsLetter(r) || unicode.IsDigit(r) {
		return false
	}
	// Otherwise, all we can do for now is treat spaces as separators.
	return unicode.IsSpace(r)
}
