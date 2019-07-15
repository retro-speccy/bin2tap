// Copyright © 2019 Marton Magyar

// SPDX-License-Identifier: MIT
// see https://spdx.org/licenses/

package tapfile

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

/*
TAP file block writer functions.

Writer functions for wrapping blocks into the TAP file format, where a block is preceded by a length block
and followed by a checksum.
*/

// A TAPfileBlockWriter implements a TAP file block wrapper
type TAPfileBlockWriter struct {
	buf bytes.Buffer
	wtr io.Writer
}

// NewTAPfileBlockWriter initializes and returns a TAPfileBlockWriter structure
func NewTAPfileBlockWriter(w io.Writer) *TAPfileBlockWriter {

	b := new(TAPfileBlockWriter)

	b.wtr = w

	return b
}

// Write buffers writes for later wrapping when a block is being completed
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

// CompleteBlock wraps a TAP file block with a preceding length and trailing checksum
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

// xorChecksum calculates a TAP block checksum by simply xor-ing all bytes
func xorChecksum(data []byte) uint8 {

	var cs byte = 0
	for _, b := range data {
		cs = cs ^ b
	}
	return cs
}
