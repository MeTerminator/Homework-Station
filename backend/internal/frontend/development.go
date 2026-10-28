//go:build dev

package frontend

import "io/fs"

// Development uses Vite directly and does not require a frontend build.
func Files() fs.FS { return nil }
