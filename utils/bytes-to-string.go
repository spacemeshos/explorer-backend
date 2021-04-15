package utils

import (
    "encoding/hex"
)

func BytesToAddressString(a []byte) string {
    return "0x" + hex.EncodeToString(a[:])
}

func BytesToHex(a []byte) string { return hex.EncodeToString(a[:]) }

func NBytesToHex(a []byte, n int) string { return hex.EncodeToString(a[:n]) }
