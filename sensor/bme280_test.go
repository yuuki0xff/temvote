package sensor

import "testing"

func TestUint16ToInt16(t *testing.T) {
	var result int16

	result = Uint16ToInt16(0)
	if result != 0 {
		t.Errorf("should convert uint16(0) to int16(0), but result is %d", result)
	}

	result = Uint16ToInt16(0xffff)
	if result != -1 {
		t.Errorf("should convert uint16(0xffff) to int16(-1), but result is %d", result)
	}

	result = Uint16ToInt16(0x7fff)
	if result != 0x7fff {
		t.Errorf("should convert uint16(0x7fff) to int16(0x7fff), but result is %x", result)
	}
}
