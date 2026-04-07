// os_fs.go provides OS filesystem helpers for the milestone package.
// Separated to allow build-tag-free testing.
// SPDX-License-Identifier: GPL-2.0-only

package milestone

import "os"

func readFileOS(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func writeFileOS(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}
