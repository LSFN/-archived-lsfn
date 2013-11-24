package common

type Variant struct {
	data []byte
}

func (v *Variant) ConnectByte(b byte) {
	v.data = append(v.data, b)
}

func (v *Variant) IsComplete() bool {
	return v.data[len(v.data)-1]&0x80 == 0
}

func (v *Variant) Uint64() uint64 {
	if v.data == nil {
		return 0
	}
	var value uint64
	var multiplier uint8
	for i := range v.data {
		value += uint64(v.data[i]&0x7f) << multiplier
		multiplier += 7
	}
	return value
}

func (v *Variant) FromUint64(value uint64) {
	for value > 0 {
		var nextByte byte = byte(value) & 0x7f
		value >>= 7
		if value > 0 {
			nextByte |= 0x80
		}
		v.data = append(v.data, nextByte)
	}
}

func (v *Variant) Reset() {
	v.data = nil
}

func (v *Variant) Bytes() []byte {
	result := make([]byte, len(v.data))
	copy(result, v.data)
	return result
}
