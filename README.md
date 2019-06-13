# bin2tap
Machine code `.bin`-file to ZX Spectrum `.tap` file converter written in Go.

This tool is handy in the workflow of creating a `.tap` file suited for loading in [ZX Spectrum emulators](http://www.worldofspectrum.org/emulators.html) or into real ZX Spectrum hardware. A `.bin` file contains the pure machine code as output by your favourite Z80 assembler. A `.tap` file contains in addition information like an original ZX Spectrum header as saved in tape files. In the main data chunk a `.tap` file can contain a BASIC program, numerical or alphanumerical arrays for use in BASIC programs, or finally, a machine code program. In case of the latter, the header in the `.tap` file will contain the start address and the length of the machine code.  

While there surely are other ways to convert `.tap` files (I have been using a simple assembler file containing just an `INCBIN` directive and the pasmo assembler), I have finally decided to write a converter on my own. Partly to learn how to use [Go](https://golang.org/), [cobra](https://github.com/spf13/cobra), and to understand the inner workings of a `.tap` file. But also because I am of "[the Unix philosophy: Write programs that do one thing and do it well](https://en.wikipedia.org/wiki/Unix_philosophy)".
# Usage
	`bin2tap infile.bin [outfile.tap] -a|--address=start_address`

If the outfile argument is omitted, the resulting file will be named infile.tap. The starting address paramter is mandatory. The resulting `.tap` file can be loaded with a simple `LOAD ""CODE` command without any further parameters.
# Credits
https://web.archive.org/web/20110711141601/http://www.zxmodules.de:80/fileformats/tapformat.html
https://faqwiki.zxnet.co.uk/wiki/TAP_format
# Open issues and features
[ ] Accept hex starting address on command line
[ ] Size check of input file
[ ] Allow input files bigger than 65534 bytes
[ ] Add support for Spectrum 128K .tap files
[ ] Add verbose mode
[ ] Add silent mode
[ ] More secure type conversions in the code
[ ] Beautfy source code...
# Contributing
You are very welcome to contribute to this project, whether it be issue reports or new features. Please contact me before issueing pull requests.