package domain

// Encryptor does encrypt a plain text into an encrypted text
type Encryptor interface {
	Encrypt(string) (string, error)
}

// Decryptor does decrypt an encrypted text into a plain text
type Decryptor interface {
	Decrypt(string) (string, error)
}

// Crypto is for encrypting a plain text into an encrypted string and vice versa
type Crypto interface {
	Encryptor
	Decryptor
}

type SensitiveInformation interface {
	Encrypt() error
	Decrypt() error
}

type SensitiveConfig interface {
	SensitiveInformation
	Validate() error
}
