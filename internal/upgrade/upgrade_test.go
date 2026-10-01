package upgrade

import (
	"bytes"
	"strings"
	"testing"
)

func TestLookupSHA256_ExactMatch(t *testing.T) {
	sums := "abc123  free-kiro_0.1.0_darwin_arm64.tar.gz\n" +
		"def456  free-kiro_0.1.0_darwin_amd64.tar.gz\n"
	got, err := LookupSHA256(sums, "free-kiro_0.1.0_darwin_arm64.tar.gz")
	if err != nil {
		t.Fatal(err)
	}
	if got != "abc123" {
		t.Errorf("expected abc123, got %s", got)
	}
}

// TestVerifyTarballSHA256 covers the path.Base + LookupSHA256 integration
// path that Apply uses after the upgrade-shasums-filename fix. Splitting
// this out of Apply lets the regression run without an HTTP Download.
// The negative case guards against accidentally reintroducing the
// "free-kiro" fallback that was deleted in the same fix.
func TestVerifyTarballSHA256(t *testing.T) {
	sums := []byte("abc123  free-kiro_0.8.0_darwin_arm64.tar.gz\n" +
		"def456  free-kiro_0.8.0_darwin_amd64.tar.gz\n" +
		"789abc  free-kiro_0.8.0_linux_arm64.tar.gz\n")

	cases := []struct {
		name       string
		tarballURL string
		wantHash   string
		wantErr    bool
	}{
		{
			name:       "darwin_arm64 happy path",
			tarballURL: "https://github.com/jingyu525/free-kiro/releases/download/v0.8.0/free-kiro_0.8.0_darwin_arm64.tar.gz",
			wantHash:   "abc123",
		},
		{
			name:       "darwin_amd64 happy path",
			tarballURL: "https://github.com/jingyu525/free-kiro/releases/download/v0.8.0/free-kiro_0.8.0_darwin_amd64.tar.gz",
			wantHash:   "def456",
		},
		{
			name:       "linux_arm64 happy path",
			tarballURL: "https://github.com/jingyu525/free-kiro/releases/download/v0.8.0/free-kiro_0.8.0_linux_arm64.tar.gz",
			wantHash:   "789abc",
		},
		{
			name:       "internal binary name rejected (regression guard)",
			tarballURL: "https://github.com/jingyu525/free-kiro/releases/download/v0.8.0/free-kiro",
			wantErr:    true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := verifyTarballSHA256(sums, tc.tarballURL)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error for URL %q, got hash %q", tc.tarballURL, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for URL %q: %v", tc.tarballURL, err)
			}
			if got != tc.wantHash {
				t.Errorf("hash mismatch for %q: got %q, want %q", tc.tarballURL, got, tc.wantHash)
			}
		})
	}
}

func TestLookupSHA256_NotFound(t *testing.T) {
	sums := "abc123  some-other-binary.tar.gz\n"
	_, err := LookupSHA256(sums, "free-kiro_0.1.0_darwin_arm64.tar.gz")
	if err == nil {
		t.Fatal("expected not-found error")
	}
	if !strings.Contains(err.Error(), "no SHA256") {
		t.Errorf("error message should mention 'no SHA256'; got %v", err)
	}
}

func TestLookupSHA256_HandlesBlankLines(t *testing.T) {
	// The SHA256SUMS format allows blank lines and comments; verify
	// we tolerate them.
	sums := "\n\nabc123  free-kiro_0.1.0_darwin_arm64.tar.gz\n\n"
	got, err := LookupSHA256(sums, "free-kiro_0.1.0_darwin_arm64.tar.gz")
	if err != nil {
		t.Fatal(err)
	}
	if got != "abc123" {
		t.Errorf("expected abc123, got %s", got)
	}
}

func TestFindAssets_PlatformSuffix(t *testing.T) {
	tarball, sums, binary := findAssets("v0.4.1")
	if !strings.Contains(tarball, "free-kiro_0.4.1_") {
		t.Errorf("tarball should embed version + platform; got %s", tarball)
	}
	if !strings.HasPrefix(tarball, "https://github.com/"+GitHubRepo+"/releases/download/v0.4.1/") {
		t.Errorf("tarball URL prefix wrong: %s", tarball)
	}
	if !strings.HasSuffix(tarball, ".tar.gz") {
		t.Errorf("tarball should end with .tar.gz; got %s", tarball)
	}
	if !strings.Contains(sums, "_SHA256SUMS") {
		t.Errorf("sums URL wrong: %s", sums)
	}
	if binary != "free-kiro" {
		t.Errorf("binary should be 'free-kiro'; got %s", binary)
	}
}

func TestBytesReader_EOFAndPartial(t *testing.T) {
	r := bytes.NewReader([]byte("hello"))
	buf := make([]byte, 3)
	n, err := r.Read(buf)
	if err != nil || n != 3 || string(buf) != "hel" {
		t.Errorf("first read: n=%d err=%v buf=%q", n, err, buf)
	}
	n, err = r.Read(buf)
	if err != nil || n != 2 || string(buf[:n]) != "lo" {
		t.Errorf("second read: n=%d err=%v buf=%q", n, err, buf[:n])
	}
	n, err = r.Read(buf)
	if err == nil || n != 0 {
		t.Errorf("third read should return EOF with n=0; got n=%d err=%v", n, err)
	}
	// io.EOF comparison is conventional; we just want to confirm.
	if err.Error() != "EOF" {
		t.Errorf("expected EOF sentinel; got %v", err)
	}
	// Use bytes to confirm the unused import doesn't break the build.
	_ = bytes.NewReader(nil)
}
