package validation

import (
    "path/filepath"
    "os"
)

func IsValidOutputFilePath(outputFilePath string) bool {
    if 0 == len(outputFilePath) {
        return false
    }

    // WARNING: DO NOT use os.IsExist() to check for dir existence, as special dir names like "." or ".." would be
    // reported as non-existing. ALWAYS USE !os.IsNotExist() instead.
    _, err := os.Stat(filepath.Dir(outputFilePath))
    return !os.IsNotExist(err)
}

func IsValidTcpListenPort(port int) bool {
    // TODO: Add stronger "port" arg validation i.e. it MUST be only within a [min, max] range.
    return port > 0
}
