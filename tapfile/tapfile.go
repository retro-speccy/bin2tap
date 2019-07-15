// Copyright © 2019 Marton Magyar

// SPDX-License-Identifier: MIT
// see https://spdx.org/licenses/

package tapfile

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"strconv"
	"strings"
)

/*
TAP file format definitions.
See: https://web.archive.org/web/20110711141601/http://www.zxmodules.de:80/fileformats/tapformat.html
See: https://faqwiki.zxnet.co.uk/wiki/TAP_format

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

//const tapBASICHeader uint8 = 0
//const tapNumArrayHeader uint8 = 1
//const tapStringArrayHeader uint8 = 2
const tapBytesHeader uint8 = 3

const tapHeaderBlock uint8 = 0
const tapDataBlock uint8 = 0xff

const tapBlockMaxLength uint16 = math.MaxUint16 - 2

const tapUnused uint16 = 32768

// length block containing the length of the following block
type Block_Length struct {
	length uint16 // length of the following data block, not counting this block
}

/*
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
    unused       uint8
    variablename uint8  // (1..26 meaning A..Z) +128
    unused       uint16 // = 32768
    checksum     uint8  // simply all bytes (including flag byte) XORed
}

type Block_StringArrayHeader struct {
    flag         uint8    // always 0. Byte indicating a standard ROM loading header
    datatype     uint8    // always 2: Byte indicating an alphanumeric array
    filename     [10]byte // loading name of the program. filled with spaces (CHR$(32))
    datalength   uint16   // length of the following data (after the header) = length of string array +3
    unused       uint8
    variablename uint8  // (1..26 meaning A..Z) +192
    unused       uint16 // = 32768
    checksum     uint8  // simply all bytes (including flag byte) XORed
}
*/
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

type BINdata struct {
	filename     [10]byte // loading name of the program. filled with spaces (CHR$(32))
	datablock    []byte   // the essential data (may be empty)
	startaddress uint16   // start address of the code in the Z80 address space, in case of a SCREEN$ header = 16384
}

/*
type tapBlockCompleter interface {
	CompleteBlock() (n int, err error)
}

type writeCompleter interface {
	io.Writer
	tapBlockCompleter
}
*/
type TAPfileBlockWriter struct {
	buf bytes.Buffer
	wtr io.Writer
}

func NewTAPfileBlockWriter(w io.Writer) *TAPfileBlockWriter {

	b := new(TAPfileBlockWriter)

	b.wtr = w

	return b
}

func (b *TAPfileBlockWriter) Write(p []byte) (int, error) {

	if (len(p) + b.buf.Len()) > int(tapBlockMaxLength) {
		return 0, fmt.Errorf("Write error, TAP file block is going to become longer than max length of %d", tapBlockMaxLength)
	}

	n, err := b.buf.Write(p)
	if err != nil {
		return 0, err
	}

	return n, nil
}

func (b *TAPfileBlockWriter) CompleteBlock() error {

	defer b.buf.Reset()

	if err := binary.Write(b.wtr, binary.LittleEndian, uint16(b.buf.Len()+1)); err != nil {
		return err
	}
	if err := binary.Write(b.wtr, binary.LittleEndian, b.buf.Bytes()); err != nil {
		return err
	}
	if err := binary.Write(b.wtr, binary.LittleEndian, xorChecksum(b.buf.Bytes())); err != nil {
		return err
	}

	return nil
}

func xorChecksum(data []byte) uint8 {

	var cs byte = 0
	for _, b := range data {
		cs = cs ^ b
	}
	return cs
}

func (b *BINdata) setFilename(f string) error {

	quoted := strconv.QuoteToASCII(f)
	asciif := strings.Trim(quoted, "\"")
	//log.Println("f: %s, quoted: %s, asciif: %s", f, quoted, asciif)
	if f != asciif {
		return fmt.Errorf("Illegal characters in tap file name: %s", asciif)
	}

	copy(b.filename[:], "          ")
	copy(b.filename[:], asciif)

	return nil
}

//TODO: size check of input file
//TODO: size bigger than 65534
func (b *BINdata) setBinData(bindata io.Reader) error {

	var err error
	b.datablock, err = ioutil.ReadAll(bindata)
	if err != nil {
		b.datablock = nil
		return err
	}

	log.Println("Number of bytes read: ", len(b.datablock))

	return nil
}

func (b *BINdata) setStartAddress(a uint16) error {

	if (int(a) + len(b.datablock)) > math.MaxUint16 {
		return fmt.Errorf("Start address too high, code will roll over 64K-boundary. Address: %d, Length: %d", a, len(b.datablock))
	}

	b.startaddress = a

	return nil
}

func NewBINdata(name string, bindata io.Reader, startaddress uint16) (*BINdata, error) {

	t := new(BINdata)

	if err := t.setFilename(name); err != nil {
		return nil, err
	}
	if err := t.setBinData(bindata); err != nil {
		return nil, err
	}
	if err := t.setStartAddress(startaddress); err != nil {
		return nil, err
	}

	return t, nil
}

func (b *BINdata) Read(r io.Reader) error {

	return nil
}

func (b *BINdata) writeHeader(w *TAPfileBlockWriter) error {

	endianness := binary.LittleEndian

	if err := binary.Write(w, endianness, tapHeaderBlock); err != nil {
		return err
	}
	if err := binary.Write(w, endianness, tapBytesHeader); err != nil {
		return err
	}
	if err := binary.Write(w, endianness, b.filename); err != nil {
		return err
	}
	if err := binary.Write(w, endianness, uint16(len(b.datablock))); err != nil {
		return err
	}
	if err := binary.Write(w, endianness, uint16(b.startaddress)); err != nil {
		return err
	}
	if err := binary.Write(w, endianness, tapUnused); err != nil {
		return err
	}

	return w.CompleteBlock()
}

func (b *BINdata) writeData(w *TAPfileBlockWriter) error {

	endianness := binary.LittleEndian

	if err := binary.Write(w, endianness, tapDataBlock); err != nil {
		return err
	}
	if err := binary.Write(w, endianness, b.datablock); err != nil {
		return err
	}

	return w.CompleteBlock()
}

func (b *BINdata) Write(w *TAPfileBlockWriter) error {

	if err := b.writeHeader(w); err != nil {
		return err
	}

	if err := b.writeData(w); err != nil {
		return err
	}

	return nil
}
