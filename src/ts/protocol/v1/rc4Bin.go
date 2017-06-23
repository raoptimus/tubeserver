package v1

import "strconv"

type Rc4Bin struct {
	key  [256]uint32
	i, j uint8
}

type KeySizeError int

func (k KeySizeError) Error() string {
	return "rc4bin: invalid key size " + strconv.Itoa(int(k))
}

var bin2Hex = []byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f'}
var hex2bin = []uint8{
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 0, 0, 0, 0, 0, // 0-9
	0, 10, 11, 12, 13, 14, 15, 0, 0, 0, 0, 0, 0, 0, 0, 0, // A-F
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 10, 11, 12, 13, 14, 15, 0, 0, 0, 0, 0, 0, 0, 0, 0, // a-f
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
}

func NewRc4Bin(key []byte) (*Rc4Bin, error) {
	k := len(key)

	if k < 1 || k > 256 {
		return nil, KeySizeError(k)
	}

	var r Rc4Bin

	for i := 0; i < 256; i++ {
		r.key[i] = uint32(i)
	}

	var j uint8 = 0

	for i := 0; i < 256; i++ {
		j += uint8(r.key[i]) + key[i%k]
		r.key[i], r.key[j] = r.key[j], r.key[i]
	}

	return &r, nil
}

func (s *Rc4Bin) Crypt(src []byte) []byte {
	dst := make([]byte, len(src)*2)
	key := s.key
	i, j := s.i, s.j

	var ch uint32
	for l, v := range src {
		i += 1
		j += uint8(key[i])
		key[i], key[j] = key[j], key[i]
		ch = uint32(v) ^ key[uint8(key[i]+key[j])]

		dst[l*2] = bin2Hex[(ch&0xf0)>>4]
		dst[l*2+1] += bin2Hex[ch&0x0f]
	}

	return dst
}

func (s *Rc4Bin) Decrypt(src []byte) []byte {
	dst := make([]byte, len(src)/2)
	key := s.key
	i, j := s.i, s.j

	var ch uint32
	for l := 0; l < len(src)-1; l += 2 {
		v := hex2bin[uint32(src[l])]<<4 + hex2bin[uint32(src[l+1])]
		i += 1
		j += uint8(key[i])
		key[i], key[j] = key[j], key[i]
		ch = uint32(v) ^ key[uint8(key[i]+key[j])]
		dst[l/2] = byte(ch)
	}

	return dst
}
