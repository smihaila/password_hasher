package utils

import (
    "crypto/sha512"
    "encoding/base64"
)

func HashString(strData string) string {
    sha512Digest := sha512.Sum512([]byte(strData))
    return base64.StdEncoding.EncodeToString(sha512Digest[:])
}
