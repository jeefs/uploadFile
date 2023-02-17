package utils

import (
	"crypto/rand"
	"math/big"
	"os"
	"path/filepath"
	"time"
)

func Mkdir(basePath string) (string, error) {
	folderName := time.Now().Format("2006/01/02")
	folderPath := filepath.Join(basePath, folderName)
	//使用mkdirall会创建多层级目录
	err := os.MkdirAll(folderPath, os.ModePerm)
	if err != nil {
		return "", err
	}
	return folderPath, nil
}

var defaultPool = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func RandString(n int, pool []rune) string {
	if len(pool) == 0 {
		pool = defaultPool
	}
	b := make([]rune, n)
	for i := range b {
		l := len(pool)
		randN, _ := rand.Int(rand.Reader, big.NewInt(int64(l)))
		b[i] = pool[randN.Int64()]
	}
	return string(b)
}
