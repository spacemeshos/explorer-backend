package utils

import (
    "encoding/hex"
    "github.com/spacemeshos/go-spacemesh/crypto/sha3"
    "github.com/spacemeshos/go-spacemesh/common/util"
)

// Hex returns an EIP55-compliant hex string representation of the address.
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

func BytesToHex(a []byte) string { return util.Encode(a[:]) }

func NBytesToHex(a []byte, n int) string { return util.Encode(a[:n]) }
