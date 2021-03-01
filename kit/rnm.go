package kit

import (
	"encoding/base64"
	"os"
	"strings"

	"github.com/cinus-ue/securekit/kit/aes"
	"github.com/cinus-ue/securekit/kit/kvdb"
	"github.com/cinus-ue/securekit/kit/pass"
)

const RnmVersion = "SKTRNMV1"

func Rename(source string, passphrase []byte, db *kvdb.DataBase) error {
	fileName := GetFileName(source)
	if strings.HasPrefix(fileName, RnmVersion) {
		return nil
	}
	dk, salt, err := aes.DeriveKey(passphrase, nil, KeyLen)
	if err != nil {
		return err
	}

	ciphertext, err := aes.GCMEncrypt([]byte(fileName), dk, salt)
	if err != nil {
		return err
	}

	name := base64.URLEncoding.EncodeToString(ciphertext)
	id := RnmVersion + pass.GenerateRandomString(false, false, 20)

	err = os.Rename(source, GetBasePath(source)+id)
	if err != nil {
		return err
	}
	return db.Set(id, name)
}

func Recover(source string, passphrase []byte, db *kvdb.DataBase) error {
	id := GetFileName(source)
	if !strings.HasPrefix(id, RnmVersion) {
		return nil
	}
	if fileName, ok := db.Get(id); ok {
		ciphertext, err := base64.URLEncoding.DecodeString(fileName)
		if err != nil {
			return err
		}

		salt := ciphertext[len(ciphertext)-SaltLen:]
		dk, _, err := aes.DeriveKey(passphrase, salt, KeyLen)
		if err != nil {
			return err
		}

		plaintext, err := aes.GCMDecrypt(ciphertext, dk, salt)
		if err != nil {
			return err
		}

		fileName = GetBasePath(source) + string(plaintext)
		err = os.Rename(source, fileName)
		if err != nil {
			return err
		}
	}
	return nil
}