package minio

import "testing"

func TestConf_Validate_OK(t *testing.T) {
	c := Conf{
		Endpoints:        []string{"127.0.0.1:9000"},
		AccessKeyID:      "admin",
		SecretAccessKey:  "password",
		SignatureVersion: SignatureV4,
		SlowThreshold:    1000,
	}
	if err := c.Validate(); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestConf_Validate_EmptyEndpoints(t *testing.T) {
	c := Conf{
		Endpoints:        nil,
		AccessKeyID:      "admin",
		SecretAccessKey:  "password",
		SignatureVersion: SignatureV4,
	}
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for empty endpoints")
	}
}

func TestConf_Validate_EmptyStringInEndpoints(t *testing.T) {
	c := Conf{
		Endpoints:        []string{"127.0.0.1:9000", ""},
		AccessKeyID:      "admin",
		SecretAccessKey:  "password",
		SignatureVersion: SignatureV4,
	}
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for empty string in endpoints")
	}
}

func TestConf_Validate_EmptyAccessKeyID(t *testing.T) {
	c := Conf{
		Endpoints:        []string{"127.0.0.1:9000"},
		AccessKeyID:      "",
		SecretAccessKey:  "password",
		SignatureVersion: SignatureV4,
	}
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for empty access key id")
	}
}

func TestConf_Validate_EmptySecretAccessKey(t *testing.T) {
	c := Conf{
		Endpoints:        []string{"127.0.0.1:9000"},
		AccessKeyID:      "admin",
		SecretAccessKey:  "",
		SignatureVersion: SignatureV4,
	}
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for empty secret access key")
	}
}

func TestConf_Validate_InvalidSignatureVersion(t *testing.T) {
	c := Conf{
		Endpoints:        []string{"127.0.0.1:9000"},
		AccessKeyID:      "admin",
		SecretAccessKey:  "password",
		SignatureVersion: "v3",
	}
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for invalid signature version")
	}
}

func TestConf_Validate_NegativeSlowThreshold(t *testing.T) {
	c := Conf{
		Endpoints:        []string{"127.0.0.1:9000"},
		AccessKeyID:      "admin",
		SecretAccessKey:  "password",
		SignatureVersion: SignatureV4,
		SlowThreshold:    -1,
	}
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for negative slow threshold")
	}
}

func TestConf_Validate_ZeroSlowThreshold(t *testing.T) {
	c := Conf{
		Endpoints:        []string{"127.0.0.1:9000"},
		AccessKeyID:      "admin",
		SecretAccessKey:  "password",
		SignatureVersion: SignatureV4,
		SlowThreshold:    0,
	}
	if err := c.Validate(); err != nil {
		t.Fatalf("expected no error for zero slow threshold (disabled), got: %v", err)
	}
}
