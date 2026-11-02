package cfsolver

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"io/fs"
	"os"
	"path"
	"path/filepath"
)

// The browser profile cfsolve impersonates ships inside GoAnime (cfsolve is MIT
// licensed; its notice sits beside the files), so nobody has to download or
// point at one. cfsolve reads profiles from disk, so the first use unpacks them.
//
//go:embed profile
var profileFS embed.FS

const profileEntry = "safari-26-ios.json"

// defaultProfilePath unpacks the embedded profile into the user cache and
// returns the path of its entry file. The folder name carries a digest of the
// contents, so a GoAnime that embeds a newer profile never reuses a stale one.
func defaultProfilePath() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		base = os.TempDir()
	}
	sum := sha256.New()
	err = fs.WalkDir(profileFS, "profile", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := profileFS.ReadFile(p)
		if err != nil {
			return err
		}
		sum.Write([]byte(p))
		sum.Write(b)
		return nil
	})
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, "goanime", "cfsolve-profile-"+hex.EncodeToString(sum.Sum(nil))[:12])
	entry := filepath.Join(dir, profileEntry)
	if _, err := os.Stat(entry); err == nil {
		return entry, nil
	}

	// Unpack beside the target and rename, so a crash never leaves a half-written
	// profile that the Stat above would accept.
	if err := os.MkdirAll(filepath.Dir(dir), 0o750); err != nil {
		return "", err
	}
	tmp, err := os.MkdirTemp(filepath.Dir(dir), "unpack-")
	if err != nil {
		return "", err
	}
	defer func() { _ = os.RemoveAll(tmp) }()
	err = fs.WalkDir(profileFS, "profile", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel := path.Clean(p[len("profile"):])
		if rel == "/" || rel == "." {
			return nil
		}
		out := filepath.Join(tmp, filepath.FromSlash(rel))
		if d.IsDir() {
			return os.MkdirAll(out, 0o750)
		}
		b, err := profileFS.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(out, b, 0o600)
	})
	if err != nil {
		return "", err
	}
	if err := os.Rename(tmp, dir); err != nil {
		if _, serr := os.Stat(entry); serr == nil {
			return entry, nil // another GoAnime unpacked it first
		}
		return "", err
	}
	return entry, nil
}
