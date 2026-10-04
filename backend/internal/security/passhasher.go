package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

type PasswordHash struct {
	memory       uint32
	iterations   uint32
	parallelisam uint8
	saltlength   uint32
	keylength    uint32
}

func Newpassword() *PasswordHash {
	return &PasswordHash{
		memory:       19 * 1024,
		iterations:   2,
		parallelisam: 1,
		saltlength:   16,
		keylength:    32,
	}
}

func (p *PasswordHash) Hash(passwrod string) (string, error) {

	salt := make([]byte, p.saltlength)

	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate the password salt: %w", err)
	}

	hash := argon2.IDKey(
		[]byte(passwrod),
		salt,
		p.iterations,
		p.memory,
		p.parallelisam,
		p.keylength,
	)

	saltencoded := base64.RawStdEncoding.EncodeToString(salt)
	hashencoded := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf(
		"$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		p.memory,
		p.iterations,
		p.parallelisam,
		saltencoded,
		hashencoded,
	), nil
}

func (p *PasswordHash) Verify(
	passwrod string,
	encodedHash string,
) (bool, error) {

	parts := strings.Split(encodedHash, "$")

	if len(parts) != 6 {
		return false, fmt.Errorf("invalid password hash format")
	}

	if parts[1] != "argon2id" {
		return false, fmt.Errorf("unsupported password hashing algorithm")
	}

	if parts[2] != "v=19" {
		return false, fmt.Errorf("unsupported Argon2 version")
	}

	parameters := strings.Split(parts[3], ",")

	if len(parameters) != 3 {
		return false, fmt.Errorf("invalid Argon2 parameters")
	}

	var memory uint32
	var iterations uint32
	var parallelisam uint8

	for _, parameter := range parameters {

		keyValue := strings.SplitN(parameter, "=", 2)

		if len(keyValue) != 2 {
			return false, fmt.Errorf("invalid Argon2 parameter")
		}

		switch keyValue[0] {

		case "m":
			value, err := strconv.ParseUint(keyValue[1], 10, 32)

			if err != nil {
				return false, fmt.Errorf("invalid Argon2 memory parameter: %w", err)
			}

			memory = uint32(value)

		case "t":
			value, err := strconv.ParseUint(keyValue[1], 10, 32)

			if err != nil {
				return false, fmt.Errorf("invalid Argon2 iteration parameter: %w", err)
			}

			iterations = uint32(value)

		case "p":
			value, err := strconv.ParseUint(keyValue[1], 10, 8)

			if err != nil {
				return false, fmt.Errorf("invalid Argon2 parallelism parameter: %w", err)
			}

			parallelisam = uint8(value)

		default:
			return false, fmt.Errorf("unknown Argon2 parameter")
		}
	}

	if memory == 0 || iterations == 0 || parallelisam == 0 {
		return false, fmt.Errorf("invalid Argon2 parameters")
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])

	if err != nil {
		return false, fmt.Errorf("failed to decode password salt: %w", err)
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])

	if err != nil {
		return false, fmt.Errorf("failed to decode password hash: %w", err)
	}

	actualHash := argon2.IDKey(
		[]byte(passwrod),
		salt,
		iterations,
		memory,
		parallelisam,
		uint32(len(expectedHash)),
	)

	if subtle.ConstantTimeCompare(actualHash, expectedHash) == 1 {
		return true, nil
	}

	return false, nil
}

func ValidatePassword(
	password string,
) error {

	if len(password) < 12 {
		return fmt.Errorf(
			"password must contain at least 12 characters",
		)
	}

	if len(password) > 128 {
		return fmt.Errorf(
			"password must not exceed 128 characters",
		)
	}

	return nil
}
