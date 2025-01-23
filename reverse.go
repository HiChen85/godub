package googg

type SupportedReverseType interface {
	string | int | byte | float64
}

// Reverse 使用泛型定义一个反转切片的函数
func Reverse[T SupportedReverseType](s []T) []T {
	// 申请一个新的切片，长度和原切片相同
	r := make([]T, len(s))
	// 从后往前遍历原切片
	for i, v := range s {
		// 将原切片的元素放到新切片对应的位置
		r[len(s)-1-i] = v
	}
	// 返回新切片
	return r
}
