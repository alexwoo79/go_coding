package cmd

// displayValue 将空字符串显示为“<未设置>”。
func displayValue(v string) string {
	if v == "" {
		return "<未设置>"
	}
	return v
}
