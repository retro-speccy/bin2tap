// Copyright © 2019 Marton Magyar

// SPDX-License-Identifier: MIT
// see https://spdx.org/licenses/

package tapfile

import (
	"math"
)

/*
TAP file format definitions.

The TAP file contains an arbitrary numbery of data. The data is organized into a data block preceded
by a header of an appropriate type. Each header and data blocks are preceeded by a Block_Length block.

Example for a BASIC program:
    Block_Length
    Block_BASICHeader
    Block_Length
    Block_Data (containing the BASIC program)

Example for a machine code program:
    Block_Length
    Block_ByteHeader
    Block_Length
    Block_Data (containing the machine code, e.g. the contents of a .bin-file)
*/

const tapBASICHeader uint8 = 0
const tapNumArrayHeader uint8 = 1
const tapStringArrayHeader uint8 = 2
const tapBytesHeader uint8 = 3

const tapHeaderBlock uint8 = 0
const tapDataBlock uint8 = 0xff

const tapBlockMaxLength uint16 = math.MaxUint16 - 2

const tapUnused uint16 = 32768

// length block containing the length of the following block
type Block_Length struct {
	length uint16 // length of the following data block, not counting this block
}

// program header or  program autostart header - for storing BASIC programs
type Block_BASICHeader struct {
	flag          uint8    // always 0. Byte indicating a standard ROM loading header
	datatype      uint8    // always 0: Byte indicating a program header
	filename      [10]byte // loading name of the program. filled with spaces (CHR$(32))
	datalength    uint16   // length of the following data (after the header) = length of BASIC program + variables
	autostartline uint16   // LINE parameter of SAVE command. Value 32768 means "no auto-loading"; 0..9999 are valid line numbers
	programlength uint16   // length of BASIC program; remaining bytes ([data length] - [program length]) = offset of variables
	checksum      uint8    // simply all bytes (including flag byte) XORed
}

// numeric data array header - for storing numeric arrays
type Block_NumArrayHeader struct {
	flag         uint8    // always 0. Byte indicating a standard ROM loading header
	datatype     uint8    // always 1: Byte indicating a numeric array
	filename     [10]byte // loading name of the program. filled with spaces (CHR$(32))
	datalength   uint16   // length of the following data (after the header) = length of number array * 5 +3
	unused1      uint8
	variablename uint8  // (1..26 meaning A..Z) +128
	unused2      uint16 // = 32768
	checksum     uint8  // simply all bytes (including flag byte) XORed
}

type Block_StringArrayHeader struct {
	flag         uint8    // always 0. Byte indicating a standard ROM loading header
	datatype     uint8    // always 2: Byte indicating an alphanumeric array
	filename     [10]byte // loading name of the program. filled with spaces (CHR$(32))
	datalength   uint16   // length of the following data (after the header) = length of string array +3
	unused1      uint8
	variablename uint8  // (1..26 meaning A..Z) +192
	unused2      uint16 // = 32768
	checksum     uint8  // simply all bytes (including flag byte) XORed
}

type Block_BytesHeader struct {
	flag         uint8    // always 0. Byte indicating a standard ROM loading header
	datatype     uint8    // always 3: Byte indicating a bytes header
	filename     [10]byte // loading name of the program. filled with spaces (CHR$(32))
	datalength   uint16   // length of the following data (after the header), in case of a SCREEN$ header = 6912
	startaddress uint16   // start address of the code in the Z80 address space, in case of a SCREEN$ header = 16384
	unused       uint16   // = 32768
	checksum     uint8    // simply all bytes (including flag byte) XORed
}

type Block_Data struct {
	flag      uint8  // always 255 indicating a standard ROM loading data block or any other value to build a custom data block
	datablock []byte // the essential data (may be empty)
	checksum  uint8  // simply all bytes (including flag byte) XORed
}

type TAP_BIN_File struct {
	headerlength  Block_Length
	header        Block_BytesHeader
	bindatalength Block_Length
	bindata       Block_Data
}
