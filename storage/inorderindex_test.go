package storage

import (
	"fmt"
	"testing"
)

func TestFromIndexOnLevel(t *testing.T) {
	v := FromIndexOnLevel(3, 3)
	if v == 55 {
		fmt.Println("奥利给！")
	}
}

func TestFromLeafIndex(t *testing.T) {
	v := FromLeafIndex(3)
	if v == 6 {
		fmt.Println("奥利给！")
	}
}

func TestFromPostorder(t *testing.T) {
	v := FromPostorder(125)
	if v == 95 {
		fmt.Println("奥利给！")
	}
}

func TestPostorder(t *testing.T) {

}
