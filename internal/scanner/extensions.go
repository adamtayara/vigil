package scanner

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type extManifest struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
	UpdateURL   string   `json:"update_url"`
}

func ScanExtensions() ([]Extension, error) {
	var all []Extension

	for _, b := range chromiumBrowsers() {
		exts := scanChromiumExtensions(b.name, b.profilePath)
		all = append(all, exts...)
	}

	ffExts := scanFirefoxExtensions()
	all = append(all, ffExts...)

	return all, nil
}

type browserDef struct {
	name        string
	profilePath string
}

func chromiumBrowsers() []browserDef {
	home, _ := os.UserHomeDir()
	var browsers []browserDef

	switch runtime.GOOS {
	case "windows":
		local := os.Getenv("LOCALAPPDATA")
		appdata := os.Getenv("APPDATA")
		candidates := []browserDef{
			{"Chrome", filepath.Join(local, "Google", "Chrome", "User Data")},
			{"Edge", filepath.Join(local, "Microsoft", "Edge", "User Data")},
			{"Brave", filepath.Join(local, "BraveSoftware", "Brave-Browser", "User Data")},
			{"Chromium", filepath.Join(local, "Chromium", "User Data")},
			{"Opera", filepath.Join(appdata, "Opera Software", "Opera Stable")},
		}
		browsers = candidates
	case "darwin":
		lib := filepath.Join(home, "Library", "Application Support")
		candidates := []browserDef{
			{"Chrome", filepath.Join(lib, "Google", "Chrome")},
			{"Edge", filepath.Join(lib, "Microsoft Edge")},
			{"Brave", filepath.Join(lib, "BraveSoftware", "Brave-Browser")},
		}
		browsers = candidates
	default:
		cfg := filepath.Join(home, ".config")
		candidates := []browserDef{
			{"Chrome", filepath.Join(cfg, "google-chrome")},
			{"Chromium", filepath.Join(cfg, "chromium")},
			{"Brave", filepath.Join(cfg, "BraveSoftware", "Brave-Browser")},
		}
		browsers = candidates
	}

	var active []browserDef
	for _, b := range browsers {
		if _, err := os.Stat(b.profilePath); err == nil {
			active = append(active, b)
		}
	}
	return active
}

func scanChromiumExtensions(browser, profileBase string) []Extension {
	var exts []Extension

	profiles := []string{"Default"}
	entries, err := os.ReadDir(profileBase)
	if err == nil {
		for _, e := range entries {
			if e.IsDir() && strings.HasPrefix(e.Name(), "Profile ") {
				profiles = append(profiles, e.Name())
			}
		}
	}

	for _, profile := range profiles {
		extDir := filepath.Join(profileBase, profile, "Extensions")
		ids, err := os.ReadDir(extDir)
		if err != nil {
			continue
		}
		for _, id := range ids {
			if !id.IsDir() {
				continue
			}
			ext := readChromiumExtension(browser, id.Name(), filepath.Join(extDir, id.Name()))
			if ext != nil {
				ext.ProfilePath = filepath.Join(profileBase, profile)
				exts = append(exts, *ext)
			}
		}
	}
	return exts
}

func readChromiumExtension(browser, id, extPath string) *Extension {
	versions, err := os.ReadDir(extPath)
	if err != nil || len(versions) == 0 {
		return nil
	}

	// Use the highest version dir
	versionDir := filepath.Join(extPath, versions[len(versions)-1].Name())
	manifestPath := filepath.Join(versionDir, "manifest.json")

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil
	}

	// Treat as untrusted input — only extract specific fields via JSON parsing
	var m extManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil
	}

	if m.Name == "" || strings.HasPrefix(m.Name, "__MSG_") {
		return nil
	}

	info, _ := os.Stat(manifestPath)
	var updatedAt time.Time
	if info != nil {
		updatedAt = info.ModTime()
	}

	source := "Chrome Web Store"
	if m.UpdateURL == "" {
		source = "Sideloaded (no update URL)"
	} else if !strings.Contains(m.UpdateURL, "clients2.google.com") {
		source = "Third-party source"
	}

	return &Extension{
		Browser:     browser,
		ID:          id,
		Name:        sanitize(m.Name),
		Version:     sanitize(m.Version),
		Description: sanitize(m.Description),
		Permissions: m.Permissions,
		Source:      source,
		UpdatedAt:   updatedAt,
	}
}

func scanFirefoxExtensions() []Extension {
	home, _ := os.UserHomeDir()
	var profileBase string
	switch runtime.GOOS {
	case "windows":
		profileBase = filepath.Join(os.Getenv("APPDATA"), "Mozilla", "Firefox", "Profiles")
	case "darwin":
		profileBase = filepath.Join(home, "Library", "Application Support", "Firefox", "Profiles")
	default:
		profileBase = filepath.Join(home, ".mozilla", "firefox")
	}

	profiles, err := os.ReadDir(profileBase)
	if err != nil {
		return nil
	}

	var exts []Extension
	for _, p := range profiles {
		if !p.IsDir() {
			continue
		}
		addonsPath := filepath.Join(profileBase, p.Name(), "extensions.json")
		data, err := os.ReadFile(addonsPath)
		if err != nil {
			continue
		}

		var addons struct {
			Addons []struct {
				DefaultLocale struct {
					Name        string `json:"name"`
					Description string `json:"description"`
				} `json:"defaultLocale"`
				ID          string   `json:"id"`
				Version     string   `json:"version"`
				Permissions []string `json:"permissions"`
				SourceURI   string   `json:"sourceURI"`
				UpdateDate  int64    `json:"updateDate"`
				Location    string   `json:"location"`
			} `json:"addons"`
		}

		if err := json.Unmarshal(data, &addons); err != nil {
			continue
		}

		for _, a := range addons.Addons {
			source := "Firefox Add-ons"
			if a.Location == "app-profile" && !strings.Contains(a.SourceURI, "addons.mozilla.org") {
				source = "Sideloaded"
			}
			exts = append(exts, Extension{
				Browser:     "Firefox",
				ID:          sanitize(a.ID),
				Name:        sanitize(a.DefaultLocale.Name),
				Version:     sanitize(a.Version),
				Description: sanitize(a.DefaultLocale.Description),
				Permissions: a.Permissions,
				Source:      source,
				UpdatedAt:   time.UnixMilli(a.UpdateDate),
				ProfilePath: filepath.Join(profileBase, p.Name()),
			})
		}
	}
	return exts
}

// sanitize removes control characters from user-provided strings to prevent display injection.
func sanitize(s string) string {
	s = strings.Map(func(r rune) rune {
		if r < 32 || r == 127 {
			return -1
		}
		return r
	}, s)
	if len(s) > 256 {
		s = s[:256]
	}
	return s
}
