// Пакет tokenutils предоставляет функции для работы с зашифрованными идентификаторами пользователей в виде куки.
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

// mySecret - это секретный ключ, используемый для шифрования и дешифрования.
const mySecret = "abc&1*~#^2^#s0^=)^^7%b34"

// UserIDCookieName - имя куки, в которой хранится зашифрованный идентификатор пользователя.
const UserIDCookieName = "user-id"

// GetUserID извлекает расшифрованный идентификатор пользователя из куки запроса.
// Если куки отсутствует или расшифровка не удалась, создается и возвращается новый UUID.
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

// AddEncryptedUserIDToCookie добавляет зашифрованный идентификатор пользователя в куки ответа.
// Идентификатор пользователя шифруется с использованием шифрования AES с заданным секретным ключом.
func AddEncryptedUserIDToCookie(w *http.ResponseWriter, userID string) error {
	encryptedUserID, err := Encrypt(userID, mySecret)
	if err != nil {
		return fmt.Errorf("ошибка шифрования: %w", err)
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

// Decode декодирует строку base64 в байтовый срез.
func Decode(s string) []byte {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return data
}

// Decrypt дешифрует заданный текст с использованием расшифровки AES с заданным секретным ключом.
func Decrypt(text, mySecret string) (string, error) {
	block, err := aes.NewCipher([]byte(mySecret))
	if err != nil {
		return "", fmt.Errorf("ошибка создания нового шифра: %w", err)
	}
	cipherText := Decode(text)
	cfb := cipher.NewCFBDecrypter(block, someBytes)
	plainText := make([]byte, len(cipherText))
	cfb.XORKeyStream(plainText, cipherText)
	return string(plainText), nil
}

// Encrypt шифрует заданный текст с использованием шифрования AES с заданным секретным ключом.
func Encrypt(text, mySecret string) (string, error) {
	block, err := aes.NewCipher([]byte(mySecret))
	if err != nil {
		return "", fmt.Errorf("ошибка создания нового шифра: %w", err)
	}
	plainText := []byte(text)
	cfb := cipher.NewCFBEncrypter(block, someBytes)
	cipherText := make([]byte, len(plainText))
	cfb.XORKeyStream(cipherText, plainText)
	return Encode(cipherText), nil
}

// someBytes используется в качестве вектора инициализации (IV) для шифрования AES.
var someBytes = []byte{35, 46, 57, 24, 85, 35, 24, 74, 87, 35, 88, 98, 66, 32, 14, 05}

// Encode кодирует байтовый срез в строку base64.
func Encode(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}
