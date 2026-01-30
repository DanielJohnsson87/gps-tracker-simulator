package protocol

import "encoding/binary"

// BuildAVLPacket constructs a complete Teltonika Codec 8 AVL data packet.
//
// Packet structure:
//
//	Preamble:         4 bytes (0x00000000)
//	Data length:      4 bytes (big-endian, length of data between length and CRC fields)
//	Codec ID:         1 byte  (0x08)
//	Number of records: 1 byte
//	AVL records:      variable
//	Number of records: 1 byte (repeated)
//	CRC-16:           4 bytes (CRC of data from Codec ID to second record count)
func BuildAVLPacket(records []AVLRecord) []byte {
	count := byte(len(records))

	// Build the data section (codec ID + count + records + count)
	var data []byte
	data = append(data, 0x08) // Codec ID
	data = append(data, count)
	for _, r := range records {
		data = append(data, r.Data...)
	}
	data = append(data, count)

	// Calculate CRC over the data section
	crc := CalculateCRC16(data)

	// Assemble full packet
	dataLen := make([]byte, 4)
	binary.BigEndian.PutUint32(dataLen, uint32(len(data)))

	crcBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(crcBytes, crc)

	var packet []byte
	packet = append(packet, 0x00, 0x00, 0x00, 0x00) // Preamble
	packet = append(packet, dataLen...)
	packet = append(packet, data...)
	packet = append(packet, crcBytes...)

	return packet
}
