package segment

const (
	// SGM00001 stands for "Invalid argument" error
	SGM00001 int = iota + 1
	// SGM00002 stands for "data not found" error
	SGM00002
	// SGM00003 stands for "unauthenticated" error
	SGM00003
	// SGM00004 stands for "data already exist" error
	SGM00004
	// SGM00005 stands for "internal server" error
	SGM00005
)
