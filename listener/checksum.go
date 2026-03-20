package main

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
)

func main() {
	addr := common.HexToAddress("0xb59e0050ac3449d8a9f7a40670ed86da7d89d5ac")
	fmt.Printf("\n[成功] 你的 EIP-55 格式地址为: %s\n\n", addr.Hex())
}
