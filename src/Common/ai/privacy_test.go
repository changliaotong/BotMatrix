package ai

import (
	"BotMatrix/common/types"
	"testing"
)

func TestPrivacyFilter(t *testing.T) {
	f := types.NewPrivacyFilter()
	ctx := types.NewMaskContext()

	input := "我的电话是 13800138000，邮箱是 test@example.com，IP 地址是 192.168.1.1。记得回电！"
	masked := f.Mask(input, ctx)
	t.Logf("Masked: %s", masked)

	if !contains(masked, "[PHONE_1]") {
		t.Errorf("Expected [PHONE_1] in masked text")
	}
	if !contains(masked, "[EMAIL_2]") {
		t.Errorf("Expected [EMAIL_2] in masked text")
	}
	if !contains(masked, "[IP_3]") {
		t.Errorf("Expected [IP_3] in masked text")
	}

	unmasked := f.Unmask(masked, ctx)
	t.Logf("Unmasked: %s", unmasked)

	if unmasked != input {
		t.Errorf("Unmasked text does not match original. Got: %s, Expected: %s", unmasked, input)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(substr) > 0 && (func() bool {
		for i := 0; i <= len(s)-len(substr); i++ {
			if s[i:i+len(substr)] == substr {
				return true
			}
		}
		return false
	})()))
}
