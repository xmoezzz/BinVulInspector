package bha

import "github.com/h2non/filetype"

var (
	FileTypeSSFS = filetype.NewType("ssfs", "application/ssfs")
	FileTypeBSD  = filetype.NewType("bsd", "application/bsd")
)

// ssfsMatcher
// 0x73 0x73 0x66 0x73
func ssfsMatcher(buf []byte) bool {
	return len(buf) > 3 &&
		buf[0] == 0x73 && buf[1] == 0x73 && buf[2] == 0x66 && buf[3] == 0x73
}

// bsdMatcher
// 0x50 0x4b 0x03 0x04
// 和zip一样
func bsdMatcher(buf []byte) bool {
	return len(buf) > 3 &&
		buf[0] == 0x50 && buf[1] == 0x4B &&
		(buf[2] == 0x3 || buf[2] == 0x5 || buf[2] == 0x7) &&
		(buf[3] == 0x4 || buf[3] == 0x6 || buf[3] == 0x8)
}

func init() {
	filetype.AddMatcher(FileTypeSSFS, ssfsMatcher)
	filetype.AddMatcher(FileTypeBSD, bsdMatcher)
}
