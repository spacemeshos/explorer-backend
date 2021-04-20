package utils

import (
    "encoding/hex"
)

func BytesToAddressString(a []byte) string {
    unchecksummed := hex.EncodeToString(a[:])
    sha := sha3.NewKeccak256()
    sha.Write([]byte(unchecksummed))
    hash := sha.Sum(nil)

    result := []byte(unchecksummed)
    for i := 0; i < len(result); i++ {
        hashByte := hash[i/2]
        if i%2 == 0 {
            hashByte = hashByte >> 4
        } else {
            hashByte &= 0xf
        }
        if result[i] > '9' && hashByte > 7 {
            result[i] -= 32
        }
    }
    return "0x" + string(result)
}

func BytesToHex(a []byte) string { return "0x" + hex.EncodeToString(a[:]) }

func NBytesToHex(a []byte, n int) string { return "0x" + hex.EncodeToString(a[:n]) }
