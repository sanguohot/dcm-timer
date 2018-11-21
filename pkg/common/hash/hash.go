package hash

import (
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
)

func Sha256Hash(input []byte) (common.Hash) {
	hash := sha256.New()
	hash.Write(input)
	//h := sha256.New()
	//h.Write(input)
	//bs := h.Sum(nil)
	return common.BytesToHash(hash.Sum(nil))
}
func Md5(data []byte) string {
	has := md5.Sum(data)
	return  fmt.Sprintf("%x", has) //将[]byte转成16进制
}
