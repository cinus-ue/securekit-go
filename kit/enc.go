package kit

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/cinus-ue/securekit/kit/base"
	"github.com/cinus-ue/securekit/kit/suite"
	"github.com/cinus-ue/securekit/kit/suite/rsa"
)

const (
	pSizeLen = 8
	sktExt   = ".skt"
	aesAlgo  = suite.Aes256Ctr
)

var (
	SKTAESVersion = []byte{0x53, 0x4B, 0x54, 0x00, 0x02}
	SKTRSAVersion = []byte{0x53, 0x4B, 0x54, 0x01, 0x02}
	SKTRC4Version = []byte{0x53, 0x4B, 0x54, 0x02, 0x01}
)

func versionCheck(src io.Reader, versionRequirement []byte) error {
	version := make([]byte, len(versionRequirement))
	_, _ = src.Read(version)
	if !bytes.Equal(version, versionRequirement) {
		return errors.New("version mismatch error")
	}
	return nil
}

func beforeEncrypt(filepath string) (src, dest *os.File, err error) {
	src, err = os.Open(filepath)
	if err != nil {
		return
	}
	dest, err = os.Create(filepath + sktExt)
	return
}

func beforeDecrypt(filepath string, version []byte) (src, dest *os.File, err error) {
	src, err = os.Open(filepath)
	if err != nil {
		return
	}
	err = versionCheck(src, version)
	if err != nil {
		return
	}
	dest, err = os.Create(strings.TrimSuffix(filepath, sktExt))
	return
}

func closeFile(src, dest *os.File) {
	src.Close()
	dest.Close()
}

func deleteFile(file *os.File, delete bool) {
	if delete {
		os.Remove(file.Name())
	}
}

func RC4FileEncrypt(filepath string, passphrase []byte, delete bool) error {
	if path.Ext(filepath) == sktExt {
		return nil
	}
	src, dest, err := beforeEncrypt(filepath)
	if err != nil {
		return err
	}
	dest.Write(SKTRC4Version)
	dest.Write(SHA256(passphrase))
	err = suite.StreamEnc(src, dest, passphrase, suite.RC4)
	closeFile(src, dest)
	if err != nil {
		os.Remove(dest.Name())
		return err
	}
	deleteFile(src, delete)
	return nil
}

func RC4FileDecrypt(filepath string, passphrase []byte, delete bool) error {
	if path.Ext(filepath) != sktExt {
		return nil
	}
	src, dest, err := beforeDecrypt(filepath, SKTRC4Version)
	if err != nil {
		return err
	}
	hashSum := make([]byte, sha256.Size)
	src.Read(hashSum)
	if !bytes.Equal(SHA256(passphrase), hashSum) {
		dest.Close()
		os.Remove(dest.Name())
		return errors.New("wrong passphrase")
	}
	err = suite.StreamDec(src, dest, passphrase, suite.RC4)
	closeFile(src, dest)
	if err != nil {
		os.Remove(dest.Name())
		return err
	}
	deleteFile(src, delete)
	return nil
}

func AESFileEncrypt(filepath string, passphrase []byte, delete bool) error {
	if path.Ext(filepath) == sktExt {
		return nil
	}
	src, dest, err := beforeEncrypt(filepath)
	if err != nil {
		return err
	}
	dest.Write(SKTAESVersion)
	err = suite.StreamEnc(src, dest, passphrase, aesAlgo)
	closeFile(src, dest)
	if err != nil {
		os.Remove(dest.Name())
		return err
	}
	deleteFile(src, delete)
	return nil
}

func AESFileDecrypt(filepath string, passphrase []byte, delete bool) error {
	if path.Ext(filepath) != sktExt {
		return nil
	}
	src, dest, err := beforeDecrypt(filepath, SKTAESVersion)
	if err != nil {
		return err
	}
	err = suite.StreamDec(src, dest, passphrase, aesAlgo)
	closeFile(src, dest)
	if err != nil {
		os.Remove(dest.Name())
		return err
	}
	deleteFile(src, delete)
	return nil
}

func RSAFileEncrypt(filepath, keyfile string, delete bool) error {
	puk, err := ioutil.ReadFile(keyfile)
	if err != nil {
		return err
	}
	if path.Ext(filepath) == sktExt {
		return nil
	}
	src, dest, err := beforeEncrypt(filepath)
	if err != nil {
		return err
	}
	passphrase := base.GenerateRandomBytes(20)
	pbytes, err := rsa.RSAEncrypt(passphrase, puk)
	if err != nil {
		return err
	}
	psize := make([]byte, pSizeLen)
	binary.BigEndian.PutUint64(psize, uint64(len(pbytes)))
	dest.Write(SKTRSAVersion)
	dest.Write(psize)
	dest.Write(pbytes)
	err = suite.StreamEnc(src, dest, passphrase, aesAlgo)
	closeFile(src, dest)
	if err != nil {
		os.Remove(dest.Name())
		return err
	}
	deleteFile(src, delete)
	return nil
}

func RSAFileDecrypt(filepath, keyfile string, delete bool) error {
	prk, err := ioutil.ReadFile(keyfile)
	if err != nil {
		return err
	}
	if path.Ext(filepath) != sktExt {
		return nil
	}
	src, dest, err := beforeDecrypt(filepath, SKTRSAVersion)
	if err != nil {
		return err
	}
	psize := make([]byte, pSizeLen)
	src.Read(psize)
	pbytes := make([]byte, binary.BigEndian.Uint64(psize))
	src.Read(pbytes)
	passphrase, err := rsa.RSADecrypt(pbytes, prk)
	if err != nil {
		return err
	}
	err = suite.StreamDec(src, dest, passphrase, aesAlgo)
	closeFile(src, dest)
	if err != nil {
		os.Remove(dest.Name())
		return err
	}
	deleteFile(src, delete)
	return nil
}
