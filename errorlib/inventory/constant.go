package inventory

const (
	// INV00001 stands for "invalid argument" error
	INV00001 int = iota + 1
	// INV00002 stands for "data not found" error
	INV00002
	// INV00003 stands for "unauthenticated" error
	INV00003
	// INV00004 stands for "already exist" error
	INV00004
	// INV00005 stands for "internal server" error
	INV00005
	// INV00006 stands for "data not exist" error
	INV00006
	// INV00007 stands for "insufficient qty" error
	INV00007
	// INV00008 stands for "excessive qty" error
	INV00008
	// INV00009 stands for "product not found" error
	INV00009
	// INV00010 stands for "invalid min max qty" error
	INV00010
	// INV00011 stands for "area mismatch" error
	INV00011
	// INV00012 stands for "warehouse not found" error
	INV00012
	// INV00013 stands for "user retailer not found" error
	INV00013
)
