// Copyright © 2019 Marton Magyar

// SPDX-License-Identifier: MIT
// see https://spdx.org/licenses/

package tapfile

import (
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"strconv"
	"strings"
)

/*
Binary TAP file functions.

This file provides functions to store, read and write binary(bytes) data from and into a TAP file.
*/

// BINData holds essential information and data for binary data in a TAP file
type BINdata struct {
	filename     [10]byte // loading name of the program. filled with spaces (CHR$(32))
	datablock    []byte   // the essential data (may be empty)
	startaddress uint16   // start address of the code in the Z80 address space, in case of a SCREEN$ header = 16384
}

// setFilename validates and sets a new file name
func (b *BINdata) setFilename(f string) error {

	quoted := strconv.QuoteToASCII(f)
	asciif := strings.Trim(quoted, "\"")
	if f != asciif {
		return fmt.Errorf("Illegal characters in tap file name: %s", asciif)
	}

	copy(b.filename[:], "          ")
	copy(b.filename[:], asciif)

	return nil
}

// setBinData sets new binary data
func (b *BINdata) setBinData(bindata io.Reader) error {
	//TODO: size check of input file
	//TODO: size bigger than 65534
	var err error
	b.datablock, err = ioutil.ReadAll(bindata)
	if err != nil {
		b.datablock = nil
		return err
	}

	//TODO: verbose mode
	//log.Println("Number of bytes read: ", len(b.datablock))

	return nil
}

// setStartAddress validates and sets a new starting address for the binary data
func (b *BINdata) setStartAddress(a uint16) error {

	if (int(a) + len(b.datablock)) > math.MaxUint16 {
		return fmt.Errorf("Start address too high, code will roll over 64K-boundary. Address: %d, Length: %d", a, len(b.datablock))
	}

	b.startaddress = a

	return nil
}

// NewBINdata initializes and returns a BINdata structure
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

// Read reads data into a BINdata structure from an io.Reader providing a raw TAP file stream
func (b *BINdata) Read(r io.Reader) error {

	//TODO: fill with function!
	return nil
}

// writeHeader writes raw bytes header data into a specialized TAPfileBlockWriter
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

// writeData writes raw binary block data into a specialized TAPfileBlockWriter
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

// Write writes a binary TAP file
func (b *BINdata) Write(w *TAPfileBlockWriter) error {

	if err := b.writeHeader(w); err != nil {
		return err
	}

	if err := b.writeData(w); err != nil {
		return err
	}

	return nil
}
