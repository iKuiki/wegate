package common

// ForceString 强制把object转为string，失败返回空字符串
func ForceString(object interface{}) string {
	if r, ok := object.(string); ok {
		return r
	}
	return ""
}
