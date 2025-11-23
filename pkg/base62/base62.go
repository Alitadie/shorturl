package base62

import (
	"errors"
	"strings"
)

// 打乱顺序的 62 进制字符集 (Shuffle过的)
// 为了演示 TDD 测试通过，我先用 "标准顺序" 的一部分或者你自定义的顺序
// 在这里，为了配合刚才测试用例 {0, "a"}, {1, "b"}，我使用从小写字母开始的标准 Base62 字符集
// 实际生产建议用脚本随机生成一个打乱的字符串: "09YZxya..."
const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const base = 62

func Encode(num uint64) string {
	if num == 0 {
		return string(charset[0])
	}
	var sb strings.Builder
	for num > 0 {
		rem := num % base
		sb.WriteByte(charset[rem])
		num /= base
	}

	// 此时 sb 里的字符串是反的 (如 62 是 "ab" 而不是 "ba")，需要反转
	// 为了简化代码，我们可以不反转，只要 Decode 对应即可。
	// 但为了人类阅读习惯，通常还是翻转过来。
	chars := []byte(sb.String())
	for i, j := 0, len(chars)-1; i < j; i, j = i+1, j-1 {
		chars[i], chars[j] = chars[j], chars[i]
	}
	return string(chars)
}

func Decode(token string) (uint64, error) {
	var num uint64
	for i := 0; i < len(token); i++ {
		char := token[i]
		idx := strings.IndexByte(charset, char)
		if idx == -1 {
			return 0, errors.New("invalid token")
		}
		num = num*base + uint64(idx)
	}
	return num, nil
}
