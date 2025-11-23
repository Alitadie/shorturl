package base62

import (
	"testing"
)

// 测试编码：数字 -> 字符串
func TestEncode(t *testing.T) {
	// 定义测试用例 (Table Driven Tests)
	tests := []struct {
		input    uint64
		expected string
	}{
		{0, "a"},   // 假设我们的字母表是以 'a' 开头的
		{1, "b"},   // a->0, b->1...
		{61, "9"},  // 最后一位
		{62, "ba"}, // 进位测试
	}

	for _, tt := range tests {
		got := Encode(tt.input)
		if got != tt.expected {
			t.Errorf("Encode(%d) = %s; want %s", tt.input, got, tt.expected)
		}
	}
}

// 测试解码：字符串 -> 数字
func TestDecode(t *testing.T) {
	tests := []struct {
		input    string
		expected uint64
	}{
		{"a", 0},
		{"ba", 62},
	}

	for _, tt := range tests {
		got, err := Decode(tt.input)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if got != tt.expected {
			t.Errorf("Decode(%s) = %d; want %d", tt.input, got, tt.expected)
		}
	}
}
