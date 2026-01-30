package protocol

// CRC-16/IBM (CRC-16/ARC) lookup table
var crc16Table [256]uint16

func init() {
	const polynomial uint16 = 0xA001
	for i := 0; i < 256; i++ {
		crc := uint16(i)
		for j := 0; j < 8; j++ {
			if crc&1 != 0 {
				crc = (crc >> 1) ^ polynomial
			} else {
				crc >>= 1
			}
		}
		crc16Table[i] = crc
	}
}

// CalculateCRC16 computes CRC-16/IBM over the given data.
// Returns the CRC as a uint32 to match the 4-byte field in the Teltonika packet.
func CalculateCRC16(data []byte) uint32 {
	var crc uint16
	for _, b := range data {
		crc = (crc >> 8) ^ crc16Table[byte(crc)^b]
	}
	return uint32(crc)
}
