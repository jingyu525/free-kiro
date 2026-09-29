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

func TestLookupSHA256_FallbackToBinaryName(t *testing.T) {
	// When the SHA256SUMS file uses just the binary name, the lookup
	// should still match.
	sums := "deadbeef  free-kiro\n"
	got, err := LookupSHA256(sums, "free-kiro_0.1.0_linux_amd64.tar.gz")
	if err != nil {
		t.Fatal(err)
	}
	if got != "deadbeef" {
		t.Errorf("expected deadbeef, got %s", got)
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