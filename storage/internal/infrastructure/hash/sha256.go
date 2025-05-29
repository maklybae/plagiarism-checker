package hash

import (
	"crypto/sha256"
	"hash"
)

type SHA256 struct{}

func NewSHA256() *SHA256 {
	return &SHA256{}
}

func (s *SHA256) Hash() hash.Hash {
	return sha256.New()
}
