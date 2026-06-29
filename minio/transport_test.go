package minio

import "testing"

func TestExtractBucket_Normal(t *testing.T) {
	tests := []struct {
		path   string
		expect string
	}{
		{"/bucket-name/key/subkey", "bucket-name"},
		{"/my-bucket/some/object.txt", "my-bucket"},
		{"/test/", "test"},
	}
	for _, tt := range tests {
		got := extractBucket(tt.path)
		if got != tt.expect {
			t.Errorf("extractBucket(%q) = %q, want %q", tt.path, got, tt.expect)
		}
	}
}

func TestExtractBucket_RootPath(t *testing.T) {
	got := extractBucket("/")
	if got != "" {
		t.Errorf("extractBucket('/') = %q, want empty", got)
	}
}

func TestExtractBucket_EmptyPath(t *testing.T) {
	got := extractBucket("")
	if got != "" {
		t.Errorf("extractBucket('') = %q, want empty", got)
	}
}

func TestExtractBucket_NoBucketOnly(t *testing.T) {
	// Path with only bucket, no trailing key
	got := extractBucket("/onlybucket")
	if got != "onlybucket" {
		t.Errorf("extractBucket('/onlybucket') = %q, want 'onlybucket'", got)
	}
}

func TestNewInstrumentedTransport(t *testing.T) {
	tr := newInstrumentedTransport("ep1:9000", nil, 1000)
	if tr == nil {
		t.Fatal("expected non-nil transport")
	}
	if tr.name != "ep1:9000" {
		t.Fatalf("expected name 'ep1:9000', got %q", tr.name)
	}
	if tr.slowThreshold != 1000 {
		t.Fatalf("expected slowThreshold=1000, got %d", tr.slowThreshold)
	}
	if tr.base == nil {
		t.Fatal("expected base transport to default to http.DefaultTransport")
	}
}
