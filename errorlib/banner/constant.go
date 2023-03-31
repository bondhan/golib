package banner

const (
	// BNR00001 stands for "bad request" error
	BNR00001 int = iota + 1
	// BNR00002 stands for "document not found" error
	BNR00002
	// BNR00003 stands for "unexpected error" error
	BNR00003
	// BNR00004 stands for errors originating outside of the banner service context
	BNR00004
)
