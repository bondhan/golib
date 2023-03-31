package product

const (
	// PRD00001 stands for "Invalid argument" error
	PRD00001 int = iota + 1
	// PRD00002 stands for "data not found" error
	PRD00002
	// PRD00003 stands for "unauthenticated" error
	PRD00003
	// PRD00004 stands for "already exist" error
	PRD00004
	// PRD00005 stands for "internal server" error
	PRD00005
	// PRD00006 stands for "parent sku not exists" error
	PRD00006
	// PRD00007 stands for "invalid parent sku code" error
	PRD00007
	// PRD00008 stands for "sku code already exists" error
	PRD00008
	// PRD00009 stands for "barcode already exist" error
	PRD00009
	// PRD00010 stands for "principle id mismatch" error
	PRD00010
	// PRD00011 stands for "empty retailer area" error
	PRD00011
	// PRD00012 stands for "invalid retailer area" error
	PRD00012
	// PRD00013 stands for "invalid child status" error
	PRD00013
	// PRD00014 stands for "invalid parent status" error
	PRD00014
)
