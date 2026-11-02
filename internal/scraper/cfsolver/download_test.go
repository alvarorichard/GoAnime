package cfsolver

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

const (
	goodSig   = "signed-by-the-maintainer"
	testTag   = "cfsolve-abc"
	testAsset = "challenge-linux-amd64"
)

func sum(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }

// release serves a fake GoAnime release: binary plus SHA256SUMS. sums overrides
// the checksum line; hits counts requests.
func release(t *testing.T, binary []byte, sums string) (*Installer, *atomic.Int32) {
	t.Helper()
	var hits atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("/"+testTag+"/"+testAsset, func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		_, _ = w.Write(binary)
	})
	mux.HandleFunc("/"+testTag+"/"+sumsName, func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		_, _ = w.Write([]byte(sums))
	})
	mux.HandleFunc("/"+testTag+"/"+sumsName+sigSuffix, func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		_, _ = w.Write([]byte(goodSig))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return &Installer{
		BaseURL: srv.URL, Tag: testTag, Asset: testAsset,
		Dir: filepath.Join(t.TempDir(), "cfsolve"), Client: srv.Client(), Now: time.Now,
		Verify: func(_, sig []byte) error {
			if string(sig) != goodSig {
				return errors.New("bad signature")
			}
			return nil
		},
	}, &hits
}

func TestInstallVerifiesAndInstalls(t *testing.T) {
	bin := []byte("#!/bin/sh\necho hi\n")
	in, _ := release(t, bin, "deadbeef  other\n"+sum(bin)+"  "+testAsset+"\n")
	path, err := in.Install(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != string(bin) {
		t.Fatalf("installed %q, err=%v", got, err)
	}
	if st, _ := os.Stat(path); st.Mode().Perm()&0o100 == 0 {
		t.Errorf("not executable: %v", st.Mode())
	}
	left, _ := filepath.Glob(filepath.Join(in.Dir, "*.part"))
	if len(left) != 0 {
		t.Errorf("temp files left behind: %v", left)
	}
}

// A binary that does not match the published checksum must never land where
// Discover would run it.
func TestInstallRejectsChecksumMismatch(t *testing.T) {
	in, _ := release(t, []byte("evil"), sum([]byte("good"))+"  "+testAsset+"\n")
	if _, err := in.Install(context.Background()); err == nil || !strings.Contains(err.Error(), "checksum") {
		t.Fatalf("err = %v, want a checksum error", err)
	}
	entries, _ := os.ReadDir(in.Dir)
	for _, e := range entries {
		if e.Name() != failureMarker {
			t.Errorf("%s was left in the install folder", e.Name())
		}
	}
}

// The checksum list is only as trustworthy as its signature. A release that
// swaps both the binary and the list must not install.
func TestInstallRejectsAnUnsignedChecksumList(t *testing.T) {
	bin := []byte("evil")
	in, _ := release(t, bin, sum(bin)+"  "+testAsset+"\n")
	in.Verify = func(_, _ []byte) error { return errors.New("not signed by a trusted key") }
	if _, err := in.Install(context.Background()); err == nil || !strings.Contains(err.Error(), "signature") {
		t.Fatalf("err = %v, want a signature error", err)
	}
	entries, _ := os.ReadDir(in.Dir)
	for _, e := range entries {
		if e.Name() != failureMarker {
			t.Errorf("%s was installed from an unsigned list", e.Name())
		}
	}
}

func TestDefaultInstallerUsesTheEmbeddedTrustedKey(t *testing.T) {
	in, ok := DefaultInstaller()
	if !ok {
		t.Skip("no prebuilt cfsolve for this platform")
	}
	if in.Verify == nil {
		t.Fatal("the default installer must verify signatures")
	}
	if err := in.Verify([]byte("x"), []byte("not a signature")); err == nil {
		t.Fatal("garbage was accepted as a signature")
	}
}

func TestInstallFailsWhenSumsOmitTheAsset(t *testing.T) {
	in, _ := release(t, []byte("x"), sum([]byte("x"))+"  challenge-other\n")
	if _, err := in.Install(context.Background()); err == nil {
		t.Fatal("want an error")
	}
}

func TestInstallFailsOnMalformedChecksum(t *testing.T) {
	in, _ := release(t, []byte("x"), "zz  "+testAsset+"\n")
	if _, err := in.Install(context.Background()); err == nil {
		t.Fatal("want an error")
	}
}

func TestInstallBacksOffAfterAFailure(t *testing.T) {
	bin := []byte("ok")
	in, hits := release(t, bin, "bad  "+testAsset+"\n")
	if _, err := in.Install(context.Background()); err == nil {
		t.Fatal("first attempt should fail")
	}
	after := hits.Load()
	if _, err := in.Install(context.Background()); err == nil || hits.Load() != after {
		t.Fatalf("second attempt must not touch the network (err=%v, hits %d -> %d)", err, after, hits.Load())
	}
	// A day later it tries again, and a good release then installs.
	in.Now = func() time.Time { return time.Now().Add(retryAfter + time.Minute) }
	if _, err := in.Install(context.Background()); err == nil || hits.Load() == after {
		t.Fatalf("after the backoff it must retry (hits %d)", hits.Load())
	}
}

type noNetwork struct{ t *testing.T }

func (n noNetwork) RoundTrip(r *http.Request) (*http.Response, error) {
	n.t.Errorf("a request left the process: %s", r.URL)
	return nil, errors.New("blocked by test")
}

// A binary fetched over plain http could be swapped in transit, and the
// checksum list would travel the same way. Nothing may even be requested.
func TestInstallRefusesPlainHTTPToTheInternet(t *testing.T) {
	in := &Installer{BaseURL: "http://example.com/dl", Tag: "t", Asset: "a", Dir: t.TempDir(),
		Client: &http.Client{Transport: noNetwork{t}}, Now: time.Now}
	if _, err := in.Install(context.Background()); err == nil || !strings.Contains(err.Error(), "https") {
		t.Fatalf("err = %v", err)
	}
}

func TestPinnedSHAAndTag(t *testing.T) {
	sha := pinnedSHA()
	if len(sha) != 40 {
		t.Fatalf("cfsolve.lock sha = %q, want a full commit hash", sha)
	}
	if got := releaseTag(sha); got != "cfsolve-"+sha[:12] {
		t.Errorf("tag = %q", got)
	}
	if _, ok := assetName("linux", "amd64"); !ok {
		t.Error("linux/amd64 must have an asset")
	}
	if n, _ := assetName("windows", "amd64"); !strings.HasSuffix(n, ".exe") {
		t.Errorf("windows asset = %q", n)
	}
	if _, ok := assetName("plan9", "386"); ok {
		t.Error("unsupported platform reported an asset")
	}
}

func TestAutoDownloadOptOut(t *testing.T) {
	t.Setenv(envAutoDownload, "0")
	if AutoDownloadEnabled() {
		t.Error("GOANIME_CFSOLVE_AUTODOWNLOAD=0 must disable the download")
	}
	t.Setenv(envAutoDownload, "")
	if !AutoDownloadEnabled() {
		t.Error("the download is on by default")
	}
}
