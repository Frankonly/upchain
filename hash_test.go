package upchain

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHash(t *testing.T) {
	r := require.New(t)

	inputs := []string{"encoding/hex", "placeholder", "merkle placeholder", "5638d79f9ac9e896cf275a1d7b1a4b59324984775bb9316a801583f44d798a59",
		"SHA256是安全散列算法SHA系列算法之一，其摘要长度为256bits，即32个字节，故称SHA256。对于任意长度（按bit计算）的消息，SHA256都会产生一个32个字节长度数据，称作消息摘要。当接收到消息的时候，这个消息摘要可以用来验证数据是否发生改变，即验证其完整性。"}
	expects := []string{"5638d79f9ac9e896cf275a1d7b1a4b59324984775bb9316a801583f44d798a59",
		"4097889236a2af26c293033feb964c4cf118c0224e0d063fec0a89e9d0569ef2",
		"d33966c05481764d5bfea42d79177abad4d2d245e5d245b13f65a6ea020e5ba6",
		"a8145d86bde02e70958cc47f935369740c329275c92346cd7c667e97717c8afb",
		"38a443c3927fd1674608ff1ddea1a6ac7ffec27d876dc37d24a92ffe4eb6b6f1"}

	for i, input := range inputs {
		hash := Hash([]byte(input))
		hashString := hex.EncodeToString(hash)
		r.Equal(expects[i], hashString)
	}
}

func TestHashNodes(t *testing.T) {
	r := require.New(t)

	inputs := []struct {
		Left  string
		Right string
	}{{"4097889236a2af26c293033feb964c4cf118c0224e0d063fec0a89e9d0569ef2", "4097889236a2af26c293033feb964c4cf118c0224e0d063fec0a89e9d0569ef2"},
		{"a8145d86bde02e70958cc47f935369740c329275c92346cd7c667e97717c8afb", "38a443c3927fd1674608ff1ddea1a6ac7ffec27d876dc37d24a92ffe4eb6b6f1"},
		{"05fa996c850f3f4cdf8fb6b0c70dcd732b14231fc38c3bb8a9292e4d748e627b", "d33966c05481764d5bfea42d79177abad4d2d245e5d245b13f65a6ea020e5ba6"}}
	expects := []string{"05fa996c850f3f4cdf8fb6b0c70dcd732b14231fc38c3bb8a9292e4d748e627b", "2004c2e833cd983d2a04bd568dfa25df56f11ee036ef986b3abcd07073c67d11",
		"13f3e5d5e897edc61b7aa27297f13da00f878f7b404b7302e907a648b4f4ef49"}

	for i, input := range inputs {
		hash := HashNodes([]byte(input.Left), []byte(input.Right))
		hashString := hex.EncodeToString(hash)
		r.Equal(expects[i], hashString)
	}
}
