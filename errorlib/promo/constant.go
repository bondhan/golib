package promo

const (
	// PRM00001 stands for "Invalid Request" error
	PRM00001 int = iota + 1
	// PRM00002 stands for "Voucher is Exists" error
	PRM00002
	// PRM00003 stands for "Voucher not found" error
	PRM00003
	// PRM00004 stands for "Voucher does not meet requirement" error
	PRM00004
	// PRM00005 stands for "Voucher is not active" error
	PRM00005
	// PRM00006 stands for "Invalid Validate Request" error
	PRM00006
	// PRM00007 stands for "Voucher is not active" error
	PRM00007
	// PRM00008 stands for "Voucher redeem has exceed quota" error
	PRM00008
	// PRM00009 stands for "Retailer has exceeded redeem limit" error
	PRM00009
	// PRM00010 stands for "Voucher signature mismatch with order" error
	PRM00010
	// PRM00011 stands for "Duplicate redeem request" error
	PRM00011
	// PRM00012 stands for "SKU Discount Invalid request" error
	PRM00012
	// PRM00013 stands for "Sku Discount doesnt exist" error
	PRM00013
	// PRM00014 stands for "Sku Discount doesnt meet requirement" error
	PRM00014
	// PRM00015 stands for "Order did not reach minimum order amount" error
	PRM00015
	// PRM00016 stands for "Order did not reach minimum item count" error
	PRM00016
	// PRM00017 stands for "Did not match retailer group rules" error
	PRM00017
	// PRM00018 stands for "Did not meet location province criteria" error
	PRM00018
	// PRM00019 stands for "Did not meet location city criteria" error
	PRM00019
	// PRM00020 stands for "Did not meet product sku requirements" error
	PRM00020
	// PRM00021 stands for "Did not meet product categories requirements" error
	PRM00021
	// PRM00022 stands for "Did not meet product brands requirements" error
	PRM00022
	// PRM00023 stands for "Did not meet product companies requirements" error
	PRM00023
	// PRM00024 stands for "Did not meet minimum quantity in item" error
	PRM00024
	// PRM00025 stands for "Did not meet minimum subtotal in item" error
	PRM00025
	// PRM00026 stands for "Retailer has reached redeem limit" error
	PRM00026
	// PRM00027 stands for "Order does not meet voucher validation requirement" error
	PRM00027
	// PRM00028 stands for "Did not match retailer segment rules" error
	PRM00028
	// PRM00029 stands for "Did not match retailer warehouse rules" error
	PRM00029
	// PRM00030 stands for "Did not match retailer warehouse cluster rules" error
	PRM00030
	// PRM00031 stands for "Did not match retailer warehouse region rules" error
	PRM00031
	// PRM00032 stands for "Did not match retailer warehouse sub region rules" error
	PRM00032
	// PRM00033 stands for "Did not match retailer location" error
	PRM00033
	// PRM00034 stands for "Internal server" error
	PRM00034
)
