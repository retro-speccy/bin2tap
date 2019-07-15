# bin2tap
Machine code *.bin*-file to ZX Spectrum *.tap* file converter written in Go.

This tool is handy in the workflow of creating a *.tap* file suited for loading in [ZX Spectrum emulators](http://www.worldofspectrum.org/emulators.html) or into a real [ZX Spectrum](https://en.wikipedia.org/wiki/ZX_Spectrum). A *.bin* file contains the pure machine code as output by your favourite Z80 assembler. A *.tap* file contains in addition information like an original ZX Spectrum header as saved in tape files. In the main data chunk a *.tap* file can contain a BASIC program, numerical or alphanumerical arrays for use in BASIC programs, or finally, a machine code program. In case of the latter, the header in the *.tap* file will contain the start address and the length of the machine code.  

# Usage
```shell
	bin2tap infile.bin [outfile.tap] -a=start_address

	bin2tap --help
```

If the outfile argument is omitted, the resulting file will be named *infile.tap*. The starting address is a mandatory parameter. It is a decimal number giving the base address of the machine code in the Z80 address space. The resulting *.tap* file can be loaded with a simple `LOAD ""CODE` command without any further parameters.
# Credits
http://www.zx-modules.de/fileformats/tapformat.html
https://faqwiki.zxnet.co.uk/wiki/TAP_format
# Features
- [x] Basic *.bin* file reading
- [x] Write correct *.tap* headers, including checksums and block legths
# Open issues and features
- [ ] Accept hex starting address on command line
- [ ] Size check of input file
- [ ] Allow input files bigger than 65534 bytes
- [ ] Add support for Spectrum 128K .tap files
- [ ] Allow *.scr* screen image files as input
- [ ] Allow combination of one *.scr* and of one or more *.bin* in a single *.tap* file
- [ ] Add verbose mode
- [ ] Add silent mode
- [ ] Add log to logfile capability
- [ ] More secure type conversions in the code
- [ ] Add support for other platforms (currently only Unix)
- [ ] Make source code more Go-like...
# Contributing
You are very welcome to contribute to this project, whether it be issue reports or new features.
Please contact me before issueing pull requests.

While there surely are other ways to convert *.tap* files (I have been using a simple assembler file containing just an `INCBIN` directive and the [pasmo assembler](http://pasmo.speccy.org/)), I have finally decided to write a converter on my own. Partly to learn how to use [Go](https://golang.org/) and [cobra](https://github.com/spf13/cobra), and to understand the inner workings of a *.tap* file. But also because I am a fan of "[the Unix philosophy: Write programs that do one thing and do it well](https://en.wikipedia.org/wiki/Unix_philosophy)".
