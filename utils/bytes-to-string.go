package utils

import (
	"encoding/hex"
)

func BytesToHex(a []byte) string { return "0x" + hex.EncodeToString(a[:]) }

func NBytesToHex(a []byte, n int) string { return "0x" + hex.EncodeToString(a[:n]) }
