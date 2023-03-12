package pgoldilocks

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"
)

type HashOut256 [4]uint64

var (
	// HashOut256Zero is used at Empty nodes
	HashOut256Zero = HashOut256{0, 0, 0, 0}
)

func (h HashOut256) Bytes() []byte {
	bytes := make([]byte, 32)
	binary.BigEndian.PutUint64(bytes[0:8], h[3])
	binary.BigEndian.PutUint64(bytes[8:16], h[2])
	binary.BigEndian.PutUint64(bytes[16:24], h[1])
	binary.BigEndian.PutUint64(bytes[24:32], h[0])
	return bytes
}

// MarshalText implements the marshaler for the HashOut256 type
func (h HashOut256) MarshalText() ([]byte, error) {
	return []byte(hex.EncodeToString(h.Bytes())), nil
}

func (h HashOut256) String() string {
	return hex.EncodeToString(h.Bytes())
}

func (h HashOut256) ElementString() string {
	return fmt.Sprintf("[%d,%d,%d,%d]", h[0], h[1], h[2], h[3])
}

func (h HashOut256) Hex() string {
	return hex.EncodeToString(h.Bytes())
}
func (h *HashOut256) Copy(x *HashOut256) {
	h[0] = x[0]
	h[1] = x[1]
	h[2] = x[2]
	h[3] = x[3]
}
func (h HashOut256) IsZeroHash() bool {
	return h[0] == 0 && h[1] == 0 && h[2] == 0 && h[3] == 0
}

func (h *HashOut256) UnmarshalText(b []byte) error {
	ha, err := NewHashOut256FromString(string(b))
	if err != nil {
		return err
	}
	copy(h[:], ha[:])
	return nil
}
func (h *HashOut256) IsInGoldilocksField() bool {
	return h[0] < 18446744069414584321 && h[1] < 18446744069414584321 && h[2] < 18446744069414584321 && h[3] < 18446744069414584321
}

func (h *HashOut256) Equals(h2 *HashOut256) bool {
	return h[0] == h2[0] && h[1] == h2[1] && h[2] == h2[2] && h[3] == h2[3]
}

func NewHashOut256FromBytes(b []byte) (*HashOut256, error) {
	if len(b) != 32 {
		return nil, fmt.Errorf("NewHashOut256FromBytes: input bytes must be of length 32")
	}
	h := &HashOut256{binary.BigEndian.Uint64(b[24:32]), binary.BigEndian.Uint64(b[16:24]), binary.BigEndian.Uint64(b[8:16]), binary.BigEndian.Uint64(b[0:8])}
	if !h.IsInGoldilocksField() {
		return nil, fmt.Errorf("NewHashOut256FromBytes: contains element not in goldilocks field")
	}
	return h, nil
}
func NewHashOut256FromBytes32(b [32]byte) HashOut256 {
	return HashOut256{binary.BigEndian.Uint64(b[24:32]), binary.BigEndian.Uint64(b[16:24]), binary.BigEndian.Uint64(b[8:16]), binary.BigEndian.Uint64(b[0:8])}
}
func NewHashOut256FromKnownSize(b []byte) HashOut256 {
	return HashOut256{binary.BigEndian.Uint64(b[24:32]), binary.BigEndian.Uint64(b[16:24]), binary.BigEndian.Uint64(b[8:16]), binary.BigEndian.Uint64(b[0:8])}
}
func NewHashOut256FromString(s string) (*HashOut256, error) {
	if strings.HasPrefix(s, "0x") {
		d, err := hex.DecodeString(s[2:])
		if err != nil {
			return nil, err
		}
		return NewHashOut256FromBytes(d)
	} else {
		d, err := hex.DecodeString(s)
		if err != nil {
			return nil, err
		}
		return NewHashOut256FromBytes(d)
	}
}
func NewHashOut256FromStringNoError(s string) *HashOut256 {

	r, err := NewHashOut256FromString(s)
	if err != nil {
		panic(err)
	}
	return r
}

func NewHashOut256FromUint64Array(d [4]uint64) HashOut256 {
	return HashOut256{d[0], d[1], d[2], d[3]}
}

func NewHashOut256FromUint64(x uint64) HashOut256 {
	return HashOut256{x, 0, 0, 0}
}

func NewHashOut256PtrFromUint64(x uint64) *HashOut256 {
	return &HashOut256{x, 0, 0, 0}
}

func NewHashOut256PtrFromUint64Last(x uint64) *HashOut256 {
	return &HashOut256{0, 0, 0, x}
}
