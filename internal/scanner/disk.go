package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	gopsdisk "github.com/shirou/gopsutil/v4/disk"
)

func ScanDisk() (DiskInfo, error) {
	info := DiskInfo{}

	parts, err := gopsdisk.Partitions(false)
	if err != nil {
		return info, fmt.Errorf("listing partitions: %w", err)
	}

	for _, p := range parts {
		usage, err := gopsdisk.Usage(p.Mountpoint)
		if err != nil {
			continue
		}
		info.Drives = append(info.Drives, DiskUsage{
			Path:        p.Mountpoint,
			Label:       p.Device,
			Total:       usage.Total,
			Used:        usage.Used,
			Free:        usage.Free,
			PercentUsed: usage.UsedPercent,
			FSType:      p.Fstype,
		})
	}

	homePath, _ := os.UserHomeDir()
	tempPath := os.TempDir()
	info.TempPath = tempPath

	scanRoot := homePath
	if scanRoot == "" {
		scanRoot = "/"
	}

	// Scan top-level dirs of home folder for size, max depth 3
	dirSizes := make(map[string]int64)
	var topFiles []DirSize
	const maxDepth = 3
	baseDepth := strings.Count(scanRoot, string(os.PathSeparator))

	filepath.Walk(scanRoot, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		depth := strings.Count(path, string(os.PathSeparator)) - baseDepth
		if fi.IsDir() {
			if shouldSkipDir(path) {
				return filepath.SkipDir
			}
			if depth >= maxDepth {
				return filepath.SkipDir
			}
			return nil
		}
		size := fi.Size()
		dirSizes[filepath.Dir(path)] += size
		if size > 50*1024*1024 { // only track files >50MB
			topFiles = append(topFiles, DirSize{Path: path, Bytes: size})
		}
		return nil
	})

	// Top dirs
	var dirs []DirSize
	for p, sz := range dirSizes {
		dirs = append(dirs, DirSize{Path: p, Bytes: sz})
	}
	sort.Slice(dirs, func(i, j int) bool { return dirs[i].Bytes > dirs[j].Bytes })
	if len(dirs) > 15 {
		dirs = dirs[:15]
	}
	info.TopDirs = dirs

	// Top files
	sort.Slice(topFiles, func(i, j int) bool { return topFiles[i].Bytes > topFiles[j].Bytes })
	if len(topFiles) > 20 {
		topFiles = topFiles[:20]
	}
	info.TopFiles = topFiles

	// Temp dir size
	var tempSize int64
	filepath.Walk(tempPath, func(_ string, fi os.FileInfo, err error) error {
		if err == nil && !fi.IsDir() {
			tempSize += fi.Size()
		}
		return nil
	})
	info.TempSize = tempSize

	return info, nil
}

func shouldSkipDir(path string) bool {
	lower := strings.ToLower(path)
	skipPatterns := []string{
		"windows\\system32", "windows\\syswow64", "windows\\winsxs",
		"program files\\windows", "$recycle.bin",
		"system volume information", ".git",
		"/proc", "/sys", "/dev",
	}
	if runtime.GOOS == "windows" {
		for _, pat := range skipPatterns {
			if strings.Contains(lower, strings.ToLower(pat)) {
				return true
			}
		}
	} else {
		for _, pat := range skipPatterns {
			if strings.HasPrefix(lower, pat) {
				return true
			}
		}
	}
	return false
}

func FormatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func FormatBytesU(b uint64) string {
	return FormatBytes(int64(b))
}
