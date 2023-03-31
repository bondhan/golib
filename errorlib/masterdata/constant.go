package masterdata

const (
	// MSD00001 stands for "invalid argument" error
	MSD00001 int = iota + 1
	// MSD00002 stands for "data not found" error
	MSD00002
	// MSD00003 stands for "unauthenticated" error
	MSD00003
	// MSD00004 stands for "already exist" error
	MSD00004
	// MSD00005 stands for "internal server" error
	MSD00005
	// MSD00006 stands for "data not exist" error
	MSD00006
)
