package testdata

import (
	"path/filepath"
	"runtime"
)

// basepath 是這個 package 的 root directory
var basepath string

func init() {
	_, currentFile, _, _ := runtime.Caller(0)
	basepath = filepath.Dir(currentFile)
}

// Path 會根據給的 relative file/directory path 回傳絕對路徑（absolute path）
// 如果原本就是絕對路徑，則不做事
func Path(rel string) string {
	if filepath.IsAbs(rel) {
		return rel
	}

	return filepath.Join(basepath, rel)
}
