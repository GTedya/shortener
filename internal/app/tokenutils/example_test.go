package tokenutils

import (
	"fmt"
)

// ExampleEncryptDecrypt demonstrates how to use the Encrypt and Decrypt functions.
func ExampleEncryptDecrypt() {
	// Encrypt a user ID
	encryptedUserID, err := Encrypt("userID", "abc&1*~#^2^#s0^=)^^7%b34")
	if err != nil {
		fmt.Println("Encryption Error:", err)
		return
	}

	// Decrypt the encrypted user ID
	decryptedUserID, err := Decrypt(encryptedUserID, "abc&1*~#^2^#s0^=)^^7%b34")
	if err != nil {
		fmt.Println("Decryption Error:", err)
		return
	}

	fmt.Println("Decrypted UserID:", decryptedUserID)
	// Output: Decrypted UserID: userID
}
