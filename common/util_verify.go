package common

import (
	"crypto/ecdsa"
	"crypto/elliptic"

	ecrypto "github.com/fletaio/fleta_testnet/common/crypto"
	"github.com/fletaio/fleta_testnet/common/hash"
)

type publicKey65 [65]byte

// RecoverPubkey recover the public key using the hash value and the signature
func RecoverPubkey(h hash.Hash256, sig Signature) (PublicKey, error) {
	var pubkey65 publicKey65
	if err := ecrypto.Ecrecover(h[:], sig[:], pubkey65[:]); err != nil {
		return PublicKey{}, err
	}
	X, Y := elliptic.Unmarshal(ecrypto.S256(), pubkey65[:])
	var pubkey PublicKey
	ecrypto.CompressPubkey(&ecdsa.PublicKey{
		Curve: ecrypto.S256(),
		X:     X,
		Y:     Y,
	}, pubkey[:])
	return pubkey, nil
}

// VerifySignature checks the signature with the public key and the hash value
func VerifySignature(pubkey PublicKey, h hash.Hash256, sig Signature) error {
	if !ecrypto.VerifySignature(pubkey[:], h[:], sig[:64]) {
		return ErrInvalidSignature
	}
	return nil
}

// ValidateSignaturesMajority validates signatures with the signed hash and checks majority
func ValidateSignaturesMajority(signedHash hash.Hash256, sigs []Signature, KeyMap map[PublicHash]bool) error {
	if len(sigs) != len(KeyMap)/2+1 {
		return ErrInsufficientSignature
	}
	sigMap := map[PublicHash]bool{}
	for _, sig := range sigs {
		pubkey, err := RecoverPubkey(signedHash, sig)
		if err != nil {
			return err
		}
		pubhash := NewPublicHash(pubkey)
		if !KeyMap[pubhash] {
			return ErrInvalidSignature
		}
		sigMap[pubhash] = true
	}
	if len(sigMap) != len(sigs) {
		return ErrDuplicatedSignature
	}
	return nil
}
