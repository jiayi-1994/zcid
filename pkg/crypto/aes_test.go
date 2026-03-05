package crypto

import (
	"testing"
)

func testKey() []byte {
	return []byte("01234567890123456789012345678901") // 32 bytes
}

func TestNewAESCrypto(t *testing.T) {
	_, err := NewAESCrypto(testKey())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestNewAESCrypto_ShortKey(t *testing.T) {
	_, err := NewAESCrypto([]byte("short"))
	if err == nil {
		t.Fatal("expected error for short key")
	}
}

func TestEncryptDecrypt(t *testing.T) {
	c, _ := NewAESCrypto(testKey())

	tests := []string{
		"hello world",
		"",
		"password123!@#",
		"中文测试",
		"a very long string that contains special characters: ~!@#$%^&*()_+-={}|[]\\:\";'<>?,./",
	}

	for _, plaintext := range tests {
		encrypted, err := c.Encrypt(plaintext)
		if err != nil {
			t.Fatalf("encrypt %q: %v", plaintext, err)
		}

		if plaintext != "" && encrypted == plaintext {
			t.Fatalf("encrypted should differ from plaintext for %q", plaintext)
		}

		decrypted, err := c.Decrypt(encrypted)
		if err != nil {
			t.Fatalf("decrypt %q: %v", plaintext, err)
		}

		if decrypted != plaintext {
			t.Fatalf("expected %q, got %q", plaintext, decrypted)
		}
	}
}

func TestEncrypt_DifferentCiphertexts(t *testing.T) {
	c, _ := NewAESCrypto(testKey())
	plaintext := "same input"

	enc1, _ := c.Encrypt(plaintext)
	enc2, _ := c.Encrypt(plaintext)

	if enc1 == enc2 {
		t.Fatal("same plaintext should produce different ciphertexts (random nonce)")
	}
}

func TestDecrypt_InvalidData(t *testing.T) {
	c, _ := NewAESCrypto(testKey())

	_, err := c.Decrypt("not-base64!!!")
	if err == nil {
		t.Fatal("expected error for invalid base64")
	}

	_, err = c.Decrypt("c2hvcnQ=") // "short" in base64
	if err == nil {
		t.Fatal("expected error for too-short data")
	}
}

func TestDecrypt_WrongKey(t *testing.T) {
	c1, _ := NewAESCrypto(testKey())
	c2, _ := NewAESCrypto([]byte("abcdefghijklmnopqrstuvwxyz012345"))

	encrypted, _ := c1.Encrypt("secret")
	_, err := c2.Decrypt(encrypted)
	if err == nil {
		t.Fatal("expected error when decrypting with wrong key")
	}
}
