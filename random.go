package ippool

// compiler: could not import math/rand/v2 (no required module provides package "math/rand/v2")
// import "math/rand/v2"
import "math/rand"

// 使用 rand.Shuffle 洗牌
func Shuffle[T any](list []T) {
	rand.Shuffle(len(list), func(i, j int) {
		list[i], list[j] = list[j], list[i]
	})
}
