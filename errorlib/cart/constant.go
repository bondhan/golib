package cart

const (
	// CRT00001 stands for "not enough quantity" error
	CRT00001 int = iota + 1
	// CRT00002 stands for "quantity ordered more than maximum" error
	CRT00002
	// CRT00003 stands for "quantity ordered more than quota" error
	CRT00003
	// CRT00004 stands for "cart is locked" error
	CRT00004
	// CRT00005 stands for "quantity ordered less than minimum" error
	CRT00005
	// CRT00006 stands for "payment method not found" error
	CRT00006
	// CRT00007 stands for "cart not found" error
	CRT00007
	// CRT00008 stands for "item with zero price can not be ordered" error
	CRT00008
	// CRT00009 stands for "unit per case could not be zero" error
	CRT00009
	// CRT00010 stands for "can not have different uom from item that already in cart" error
	CRT00010
	// CRT00011 stands for "item does not have kardus uom" error
	CRT00011
	// CRT00012 stands for "invalid argument" error
	CRT00012
	// CRT00013 stands for "internal server" error
	CRT00013
	// CRT00014 stands for "bundle is not in cart" error
	CRT00014
	// CRT00015 stands for "reach max bundling product" error
	CRT00015
	// CRT00016 stands for "product not found" error
	CRT00016
	// CRT00017 stands for "inventory not found" error
	CRT00017
)
