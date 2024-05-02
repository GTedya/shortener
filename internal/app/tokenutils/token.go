package tokenutils

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

const (
	mySecret         = "abc&1*~#^2^#s0^=)^^7%b34"
	UserIDCookieName = "user-id"
)

func GetUserID(r *http.Request) string {
	encodedCookie, err := r.Cookie(UserIDCookieName)
	if err != nil {
		return uuid.NewString()
	}

	decodedCookie, err := hex.DecodeString(encodedCookie.Value)
	if err != nil {
		return uuid.NewString()
	}
	decryptedUserID, err := Decrypt(string(decodedCookie), mySecret)
	if err != nil {
		return uuid.NewString()
	}

	return decryptedUserID
}

func AddEncryptedUserIDToCookie(w *http.ResponseWriter, userID string) error {
	encryptedUserID, err := Encrypt(userID, mySecret)
	if err != nil {
		return fmt.Errorf("encryption error: %w", err)
	}

	encodedCookieValue := hex.EncodeToString([]byte(encryptedUserID))

	http.SetCookie(
		*w,
		&http.Cookie{
			Name:  UserIDCookieName,
			Value: encodedCookieValue,
		},
	)
	return nil
}

func Decode(s string) []byte {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return data
}

func Decrypt(text, mySecret string) (string, error) {
	block, err := aes.NewCipher([]byte(mySecret))
	if err != nil {
		return "", fmt.Errorf("new cipher creating error: %w", err)
	}
	cipherText := Decode(text)
	cfb := cipher.NewCFBDecrypter(block, someBytes)
	plainText := make([]byte, len(cipherText))
	cfb.XORKeyStream(plainText, cipherText)
	return string(plainText), nil
}

func Encrypt(text, mySecret string) (string, error) {
	block, err := aes.NewCipher([]byte(mySecret))
	if err != nil {
		return "", fmt.Errorf("new cipher creating error: %w", err)
	}
	plainText := []byte(text)
	cfb := cipher.NewCFBEncrypter(block, someBytes)
	cipherText := make([]byte, len(plainText))
	cfb.XORKeyStream(cipherText, plainText)
	return Encode(cipherText), nil
}

var someBytes = []byte{35, 46, 57, 24, 85, 35, 24, 74, 87, 35, 88, 98, 66, 32, 14, 05}

func Encode(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}
