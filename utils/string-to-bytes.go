package utils

import (
    "encoding/hex"
)

func StringToBytes(s string) ([]byte, error) {
    if len(s) > 1 {
        if s[0:2] == "0x" || s[0:2] == "0X" {
            s = s[2:]
        }
    }
    if len(s)%2 == 1 {
        s = "0" + s
    }
    return hex.DecodeString(s)
}
