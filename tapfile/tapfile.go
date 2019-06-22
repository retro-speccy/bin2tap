/*
Copyright © 2019 Marton Magyar

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
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

const BASICHeader = 0
const NumArrayHeader = 1
const StringArrayHeader = 2
const BytesHeader = 3

const HeaderBlock = 0
const DataBlock = 0xff

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

func calcChecksum(databytes []byte) uint8 {

    var cs byte = 0
    for _, b := range databytes {
        cs = cs ^ b
    }
    return cs
}

func (t *TAP_BIN_File) SetFilename(f string) error {

    asciif := strconv.QuoteToASCII(f)
    if strings.ContainsAny(asciif, "\\") {
        return fmt.Errorf("Illegal characters in tap file name: %s", asciif)
    }

    copy(t.header.filename[:], "          ")
    copy(t.header.filename[:], f)

    return nil
}

//TODO: size check of input file
//TODO: size bigger than 65534
func (t *TAP_BIN_File) ReadBinData(bindata io.Reader) error {

    var err error
    t.bindata.datablock, err = ioutil.ReadAll(bindata)
    log.Println("Number of bytes read: ", len(t.bindata.datablock))

    if err != nil {
        t.bindata.datablock = nil
        return nil
    }

    t.header.datalength = uint16(len(t.bindata.datablock))
    t.bindatalength.length = t.header.datalength + 2 //TODO: make this more elegant and flexible...

    return nil
}

func (t *TAP_BIN_File) SetStartAddress(a uint16) error {

    if (a + t.header.datalength) > math.MaxUint16 {
        return fmt.Errorf("Start address too high, code will roll over 64K-boundary. Address: %u, Length: %u", a, t.header.datalength)
    }

    t.header.startaddress = a

    return nil
}

func (t *TAP_BIN_File) CalcChecksums() error {

    buf := &bytes.Buffer{}
    err := binary.Write(buf, binary.BigEndian, t.header)
    if err != nil {
        return err
    }
    t.header.checksum = calcChecksum(buf.Bytes())

    t.bindata.checksum = t.bindata.flag ^ calcChecksum(t.bindata.datablock)

    return nil
}

func (t *TAP_BIN_File) Init() error {

    t.headerlength.length = 19 //TODO: uint16(len(t.header))

    t.header.flag = HeaderBlock
    t.header.datatype = BytesHeader
    t.header.unused = (math.MaxUint16 + 1) / 2
    t.header.checksum = 0

    t.bindatalength.length = math.MaxUint16

    t.bindata.flag = DataBlock
    t.bindata.datablock = nil
    t.bindata.checksum = 0

    return nil
}

func (t *TAP_BIN_File) Write(w io.Writer) error {

    log.Println("t.Write")

    err := binary.Write(w, binary.LittleEndian, t.headerlength)
    if err != nil {
        return err
    }

    err = binary.Write(w, binary.LittleEndian, t.header)
    if err != nil {
        return err
    }

    err = binary.Write(w, binary.LittleEndian, t.bindatalength)
    if err != nil {
        return err
    }

    //TODO: make writes more elegant
    err = binary.Write(w, binary.LittleEndian, t.bindata.flag)
    if err != nil {
        return err
    }
    err = binary.Write(w, binary.LittleEndian, t.bindata.datablock)
    if err != nil {
        return err
    }
    err = binary.Write(w, binary.LittleEndian, t.bindata.checksum)
    if err != nil {
        return err
    }

    return nil
}
