package util

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wI2L/jsondiff"
)

func TestEvalRoot(t *testing.T) {
	data := false

	ev, err := NewEvaluator("$. == false")
	assert.Nil(t, err)

	o, err := ev.Eval(data)
	assert.Nil(t, err)
	assert.True(t, o.(bool))
}

func TestEval(t *testing.T) {
	data := map[string]interface{}{
		"metadata": map[string]interface{}{
			"timestamp":          time.Now(),
			"record-type":        "control",
			"operation":          "update",
			"partition-key-type": "task-id",
			"schema-name":        "",
			"sequence":           87,
			"table-name":         "awsdms_apply_exceptions",
			"active":             "true",
			"items":              []string{"1", "2", "3"},
			"$num":               "-1.23211",
		},
	}

	ev, err := NewEvaluator("$metadata.record-type == 'control' && ($metadata.operation == 'create' || $metadata.operation == 'update') && $metadata.schema-name == ''")
	assert.Nil(t, err)

	o, err := ev.Eval(data)
	assert.Nil(t, err)
	assert.True(t, o.(bool))

	ev, err = NewEvaluator("match('metadata.record-type', $., 'control')")
	assert.Nil(t, err)

	o, err = ev.Eval(data)
	assert.Nil(t, err)
	assert.True(t, o.(bool))

	ev, err = NewEvaluator("$record-type != '' && $operation != ''? ('record','op') : ($record-type == '' && $operation != ''? ('op') : ($record-type != '' && $operation == ''? ('record') : null() )) ")
	assert.Nil(t, err)

	o, err = ev.Eval(map[string]interface{}{
		"record-type": "control",
		"operation":   "update",
	})
	assert.Nil(t, err)
	assert.Equal(t, []interface{}{"record", "op"}, o)

	o, err = ev.Eval(map[string]interface{}{
		"record-type": "control",
	})
	assert.Nil(t, err)
	assert.Equal(t, "record", o)

	o, err = ev.Eval(map[string]interface{}{
		"operation": "update",
	})
	assert.Nil(t, err)
	assert.Equal(t, "op", o)

	o, err = ev.Eval(map[string]interface{}{})
	assert.Nil(t, err)
	assert.Equal(t, nil, o)

	ev, err = NewEvaluator("eval:int($metadata.sequence)")
	assert.Nil(t, err)

	o, err = ev.Eval(data)
	assert.Nil(t, err)
	assert.Equal(t, 87, o)

	ev, err = NewEvaluator("eval:int($metadata.sequence)")
	assert.Nil(t, err)

	o, err = ev.Eval(data)
	assert.Nil(t, err)
	assert.Equal(t, 87, o)

	ev, err = NewEvaluator("eval:date()")
	assert.Nil(t, err)

	o, err = ev.Eval(data)
	assert.Nil(t, err)
	assert.Equal(t, time.Now().Format("2006/01/02"), o)

	ev, err = NewEvaluator("eval:date($metadata.timestamp)")
	assert.Nil(t, err)

	o, err = ev.Eval(data)
	assert.Nil(t, err)
	assert.Equal(t, time.Now().Format("2006/01/02"), o)

	ev, err = NewEvaluator("eval:date('2021-Mar-29', '2006-Jan-02')")
	assert.Nil(t, err)

	o, err = ev.Eval(data)
	assert.Nil(t, err)
	assert.Equal(t, "2021/03/29", o)

	to := time.Now()
	tb, _ := json.Marshal(to)
	ts := strings.Trim(string(tb), `"`)

	er := "eval:ftime('" + ts + "', '2006-01-02 15:04:05')"
	fmt.Println(er)

	ev, err = NewEvaluator("eval:ftime('" + ts + "', 'T2006-01-02 15:04:05')")
	assert.Nil(t, err)

	o, err = ev.Eval(data)
	assert.Nil(t, err)
	assert.Equal(t, time.Now().Format("2006-01-02 15:04:05"), o)

	ev, err = NewEvaluator("eval:len($metadata.items)")
	assert.Nil(t, err)

	o, err = ev.Eval(data)
	assert.Nil(t, err)
	assert.Equal(t, 3, o)

	ev, err = NewEvaluator("eval:float($metadata.$num)")
	assert.Nil(t, err)

	o, err = ev.Eval(data)
	assert.Nil(t, err)
	assert.Equal(t, -1.23211, o)

	ev, err = NewEvaluator("eval:env('test')")
	assert.Nil(t, err)

	os.Setenv("TEST", "testing")

	o, err = ev.Eval(nil)
	assert.Nil(t, err)
	assert.Equal(t, "testing", o)

	tnow := time.Now().Unix()
	data["ts"] = tnow
	ev, err = NewEvaluator("eval:time($ts)")
	assert.Nil(t, err)

	o, err = ev.Eval(data)
	assert.Nil(t, err)
	assert.Equal(t, time.Unix(tnow, 0), o)

	ev, err = NewEvaluator("eval:'data-' + $metadata.operation")
	assert.Nil(t, err)

	o, err = ev.Eval(data)
	assert.Nil(t, err)
	assert.Equal(t, "data-update", o)
}

func TestEvalStringJSON(t *testing.T) {
	data := map[string]interface{}{
		"metadata": map[string]interface{}{
			"timestamp":          time.Now(),
			"record-type":        "control",
			"operation":          "update",
			"partition-key-type": "task-id",
			"schema-name":        "",
			"sequence":           87,
			"table-name":         "awsdms_apply_exceptions",
			"active":             "true",
			"items":              []string{"1", "2", "3"},
			"num":                "-1.23211",
		},
	}

	b, err := json.Marshal(data)
	require.Nil(t, err)
	require.NotNil(t, b)

	ev, err := NewEvaluator("$metadata.record-type == 'control' && ($metadata.operation == 'create' || $metadata.operation == 'update') && $metadata.schema-name == ''")
	assert.Nil(t, err)

	o, err := ev.EvalStringJSON(string(b))
	assert.Nil(t, err)
	assert.True(t, o.(bool))

	ev, err = NewEvaluator("match('metadata.record-type', $., 'control')")
	assert.Nil(t, err)

	o, err = ev.EvalStringJSON(string(b))
	assert.Nil(t, err)
	fmt.Println(o)
	assert.True(t, o.(bool))

	ev, err = NewEvaluator("eval:int($metadata.sequence)")
	assert.Nil(t, err)

	o, err = ev.EvalStringJSON(string(b))
	assert.Nil(t, err)
	assert.Equal(t, 87, o)

	ev, err = NewEvaluator("eval:int($metadata.sequence)")
	assert.Nil(t, err)

	o, err = ev.EvalStringJSON(string(b))
	assert.Nil(t, err)
	assert.Equal(t, 87, o)

	ev, err = NewEvaluator("eval:date()")
	assert.Nil(t, err)

	o, err = ev.EvalStringJSON(string(b))
	assert.Nil(t, err)
	assert.Equal(t, time.Now().Format("2006/01/02"), o)

	ev, err = NewEvaluator("eval:date($metadata.timestamp)")
	assert.Nil(t, err)

	o, err = ev.EvalStringJSON(string(b))
	assert.Nil(t, err)
	assert.Equal(t, time.Now().Format("2006/01/02"), o)

	ev, err = NewEvaluator("eval:date('2021-Mar-29', '2006-Jan-02')")
	assert.Nil(t, err)

	o, err = ev.EvalStringJSON(string(b))
	assert.Nil(t, err)
	assert.Equal(t, "2021/03/29", o)

	to := time.Now()
	tb, _ := json.Marshal(to)
	ts := strings.Trim(string(tb), `"`)

	er := "eval:ftime('" + ts + "', '2006-01-02 15:04:05')"
	fmt.Println(er)

	ev, err = NewEvaluator("eval:ftime('" + ts + "', 'T2006-01-02 15:04:05')")
	assert.Nil(t, err)

	o, err = ev.EvalStringJSON(string(b))
	assert.Nil(t, err)
	assert.Equal(t, time.Now().Format("2006-01-02 15:04:05"), o)

	ev, err = NewEvaluator("eval:len($metadata.items)")
	assert.Nil(t, err)

	o, err = ev.EvalStringJSON(string(b))
	assert.Nil(t, err)
	assert.Equal(t, 3, o)

	ev, err = NewEvaluator("eval:float($metadata.num)")
	assert.Nil(t, err)

	o, err = ev.EvalStringJSON(string(b))
	assert.Nil(t, err)
	assert.Equal(t, -1.23211, o)
}

func TestEvalDiff(t *testing.T) {
	data := map[string]interface{}{
		"source": map[string]interface{}{
			"timestamp":          time.Now(),
			"record-type":        "control",
			"operation":          "update",
			"partition-key-type": "task-id",
			"schema-name":        "",
			"sequence":           87,
			"table-name":         "awsdms_apply_exceptions",
			"active":             "true",
			"items":              []string{"1", "2", "3"},
			"num":                "-1.23211",
		},
		"target": map[string]interface{}{
			"timestamp":          time.Now(),
			"record-type":        "control",
			"operation":          "update",
			"partition-key-type": "task-id",
			"schema-name":        "",
			"sequence":           88,
			"table-name":         "awsdms_apply_exceptions",
			"active":             "true",
			"items":              []string{"1", "2", "3"},
			"num":                "-1.23211",
		},
	}

	ev, err := NewEvaluator("eval:diff($source, $target, 'all')")
	assert.Nil(t, err)

	o, err := ev.Eval(data)
	assert.Nil(t, err)
	assert.NotNil(t, o)

	omap := o.(map[string]interface{})
	assert.Equal(t, 2, len(omap["diff"].(jsondiff.Patch)))

	ev, err = NewEvaluator("eval:diff($source, $target)")
	assert.Nil(t, err)

	diff, err := ev.Eval(data)
	assert.Nil(t, err)
	assert.NotNil(t, diff)

	assert.Equal(t, 2, len(diff.(jsondiff.Patch)))
}

func TestObjectEval(t *testing.T) {
	dataStr := `
	{
		"data": {
		  "id": "3fsoUG4Po8RKY2iU5mAw8qhyN3bhnYeppd1d8RkfDcFf",
		  "unitName": "Hello771",
		  "familiarName": "Hello772",
		  "code": "Hello773",
		  "definition": "Hello",
		  "status": true,
		  "type": "Hello774",
		  "createdAt": 1664791968,
		  "updatedAt": 1664793119,
		  "createdBy": "otn uchiha",
		  "updatedBy": "otn uchiha"
		},
		"metadata": {
		  "audit": {
			"action": "record-update",
			"dataOld": {
			  "id": "3fsoUG4Po8RKY2iU5mAw8qhyN3bhnYeppd1d8RkfDcFf",
			  "unitName": "Hello881",
			  "familiarName": "Hello882",
			  "code": "Hello883",
			  "definition": "Hello",
			  "status": true,
			  "type": "Hello",
			  "createdAt": 1664791968,
			  "updatedAt": 1664792892,
			  "createdBy": "otn uchiha",
			  "updatedBy": "otn uchiha"
			},
			"timestamp": 1664793119,
			"user": "otn uchiha"
		  },
		  "event": "product.salesunit-updated",
		  "hash": "X+i+NvrCggmX8C7JoWe8j7crmo8irmKZBqkZgYlahyY=",
		  "timestamp": "2022-10-03T10:31:59.153346121Z",
		  "version": 1
		}
	  }
	`

	var data map[string]interface{}
	json.Unmarshal([]byte(dataStr), &data)

	ev, err := NewEvaluator("eval:diff($metadata.audit.dataOld, $data, 'all')")
	assert.Nil(t, err)

	o, err := ev.Eval(data)
	assert.Nil(t, err)
	assert.NotNil(t, o)
}

func TestQuery(t *testing.T) {
	str := `{
		"_id": "635a29523d60355de97ddee8",
		"items": [
		  {
		  "amount": {
			"discount": "54000",
			"skuDiscount": "0",
			"bundleDiscount": "54000",
			"subtotal": "702000",
			"tax": "64216",
			"total": "648000",
			"voucherDiscount": "0"
		  },
		  "area": {
			"region": "region1",
			"regionSub": "subregion1",
			"cluster": "tangerang selatan",
			"warehouse": "bitung"
		  },
		  "discount": {
			"discountCalcType": 0,
			"discountValue": 0,
			"maxDiscountQty": 0,
			"title": "",
			"unitAmount": 0,
			"value": "0",
			"type": "",
			"description": "",
			"discountQty": 0
		  },
		  "pricing": {
			"type": "9AAPGXDh28KEtngvtqEVrKo8LwA9yeuwvWnWYDBhvywW",
			"amount": "19500",
			"qty": 1
		  },
		  "product": {
			"skuNo": "1A04AAU12365",
			"unitPerCase": 36,
			"brand": {
			"id": "ufNwvufuQzKXAqqvC65Yr43wSCHjB7RM8WywNkGuTL5",
			"name": "Sun Kara"
			},
			"category": {
			"id": "2LnW5JZsZj8o8Ssu6uXLFWArfd93Cq2TvWLaQZ5pzjSh",
			"name": "Bahan Dapur",
			"taxPercentage": "0.11"
			},
			"company": {
			"id": "21ndxgQnnAn2q39XUr8YHt3tZ7QK9mJzhLFuiw9qSXZV",
			"name": "Sun Kara"
			},
			"image": "https://storage.googleapis.com/p-gotoko-legacy-images/product/1A04AAU12365.jpg",
			"name": "Sun Kara Coconut Milk 1X65 ML (Dijual per 1 Pcs)"
		  },
		  "qty": {
			"uom": "unit",
			"itemBreakdownOrdered": 0,
			"min": 0,
			"max": 100000000,
			"inStock": 101099997,
			"ordered": 36,
			"totalUnitOrdered": 36
		  },
		  "tax": "",
		  "id": "5KAfiyugfzuXZgpo7NK94zFzj1RsXmLmMew5NDEJqAnY"
		  },
		  {
		  "tax": "",
		  "id": "3rJSVRST3jXF8HSTUpbEkJqFacgJwUJ7NBw2wHGtZLzx",
		  "amount": {
			"discount": "67500",
			"skuDiscount": "0",
			"bundleDiscount": "67500",
			"subtotal": "382050",
			"tax": "31171",
			"total": "314550",
			"voucherDiscount": "0"
		  },
		  "area": {
			"region": "region1",
			"regionSub": "subregion1",
			"cluster": "tangerang selatan",
			"warehouse": "bitung"
		  },
		  "discount": {
			"type": "",
			"discountCalcType": 0,
			"discountValue": 0,
			"title": "",
			"unitAmount": 0,
			"value": "0",
			"description": "",
			"discountQty": 0,
			"maxDiscountQty": 0
		  },
		  "pricing": {
			"type": "8MsRDWuzPrTASuUgFTz226KWFndsMRi2YEkEjWeNjkmk",
			"amount": "14150",
			"qty": 1
		  },
		  "product": {
			"name": "Ladaku Sachet 4gr (Dijual per 1 Renceng isi 12 Pcs)",
			"skuNo": "1A04AAU12355",
			"unitPerCase": 576,
			"brand": {
			"id": "D9pL3rz4AjLLS2g7FUpp7tX6qQ5dS8XSdjGbcZmgSuva",
			"name": "Ladaku"
			},
			"category": {
			"id": "2LnW5JZsZj8o8Ssu6uXLFWArfd93Cq2TvWLaQZ5pzjSh",
			"name": "Bahan Dapur",
			"taxPercentage": "0.11"
			},
			"company": {
			"id": "GLwfDErYWu25MBfQgRCFxZ6jPpAPCm6AZzgqmSkvrQGv",
			"name": "Ajinomoto"
			},
			"image": "https://storage.googleapis.com/gotoko-legacy-images/product/A04AAU12355.png"
		  },
		  "qty": {
			"min": 1,
			"max": 100000,
			"inStock": 199998,
			"ordered": 27,
			"totalUnitOrdered": 27,
			"uom": "unit",
			"itemBreakdownOrdered": 0
		  }
		  },
		  {
		  "id": "7nwJAwMJNE8JXCS4eov14cw63cH5i75j9myNiq8Zo3Qh",
		  "amount": {
			"skuDiscount": "0",
			"bundleDiscount": "0",
			"subtotal": "25700",
			"tax": "2546",
			"total": "25700",
			"voucherDiscount": "0",
			"discount": "0"
		  },
		  "area": {
			"regionSub": "subregion1",
			"cluster": "tangerang selatan",
			"warehouse": "bitung",
			"region": "region1"
		  },
		  "discount": {
			"type": "",
			"unitAmount": 0,
			"discountValue": 0,
			"maxDiscountQty": 0,
			"title": "",
			"value": "0",
			"description": "",
			"discountCalcType": 0,
			"discountQty": 0
		  },
		  "pricing": {
			"amount": "12850",
			"qty": 1,
			"type": "9AAPGXDh28KEtngvtqEVrKo8LwA9yeuwvWnWYDBhvywW"
		  },
		  "product": {
			"image": "https://storage.googleapis.com/gotoko-legacy-images/product/B04BAD12358.png",
			"name": "LIFEBUOY TS LEMONFRESH ARG DG 70G (Dijual per 3 Pcs)",
			"skuNo": "1B04BAD12358",
			"unitPerCase": 144,
			"brand": {
			"id": "E1J6vMJnsEaFqupB5zP2sHrZs1XYyTPAR2wmfJ2uUro6",
			"name": "Lifebuoy"
			},
			"category": {
			"id": "55BdNZ7eMsv5CRjvcB7jwEnnXtBoZNQDaJvTqHbr7Qvq",
			"name": "Perlengkapan Mandi",
			"taxPercentage": "0.11"
			},
			"company": {
			"id": "78nD6gc9uG3ty9YChbLf8NmA6fwefBLqNLHbY5UdG3UP",
			"name": "Unilever"
			}
		  },
		  "qty": {
			"itemBreakdownOrdered": 0,
			"min": 100000000,
			"max": 0,
			"inStock": 100123455,
			"ordered": 2,
			"totalUnitOrdered": 2,
			"uom": "unit"
		  },
		  "tax": ""
		  },
		  {
		  "id": "DQkgBZLYstFo9vhAwrUPkpJy8NRHhmjgZaMEqB4vT8qx",
		  "amount": {
			"discount": "27000",
			"skuDiscount": "0",
			"bundleDiscount": "27000",
			"subtotal": "1330560",
			"tax": "129181",
			"total": "1303560",
			"voucherDiscount": "0"
		  },
		  "area": {
			"region": "region1",
			"regionSub": "subregion1",
			"cluster": "tangerang selatan",
			"warehouse": "bitung"
		  },
		  "discount": {
			"value": "0",
			"discountQty": 0,
			"maxDiscountQty": 0,
			"title": "",
			"unitAmount": 0,
			"type": "",
			"description": "",
			"discountCalcType": 0,
			"discountValue": 0
		  },
		  "pricing": {
			"qty": 1,
			"type": "9AAPGXDh28KEtngvtqEVrKo8LwA9yeuwvWnWYDBhvywW",
			"amount": "73920"
		  },
		  "product": {
			"brand": {
			"id": "FRPdC6KWCouyD4JxFKt4FEtxo8JtdKzZGT83LwjMCztM",
			"name": "Masako"
			},
			"category": {
			"taxPercentage": "0.11",
			"id": "2LnW5JZsZj8o8Ssu6uXLFWArfd93Cq2TvWLaQZ5pzjSh",
			"name": "Bahan Dapur"
			},
			"company": {
			"id": "GLwfDErYWu25MBfQgRCFxZ6jPpAPCm6AZzgqmSkvrQGv",
			"name": "Ajinomoto"
			},
			"image": "https://storage.googleapis.com/gotoko-legacy-images/product/A04AAU12352.png",
			"name": "Masako Ayam Big Pack 250 GR (Dijual per 1 Pcs)",
			"skuNo": "1A04AAU12352",
			"unitPerCase": 48
		  },
		  "qty": {
			"max": 100000000,
			"inStock": 101099997,
			"ordered": 18,
			"totalUnitOrdered": 18,
			"uom": "unit",
			"itemBreakdownOrdered": 0,
			"min": 0
		  },
		  "tax": ""
		  }
		],
		"payment": {
		  "name": "Tunai",
		  "slug": "tunai",
		  "createdAt": 0,
		  "updatedAt": 0,
		  "code": "TUNAI",
		  "icon": "Tunai"
		},
		"amount": {
			"total": "2291810",
			"discount": "148500",
			"skuDiscount": "0",
			"bundleDiscount": "148500",
			"tax": "227114",
			"shipping": "0",
			"subtotal": "2440310",
			"voucher": "0",
			"shippingTax": "0"
		},
		"number": "GTO-1666853201537",
		"voucher": null,
		"user": "BA21LrwmpvY2Iv9RmFjS",
		"createdAt": 1666853201,
		"updatedAt": 1666853207844,
		"awb": "",
		"eta": "2022-10-28",
		"shippingStatus": "",
		"invoiceUrl": {
		  "retailer": "https://storage.googleapis.com/s-westeros-bucket-01/order-invoices/2022-10-27/RN-1631763111/GTO-1666853201537/invoice-GTO-1666853201537-retailer.pdf",
		  "ops": "https://storage.googleapis.com/s-westeros-bucket-01/order-invoices/2022-10-27/RN-1631763111/GTO-1666853201537/invoice-GTO-1666853201537-ops.pdf"
		},
		"displayStatus": "Pesanan diterima",
		"itemsBreakdown": {
		  "total": 148500,
		  "items": {
			"6cbqrTAunFxPDMMLDVePHvTukhbaEGBDMS71n44P2qZh": {
			  "discountType": 1,
			  "totalValue": 148500,
			  "items": {
				"1A04AAU12352": {
				  "qty": 18,
				  "unitQty": "unit",
				  "discountQty": 18,
				  "discountValue": 1500,
				  "totalAmount": 27000,
				  "name": "Masako Ayam Big Pack 250 GR (Dijual per 1 Pcs)",
				  "skuNo": "1A04AAU12352",
				  "unitPrice": 73920
				},
				"1A04AAU12365": {
				  "unitPrice": 19500,
				  "qty": 36,
				  "unitQty": "unit",
				  "discountQty": 36,
				  "discountValue": 1500,
				  "totalAmount": 54000,
				  "name": "Sun Kara Coconut Milk 1X65 ML (Dijual per 1 Pcs)",
				  "skuNo": "1A04AAU12365"
				},
				"1A04AAU12355": {
				  "unitPrice": 14150,
				  "qty": 27,
				  "unitQty": "unit",
				  "discountQty": 27,
				  "discountValue": 2500,
				  "totalAmount": 67500,
				  "name": "Ladaku Sachet 4gr (Dijual per 1 Renceng isi 12 Pcs)",
				  "skuNo": "1A04AAU12355"
				}
			  },
			  "qty": 9,
			  "maxQty": 10,
			  "description": "bumbu murah meriah",
			  "id": "6cbqrTAunFxPDMMLDVePHvTukhbaEGBDMS71n44P2qZh",
			  "title": "Bundle20.1"
			},
			"6cbqrTAunFxPDMMLDVePHvTukhbaEGBDMS71n44P2qZx": {
				"discountType": 0,
				"totalValue": 148500,
				"items": {
				  "1A04AAU12352": {
					"qty": 18,
					"unitQty": "unit",
					"discountQty": 18,
					"discountValue": 1500,
					"totalAmount": 27000,
					"name": "Masako Ayam Big Pack 250 GR (Dijual per 1 Pcs)",
					"skuNo": "1A04AAU12352",
					"unitPrice": 73920
				  },
				  "1A04AAU12365": {
					"unitPrice": 19500,
					"qty": 36,
					"unitQty": "unit",
					"discountQty": 36,
					"discountValue": 1500,
					"totalAmount": 54000,
					"name": "Sun Kara Coconut Milk 1X65 ML (Dijual per 1 Pcs)",
					"skuNo": "1A04AAU12365"
				  },
				  "1A04AAU12355": {
					"unitPrice": 14150,
					"qty": 27,
					"unitQty": "unit",
					"discountQty": 27,
					"discountValue": 2500,
					"totalAmount": 67500,
					"name": "Ladaku Sachet 4gr (Dijual per 1 Renceng isi 12 Pcs)",
					"skuNo": "1A04AAU12355"
				  }
				},
				"qty": 9,
				"maxQty": 10,
				"description": "bumbu murah meriah",
				"id": "6cbqrTAunFxPDMMLDVePHvTukhbaEGBDMS71n44P2qZh",
				"title": "Bundle20.1"
			}
		  }
		}
	  }`

	ev, err := NewEvaluator("eval:query($., 's:product.name;f:items;w: product.skuNo = 1A04AAU12365 |> first:name')")
	assert.Nil(t, err)

	var data map[string]interface{}
	require.Nil(t, json.Unmarshal([]byte(str), &data))
	o, err := ev.Eval(data)
	assert.Nil(t, err)
	assert.NotNil(t, o)

	assert.Equal(t, "Sun Kara Coconut Milk 1X65 ML (Dijual per 1 Pcs)", o)

	ev, err = NewEvaluator("eval:query($., 'e: itemsBreakdown.items; f:items ;s:- -> w: skuNo = 1A04AAU12352 |> sum:qty')")
	assert.Nil(t, err)

	o, err = ev.Eval(data)
	assert.Nil(t, err)
	assert.NotNil(t, o)

	assert.Equal(t, float64(36), o)

	ev, err = NewEvaluator("eval:query($., 'f:items; w: product.brand.name = Sun Kara; s:qty.ordered |> sum:ordered ')")
	assert.Nil(t, err)

	o, err = ev.Eval(data)
	assert.Nil(t, err)
	assert.NotNil(t, o)

	assert.Equal(t, float64(36), o)

	ev, err = NewEvaluator("$createdAt > 1666853201")
	assert.Nil(t, err)
	o, err = ev.Eval(data)
	assert.Nil(t, err)
	assert.NotNil(t, o)

	assert.False(t, o.(bool))

	ev, err = NewEvaluator("eval:query($., 'e: itemsBreakdown.items; w: discountType = 0 -> e:. ;f:items ; s:- -> w: skuNo = 1A04AAU12352 |> sum:qty')")
	assert.Nil(t, err)

	o, err = ev.Eval(data)
	assert.Nil(t, err)
	assert.NotNil(t, o)

	assert.Equal(t, float64(18), o)
}

func TestAnotherQuery(t *testing.T) {
	str := `{
		"_id": {
			"$oid": "62aade63c7e6feaffbe3a4c6"
		},
		"user": "tPrQPYSsfDPac8R3up5G",
		"items": [
			{
				"pricing": {
					"amount": "23500",
					"qty": {
						"$numberLong": "1"
					},
					"type": "Pack"
				},
				"product": {
					"name": "Dunhill 16 Hitam",
					"skuNo": "1B01BAA12350",
					"tax": "",
					"unitPerCase": {
						"$numberLong": "100"
					},
					"brand": {
						"id": "L2tnzj2TGAwdArJteUJLrjG9sXpQwhzMg4GtR94vKVK",
						"name": "Dunhill"
					},
					"category": {
						"id": "G8UcF5e5sfgYgqRR3tD2zfNpSUteqo4rFakZk3aCGHiP",
						"name": "Rokok"
					},
					"company": {
						"id": "6p9Q17xEAu4cmZca8sJXTGRRHwpiRmKVF3LZ1ZmSn2GE",
						"name": "British American Tobacco"
					},
					"image": "https://storage.googleapis.com/gotoko-legacy-images/product/B01BAA12350.png"
				},
				"qty": {
					"min": {
						"$numberLong": "1"
					},
					"max": {
						"$numberLong": "100000"
					},
					"inStock": {
						"$numberLong": "9999999"
					},
					"ordered": {
						"$numberLong": "1"
					}
				},
				"id": "7zmHYqcDQCsrp3NsKfRZ6tsdF95Di9kGfNsTG9VD7vWu",
				"amount": {
					"discount": "0",
					"subtotal": "23500",
					"tax": "2328.82882882882883",
					"total": "23500"
				},
				"area": {
					"region": "JABAR - JATENG",
					"regionSub": "Bandung",
					"cluster": "Bandung 1",
					"warehouse": "bitung"
				},
				"discount": {
					"value": "0",
					"description": "",
					"discountCalcType": {
						"$numberLong": "0"
					},
					"title": "",
					"unitAmount": {
						"$numberLong": "0"
					},
					"type": "",
					"discountQty": {
						"$numberLong": "0"
					},
					"discountValue": {
						"$numberLong": "0"
					},
					"maxDiscountQty": {
						"$numberLong": "0"
					}
				}
			},
			{
				"amount": {
					"subtotal": "24310",
					"tax": "2409.0990990990991",
					"total": "24310",
					"discount": "0"
				},
				"area": {
					"region": "JABAR - JATENG",
					"regionSub": "Bandung",
					"cluster": "Bandung 1",
					"warehouse": "bitung"
				},
				"discount": {
					"value": "0",
					"type": "",
					"unitAmount": {
						"$numberLong": "0"
					},
					"discountValue": {
						"$numberLong": "0"
					},
					"maxDiscountQty": {
						"$numberLong": "0"
					},
					"title": "",
					"description": "",
					"discountCalcType": {
						"$numberLong": "0"
					},
					"discountQty": {
						"$numberLong": "0"
					}
				},
				"pricing": {
					"amount": "24310",
					"qty": {
						"$numberLong": "12"
					},
					"type": "Strip"
				},
				"product": {
					"tax": "",
					"unitPerCase": {
						"$numberLong": "144"
					},
					"brand": {
						"id": "Aci8jSNNuC4SFyV1YQYCkb2iLhyiU4FWmGAQ7rte4qjL",
						"name": "Rexona"
					},
					"category": {
						"id": "7Jz9j8QuZtMYGdUqJfCuLRswY8jxLToAhK41nMKUcTj7",
						"name": "Kecantikan & Kesehatan"
					},
					"company": {
						"id": "2gvnL2PiqXFvSkqgMuk8oABNTG3eToDi6TrGyxTzh3z4",
						"name": "Unilever"
					},
					"image": "https://storage.googleapis.com/gotoko-legacy-images/product/B02BAB12360.png",
					"name": "REXONA DEO LOTION FREE SPIRIT 144X9G",
					"skuNo": "1B02BAB12360"
				},
				"qty": {
					"min": {
						"$numberLong": "1"
					},
					"max": {
						"$numberLong": "100000"
					},
					"inStock": {
						"$numberLong": "9999999"
					},
					"ordered": {
						"$numberLong": "1"
					}
				},
				"id": "AKUgRAmHPsNwZ8XDucSoxSSyoMzhnhQCGqjmXKNNkjNS"
			},
			{
				"area": {
					"region": "JABAR - JATENG",
					"regionSub": "Bandung",
					"cluster": "Bandung 1",
					"warehouse": "bitung"
				},
				"discount": {
					"type": "",
					"discountQty": {
						"$numberLong": "0"
					},
					"unitAmount": {
						"$numberLong": "0"
					},
					"value": "0",
					"description": "",
					"discountCalcType": {
						"$numberLong": "0"
					},
					"discountValue": {
						"$numberLong": "0"
					},
					"maxDiscountQty": {
						"$numberLong": "0"
					},
					"title": ""
				},
				"pricing": {
					"amount": "25000",
					"qty": {
						"$numberLong": "1"
					},
					"type": "Pack"
				},
				"product": {
					"unitPerCase": {
						"$numberLong": "100"
					},
					"brand": {
						"id": "HJP8RVuYVzyeZ8KfyRXnTX4QfbNxz5L1u4SryEDQKmpf",
						"name": "Djarum"
					},
					"category": {
						"name": "Rokok",
						"id": "G8UcF5e5sfgYgqRR3tD2zfNpSUteqo4rFakZk3aCGHiP"
					},
					"company": {
						"id": "4tfN6CWnkGNSuoboZr95Mx2rwGfcTQfqDMYBUwjMMUUW",
						"name": "Djarum"
					},
					"image": "https://storage.googleapis.com/gotoko-legacy-images/product/B01BAA12365.png",
					"name": "Djarum Super 16",
					"skuNo": "1B01BAA12365",
					"tax": ""
				},
				"qty": {
					"min": {
						"$numberLong": "1"
					},
					"max": {
						"$numberLong": "100000"
					},
					"inStock": {
						"$numberLong": "9999999"
					},
					"ordered": {
						"$numberLong": "20"
					}
				},
				"id": "5zeNJMkBKujfxNDd9AQRXZ59iDYcxDuYYKCDX2ztHWjq",
				"amount": {
					"subtotal": "500000",
					"tax": "49549.54954954954955",
					"total": "500000",
					"discount": "0"
				}
			}
		],
		"number": "GTO-1655365218977",
		"payment": {
			"name": "Tunai",
			"slug": "tunai",
			"createdAt": {
				"$numberLong": "0"
			},
			"updatedAt": {
				"$numberLong": "0"
			},
			"code": "TUNAI",
			"icon": "Tunai"
		},
		"cancelReason": "staticReason Id1655276810669",
		"eta": "2022-06-17",
		"retailer": {
			"owner": {
				"email": "yog.awigardo@gmail.com",
				"name": "TEST NAME HAMMAD vckhsd cnxkdshvjvvubjjksADhbdsjfksdafjlksdalfasdhklfvbxcnbsalzkjdfv vsdjhflksad hsdfashdlk fsbfsadlf salsaklf jkhfsaHFksajfhksfdkjjkdfhkasdhfkcsdhkfkjsdahcdsjbcsfdkjafkuvjvjvugjvigigiggigigjgjvjvjvhvhvhvuguguguvugufufuvuvuvuvuggugugug",
				"phone": "081280880226"
			},
			"user": "tPrQPYSsfDPac8R3up5G",
			"id": "6EelCyoxjTaaW2gN2GqP",
			"address": {
				"province": "banten",
				"city": "kota tangerang",
				"district": "cipondoh",
				"detail": "jalan green lake testing",
				"location": "",
				"latitude": "",
				"place": "",
				"districtSub": "cipondoh indah",
				"zipCode": "15148",
				"longitude": ""
			},
			"area": {
				"region": "JABAR - JATENG",
				"regionSub": "Bandung",
				"cluster": "Bandung 1",
				"warehouse": "bitung"
			},
			"name": "test shop name igugsdfknlhassggjf  dslkfhhdfl shasf hs sflgasfdh gewor lzxcn arhoirwfg slzdhkflksad iowaer lxv s  hsldhl sdh    sdahlkfha ewior sdlz alhwe ewo n,ghoeriz  pwjfwe hfsldafhgoar h lsakeh lsafoiseiof zlkshlk as.zdjdfdjgh  hfsda    hlskdhlk  hs fhs sdflkas  sdlha49eew slkddf a jsldk   lsdhflksa   sldfh jl la  sdlkf  lwe[iq[w sal fj zdl alj   sjfljsdf  lsdajflajs lasdfj salfjwe lsjlfj alskshfksadhfksadhflsdlfhlskadhflkasdhlkfhaslbfsadlkjhflzsjdhlkfhasdfkluvuvjvjvjvjvjvjvjvjvjgjvuguggykvjvjvjvkbjvjvjvjv",
			"number": "RT-1597202534"
		},
		"voucher": null,
		"id": "GfY8ThEbrjBraLGH7Qvc14V1rJGzo9rykR7TgJjAfD1M",
		"statusDates": {
			"created": {
				"$numberLong": "1655365218"
			},
			"processed": {
				"$numberLong": "0"
			},
			"packed": {
				"$numberLong": "0"
			},
			"intransit": {
				"$numberLong": "0"
			},
			"delivered": {
				"$numberLong": "0"
			},
			"paid": {
				"$numberLong": "0"
			},
			"cancelled": {
				"$numberLong": "1655366854"
			}
		},
		"awb": "",
		"updatedAt": {
			"$numberLong": "1655366854"
		},
		"shippingStatus": "",
		"amount": {
			"total": "547810",
			"voucher": "0",
			"discount": "0",
			"shipping": "0",
			"subtotal": "547810",
			"tax": "54287.47747747747748"
		},
		"status": {
			"$numberLong": "6"
		},
		"createdAt": {
			"$numberLong": "1655365218"
		}
	}`

	var data map[string]interface{}
	require.Nil(t, json.Unmarshal([]byte(str), &data))

	ev, err := NewEvaluator("eval: (float($amount.total) - query($., 'f:items; w: product.category.name = Rokok; s:amount.total |> sum:total ')) * 0.1")
	assert.Nil(t, err)

	o, err := ev.Eval(data)
	assert.Nil(t, err)
	assert.NotNil(t, o)

	assert.Equal(t, float64(547810-523500)*0.1, o)

	ev, err = NewEvaluator("eval: join(query($., 'f:items; s:product.skuNo |> pluck:skuNo'), ',')")
	assert.Nil(t, err)

	o, err = ev.Eval(data)
	assert.Nil(t, err, o)

	assert.Equal(t, "1B01BAA12350,1B02BAB12360,1B01BAA12365", o)
}

func TestEvalJoinPluck(t *testing.T) {
	str := `{
		"data": {
			"id": "HvhXhptnqAKyuT4fHzgQkV2A7XVxH5Vrj7y7xozQU3qV",
			"amount": {
			  "discount": "0",
			  "skuDiscount": "0",
			  "bundleDiscount": "0",
			  "shipping": "30000",
			  "subtotal": "105930",
			  "tax": "13469",
			  "total": "135930",
			  "voucher": "0",
			  "shippingTax": "2972",
			  "loyaltyPoint": "0"
			},
			"eta": "2023-01-21",
			"items": [
			  {
				"id": "5VaPHHiDVX5bUcof8m8GrcWd9PhsLcmbLnFvksNajZUo",
				"amount": {
				  "discount": "0",
				  "skuDiscount": "0",
				  "bundleDiscount": "0",
				  "subtotal": "105930",
				  "tax": "10497",
				  "total": "105930",
				  "voucherDiscount": "0",
				  "loyaltyPoint": ""
				},
				"area": {
				  "region": "jabodetabek",
				  "regionSub": "subregion1",
				  "cluster": "tangerang",
				  "warehouse": "karawaci"
				},
				"discount": {
				  "value": "0",
				  "type": "",
				  "description": "",
				  "discountCalcType": 0,
				  "discountQty": 0,
				  "discountValue": 0,
				  "maxDiscountQty": 0,
				  "title": "",
				  "unitAmount": 0
				},
				"pricing": {
				  "amount": "11770",
				  "qty": 1,
				  "type": "Strip"
				},
				"product": {
				  "brand": {
					"id": "3UFF7WZ5UyXt6Vpxs9ho7GrDXayJTpXDPVvexAtFri5C",
					"name": "Kapal Api"
				  },
				  "category": {
					"id": "7hxernyJLLjqT1XEM93oFL6o1mY74wWM2bcP3A4ngUmo",
					"name": "Sachet",
					"taxPercentage": "0.11"
				  },
				  "company": {
					"id": "Fjq2cDfj3CgY3w9TZtRfWzzxyRPpUNqYFJnSBPAxgWdD",
					"name": "Santos Jaya Abadi"
				  },
				  "image": "https://storage.googleapis.com/gotoko-legacy-images/product/A01AAI12354.png",
				  "name": "Kapal Api Special Mix 24gr (Dijual per 1 Renceng isi 10 Pcs)",
				  "skuNo": "1A01AAI12354",
				  "unitPerCase": 120,
				  "uom": "unit,kardus",
				  "unitPerKardus": 12,
				  "moqOrderNumber": 10,
				  "moqNumber": 1,
				  "moqOrderUnit": "Piece",
				  "moqUnit": "Strip",
				  "grammage": 0.27833334,
				  "volumetric": 0.00123
				},
				"qty": {
				  "min": 0,
				  "max": 100000000,
				  "inStock": 14932,
				  "ordered": 9,
				  "totalUnitOrdered": 9,
				  "uom": "unit",
				  "itemBreakdownOrdered": 0
				},
				"tax": ""
			  }
			],
			"number": "GTO-1674192918469",
			"payment": {
			  "code": "TUNAI",
			  "icon": "Tunai",
			  "name": "Tunai",
			  "slug": "tunai",
			  "createdAt": 0,
			  "updatedAt": 0
			},
			"retailer": {
			  "id": "PCKP3t5cem3CtqzMpcWjJsKBT8GtW4pndyhUdzgYZft",
			  "address": {
				"province": "bagmati",
				"city": "godawari",
				"district": "lalitpur",
				"districtSub": "pejaten timur",
				"zipCode": "12510",
				"detail": "JL nganu bla bla",
				"location": "dki jakarta jakarta selatan, pasar minggu pejaten timur, 12510",
				"latitude": "-6.276580",
				"longitude": "106.826870",
				"place": "ChIJgQ3i9yLyaS4Ru2-ry9bPPTQ"
			  },
			  "area": {
				"region": "jabodetabek",
				"regionSub": "subregion1",
				"cluster": "tangerang",
				"warehouse": "karawaci"
			  },
			  "name": "toko 12",
			  "number": "RN-1674045754506-LAT6",
			  "owner": {
				"email": "hqqqwwqammadtest1@example.com",
				"name": "+623329445777",
				"phone": "+6208294457777"
			  },
			  "user": "2TDUeCJ5zdnsDEKTQd8XcvT8mJHqcUY6fhpQBoVrUyYM",
			  "proofOfRegistration": "https://storage.googleapis.com/s-westeros-bucket-01/retailers/2021-09-02/5ac4373148320e408d4bdc8cb07d305b1630554624501.jpeg"
			},
			"status": 0,
			"statusDates": {
			  "created": 1674192918,
			  "processed": 0,
			  "packed": 0,
			  "inTransit": 0,
			  "delivered": 0,
			  "paid": 0,
			  "cancelled": 0
			},
			"voucher": null,
			"cancelReason": "",
			"user": "2TDUeCJ5zdnsDEKTQd8XcvT8mJHqcUY6fhpQBoVrUyYM",
			"isCustomOrder": false,
			"createdAt": 1674192918,
			"updatedAt": 1674192918,
			"createdBy": "2TDUeCJ5zdnsDEKTQd8XcvT8mJHqcUY6fhpQBoVrUyYM",
			"cancelValidFrom": 1674192928,
			"cancelValidUntil": 1674200128,
			"changeLog": [],
			"invoiceUrl": null,
			"displayStatus": "Pesanan diterima",
			"itemsBreakdown": {
			  "remainingItems": {
				"1A01AAI12354": {
				  "skuNo": "1A01AAI12354",
				  "unitPrice": 11770,
				  "qty": 9,
				  "unitQty": "unit",
				  "id": "5VaPHHiDVX5bUcof8m8GrcWd9PhsLcmbLnFvksNajZUo",
				  "name": "Kapal Api Special Mix 24gr (Dijual per 1 Renceng isi 10 Pcs)"
				}
			  }
			},
			"nonBundledItems": [
			  {
				"id": "5VaPHHiDVX5bUcof8m8GrcWd9PhsLcmbLnFvksNajZUo",
				"discount": {
				  "value": "0",
				  "type": "",
				  "description": "",
				  "discountCalcType": 0,
				  "discountQty": 0,
				  "discountValue": 0,
				  "maxDiscountQty": 0,
				  "title": "",
				  "unitAmount": 0
				},
				"pricing": {
				  "amount": "11770",
				  "qty": 1,
				  "type": "Strip"
				},
				"product": {
				  "brand": {
					"id": "3UFF7WZ5UyXt6Vpxs9ho7GrDXayJTpXDPVvexAtFri5C",
					"name": "Kapal Api"
				  },
				  "category": {
					"id": "7hxernyJLLjqT1XEM93oFL6o1mY74wWM2bcP3A4ngUmo",
					"name": "Sachet",
					"taxPercentage": "0.11"
				  },
				  "company": {
					"id": "Fjq2cDfj3CgY3w9TZtRfWzzxyRPpUNqYFJnSBPAxgWdD",
					"name": "Santos Jaya Abadi"
				  },
				  "image": "https://storage.googleapis.com/gotoko-legacy-images/product/A01AAI12354.png",
				  "name": "Kapal Api Special Mix 24gr (Dijual per 1 Renceng isi 10 Pcs)",
				  "skuNo": "1A01AAI12354",
				  "unitPerCase": 120,
				  "uom": "unit,kardus",
				  "unitPerKardus": 12,
				  "moqOrderNumber": 10,
				  "moqNumber": 1,
				  "moqOrderUnit": "Piece",
				  "moqUnit": "Strip",
				  "grammage": 0.27833334,
				  "volumetric": 0.00123
				},
				"qty": {
				  "min": 0,
				  "max": 100000000,
				  "inStock": 14932,
				  "ordered": 9,
				  "totalUnitOrdered": 9,
				  "uom": "unit",
				  "itemBreakdownOrdered": 9
				}
			  }
			],
			"loyaltyPoint": {
			  "referenceId": "",
			  "isActive": false,
			  "log": []
			},
			"fulfillment": {
			  "proofOfDelivery": "",
			  "completed": false,
			  "deliveryLogs": []
			},
			"schema_version": 2,
			"deliveryStartAt": 1674262800,
			"deliveryEndAt": 1674298800
		  },
		  "metadata": {
			"event": "order_created",
			"hash": "olXJBUG9qAs4pM8PysN3xMGIuiKVZ98Yjpr45OoDsAc=",
			"header": {
			  "deviceId": "ebf0cc2544e54a3e",
			  "user": "2TDUeCJ5zdnsDEKTQd8XcvT8mJHqcUY6fhpQBoVrUyYM"
			},
			"timestamp": "2023-01-20T05:35:19.425551467Z",
			"version": 1
		  }
	}`

	str_multi := `{
		"data": {
			"id": "Gy6cEbTfmtKCUd7NNHfN3UHadZQf2nk9YRtojCze8PE7",
			"bundledItems": {
				"id": "bundleID1",
				"name": "bundleName1"
			},
	
			"amount": {
				"discount": "11",
				"skuDiscount": "0",
				"bundleDiscount": "0",
				"shipping": "30000",
				"subtotal": "2620",
				"tax": "3231",
				"total": "32620",
				"voucher": "22",
				"shippingTax": "2972",
				"loyaltyPoint": "33"
			},
			"eta": "2023-01-12",
			"items": [
				{
					"id": "XcFtCRVmXo956afnxZJn3fZWKhb2Q6hKrQkwuiGuUPR",
					"amount": {
						"discount": "0",
						"skuDiscount": "0",
						"bundleDiscount": "0",
						"subtotal": "2620",
						"tax": "259",
						"total": "2620",
						"voucherDiscount": "0",
						"loyaltyPoint": ""
					},
					"area": {
						"region": "jabodetabek",
						"regionSub": "jabo 2",
						"cluster": "jakarta",
						"warehouse": "karawaci"
					},
					"discount": {
						"value": "0",
						"type": "",
						"description": "",
						"discountCalcType": 0,
						"discountQty": 0,
						"discountValue": 0,
						"maxDiscountQty": 0,
						"title": "",
						"unitAmount": 0
					},
					"pricing": {
						"amount": "2620",
						"qty": 1,
						"type": "Piece"
					},
					"product": {
						"brand": {
							"id": "G3LX5mirhozcS9HncnULy2rTTnsTmaSUgKQpPkxMXFwT",
							"name": "Teh Pucuk Harum"
						},
						"category": {
							"id": "DXLjwoXR2tEx8wHpbu2RPemGhooefk2GFRznaTdts1Ww",
							"name": "Minuman Siap Saji",
							"taxPercentage": "0.11"
						},
						"company": {
							"id": "2zs8CWdMybHCfLEk9jKeF7UMrbfogvNC21Y4VJeWyArq",
							"name": "Mayora"
						},
						"image": "https://storage.googleapis.com/gotoko-legacy-images/product/A01AAH12351.png",
						"name": "Teh Pucuk Harum Teh Melati 350ml (Dijual per 1 Pcs)",
						"skuNo": "1A01AAH12351",
						"unitPerCase": 24,
						"uom": "unit,kardus",
						"unitPerKardus": 24,
						"moqOrderNumber": 1,
						"moqNumber": 1,
						"moqOrderUnit": "Piece",
						"moqUnit": "Piece",
						"grammage": 0.38333333,
						"volumetric": 0.0007
					},
					"qty": {
						"min": 0,
						"max": 100000000,
						"inStock": 37553,
						"ordered": 1,
						"totalUnitOrdered": 1,
						"uom": "unit",
						"itemBreakdownOrdered": 0
					},
					"tax": "",
					"displayfield": {
						"remainingStock": "",
						"qtyPerUnit": "1 piece",
						"unitConversion": "Delivered in 1 piece"
					}
				},
				{
					"id": "testitem-2",
					"amount": {
						"discount": "01",
						"skuDiscount": "02",
						"bundleDiscount": "0",
						"subtotal": "3620",
						"tax": "259",
						"total": "3620",
						"voucherDiscount": "0",
						"loyaltyPoint": ""
					},
					"area": {
						"region": "jabodetabek",
						"regionSub": "jabo 2",
						"cluster": "jakarta",
						"warehouse": "karawaci"
					},
					"discount": {
						"value": "0",
						"type": "",
						"description": "",
						"discountCalcType": 0,
						"discountQty": 0,
						"discountValue": 0,
						"maxDiscountQty": 0,
						"title": "",
						"unitAmount": 0
					},
					"pricing": {
						"amount": "2620",
						"qty": 1,
						"type": "Piece"
					},
					"product": {
						"brand": {
							"id": "G3LX5mirhozcS9HncnULy2rTTnsTmaSUgKQpPkxMXFwT",
							"name": "Teh Pucuk Harum"
						},
						"category": {
							"id": "DXLjwoXR2tEx8wHpbu2RPemGhooefk2GFRznaTdts1Ww",
							"name": "Minuman Siap Saji",
							"taxPercentage": "0.11"
						},
						"company": {
							"id": "2zs8CWdMybHCfLEk9jKeF7UMrbfogvNC21Y4VJeWyArq",
							"name": "Mayora"
						},
						"image": "https://storage.googleapis.com/gotoko-legacy-images/product/A01AAH12351.png",
						"name": "Test Sku Name-2",
						"skuNo": "TestSku-2",
						"unitPerCase": 24,
						"uom": "unit,kardus",
						"unitPerKardus": 24,
						"moqOrderNumber": 1,
						"moqNumber": 1,
						"moqOrderUnit": "Piece",
						"moqUnit": "Piece",
						"grammage": 0.38333333,
						"volumetric": 0.0007
					},
					"qty": {
						"min": 0,
						"max": 100000000,
						"inStock": 37550,
						"ordered": 3,
						"totalUnitOrdered": 3,
						"uom": "unit",
						"itemBreakdownOrdered": 0
					},
					"tax": "",
					"displayfield": {
						"remainingStock": "",
						"qtyPerUnit": "1 piece",
						"unitConversion": "Delivered in 1 piece"
					}
				}
	
			],
			"number": "GTO-1673357618156",
			"payment": {
				"code": "TUNAI",
				"icon": "Tunai",
				"name": "Tunai",
				"slug": "tunai",
				"createdAt": 0,
				"updatedAt": 0
			},
			"retailer": {
				"id": "2ywjcQmWJKq8fchpbLQ5L2SBa8MTkhWaTVuTCv2Z6swU",
				"address": {
					"province": "dki jakarta",
					"city": "jakarta utara",
					"district": "penjaringan",
					"districtSub": "penjaringan",
					"zipCode": "14440",
					"detail": " jl. pluit dalam 3 no.6 rt.12/rw.6 penjaringan kec. penjaringan kota jkt utara daerah khusus ibukota jakarta 14440 indonesia",
					"location": "",
					"latitude": "-6.112852966813288",
					"longitude": "106.77834531389269",
					"place": ""
				},
				"area": {
					"region": "jabodetabek",
					"regionSub": "jabo 2",
					"cluster": "jakarta",
					"warehouse": "karawaci"
				},
				"name": "warung ibu hani",
				"number": "RN-1661965274-Tw39",
				"owner": {
					"email": "88296431398@gmail.com",
					"name": "Toko Nina",
					"phone": "88296431398"
				},
				"user": "413pLLdbV9Fa2cTYGe3T7anqtKUCJTxUesiteB5nhc7p",
				"proofOfRegistration": ""
			},
			"status": 0,
			"statusDates": {
				"created": 1673357618,
				"processed": 0,
				"packed": 0,
				"inTransit": 0,
				"delivered": 0,
				"paid": 0,
				"cancelled": 0
			},
			"voucher": {
				"code" : "voucherCode1"
			},
			"cancelReason": "",
			"user": "413pLLdbV9Fa2cTYGe3T7anqtKUCJTxUesiteB5nhc7p",
			"isCustomOrder": false,
			"createdAt": 1673357618,
			"updatedAt": 1673357618,
			"createdBy": "413pLLdbV9Fa2cTYGe3T7anqtKUCJTxUesiteB5nhc7p",
			"cancelValidFrom": 1673357628,
			"cancelValidUntil": 1673364828,
			"changeLog": [],
			"invoiceUrl": null,
			"displayStatus": "Pesanan diterima",
			"itemsBreakdown": {
				"remainingItems": {
					"1A01AAH12351": {
						"skuNo": "1A01AAH12351",
						"unitPrice": 2620,
						"qty": 1,
						"unitQty": "unit",
						"id": "XcFtCRVmXo956afnxZJn3fZWKhb2Q6hKrQkwuiGuUPR",
						"name": "Teh Pucuk Harum Teh Melati 350ml (Dijual per 1 Pcs)"
					}
				}
			},
			"nonBundledItems": [
				{
					"id": "XcFtCRVmXo956afnxZJn3fZWKhb2Q6hKrQkwuiGuUPR",
					"discount": {
						"value": "0",
						"type": "",
						"description": "",
						"discountCalcType": 0,
						"discountQty": 0,
						"discountValue": 0,
						"maxDiscountQty": 0,
						"title": "",
						"unitAmount": 0
					},
					"pricing": {
						"amount": "2620",
						"qty": 1,
						"type": "Piece"
					},
					"product": {
						"brand": {
							"id": "G3LX5mirhozcS9HncnULy2rTTnsTmaSUgKQpPkxMXFwT",
							"name": "Teh Pucuk Harum"
						},
						"category": {
							"id": "DXLjwoXR2tEx8wHpbu2RPemGhooefk2GFRznaTdts1Ww",
							"name": "Minuman Siap Saji",
							"taxPercentage": "0.11"
						},
						"company": {
							"id": "2zs8CWdMybHCfLEk9jKeF7UMrbfogvNC21Y4VJeWyArq",
							"name": "Mayora"
						},
						"image": "https://storage.googleapis.com/gotoko-legacy-images/product/A01AAH12351.png",
						"name": "Teh Pucuk Harum Teh Melati 350ml (Dijual per 1 Pcs)",
						"skuNo": "1A01AAH12351",
						"unitPerCase": 24,
						"uom": "unit,kardus",
						"unitPerKardus": 24,
						"moqOrderNumber": 1,
						"moqNumber": 1,
						"moqOrderUnit": "Piece",
						"moqUnit": "Piece",
						"grammage": 0.38333333,
						"volumetric": 0.0007
					},
					"qty": {
						"min": 0,
						"max": 100000000,
						"inStock": 37553,
						"ordered": 1,
						"totalUnitOrdered": 1,
						"uom": "unit",
						"itemBreakdownOrdered": 1
					},
					"displayfield": {
						"remainingStock": "",
						"qtyPerUnit": "1 piece",
						"unitConversion": "Delivered in 1 piece"
					}
				}
			],
			"loyaltyPoint": {
				"referenceId": "",
				"isActive": true,
				"log": []
			},
			"fulfillment": {
				"proofOfDelivery": "",
				"completed": false,
				"deliveryLogs": []
			},
			"schema_version": 2
		},
		"metadata": {
			"header" : {
				"deviceId": "f29720480d490d42"
			},
			"user": "ByTASRE3S6gjsvE5yUJ2QTywx9Se1FJjRDxdLkWbWfB3",
			"event": "order_created",
			"hash": "hUJbsnOhox4d1x3FMEc/24SD7of7gI2km1D1s2yDXu0=",
			"timestamp": "2023-01-10T13:33:39.280770774Z",
			"version": 2
		}
	}`

	var data map[string]interface{}
	require.Nil(t, json.Unmarshal([]byte(str), &data))

	var data_multi map[string]interface{}
	require.Nil(t, json.Unmarshal([]byte(str_multi), &data_multi))

	ev, err := NewEvaluator("eval: join(query($., 'f:data.items; s:product.skuNo |> pluck:skuNo'), ',')")
	assert.Nil(t, err)

	o, err := ev.Eval(data)
	assert.Nil(t, err, o)

	assert.Equal(t, "1A01AAI12354", o)

	o, err = ev.Eval(data_multi)
	assert.Nil(t, err, o)

	assert.Equal(t, "1A01AAH12351,TestSku-2", o)

}

func TestEvalStringFn(t *testing.T) {

	ev, err := NewEvaluator("trim($.)")
	assert.Nil(t, err)
	o, err := ev.Eval("  hello world  ")
	assert.Nil(t, err)
	assert.NotNil(t, o)

	assert.Equal(t, "hello world", o)

	ev, err = NewEvaluator("trim($., 'hello', 'prefix')")
	assert.Nil(t, err)
	o, err = ev.Eval("hello world")
	assert.Nil(t, err)
	assert.NotNil(t, o)

	assert.Equal(t, " world", o)

	ev, err = NewEvaluator("trim($., 'world', 'suffix')")
	assert.Nil(t, err)
	o, err = ev.Eval("hello world")
	assert.Nil(t, err)
	assert.NotNil(t, o)

	assert.Equal(t, "hello ", o)

	ev, err = NewEvaluator("trim($., 'hello')")
	assert.Nil(t, err)
	o, err = ev.Eval("hello world hello")
	assert.Nil(t, err)
	assert.NotNil(t, o)

	assert.Equal(t, " world ", o)

	ev, err = NewEvaluator("replace($., 'hello', '')")
	assert.Nil(t, err)
	o, err = ev.Eval("hello world hello")
	assert.Nil(t, err)
	assert.NotNil(t, o)

	assert.Equal(t, " world ", o)

	ev, err = NewEvaluator("case($.)")
	assert.Nil(t, err)
	o, err = ev.Eval("hello world")
	assert.Nil(t, err)
	assert.NotNil(t, o)

	assert.Equal(t, "Hello World", o)

	ev, err = NewEvaluator("float($balance) - float($amount, 0)")
	assert.Nil(t, err)
	o, err = ev.Eval(map[string]interface{}{"balance": "1000"})
	assert.Nil(t, err)
	assert.NotNil(t, o)

	assert.Equal(t, float64(1000), o)

	ev, err = NewEvaluator("round(float($balance) - float($amount, 0))")
	assert.Nil(t, err)
	o, err = ev.Eval(map[string]interface{}{"balance": "1000", "amount": "100.5"})
	assert.Nil(t, err)
	assert.NotNil(t, o)

	assert.Equal(t, float64(900), o)

	ev, err = NewEvaluator("contains($., 'hello')")
	assert.Nil(t, err)
	o, err = ev.Eval("hello world hello")
	assert.Nil(t, err)
	assert.NotNil(t, o)

	assert.True(t, o.(bool))

	ev, err = NewEvaluator("contains($., 'hello')")
	assert.Nil(t, err)
	o, err = ev.Eval([]string{"hello", "world"})
	assert.Nil(t, err)
	assert.NotNil(t, o)

	assert.True(t, o.(bool))

	ev, err = NewEvaluator("split($., ' ')")
	assert.Nil(t, err)
	o, err = ev.Eval("hello world hello")
	assert.Nil(t, err)
	assert.NotNil(t, o)

	assert.Equal(t, []string{"hello", "world", "hello"}, o)

	ev, err = NewEvaluator("split($., ' ', 1)")
	assert.Nil(t, err)
	o, err = ev.Eval("hello world hello")
	assert.Nil(t, err)
	assert.NotNil(t, o)

	assert.Equal(t, "world", o)

	ev, err = NewEvaluator("join($., ' ')")
	assert.Nil(t, err)
	o, err = ev.Eval([]string{"hello", "world", "hello"})
	assert.Nil(t, err)
	assert.NotNil(t, o)

	assert.Equal(t, "hello world hello", o)

	ev, err = NewEvaluator("sum($.)")
	assert.Nil(t, err)
	o, err = ev.Eval([]int64{1, 2, 3, 4, 5})
	assert.Nil(t, err)
	assert.NotNil(t, o)

	assert.Equal(t, float64(15), o)

	ev, err = NewEvaluator("sum($., 'avg')")
	assert.Nil(t, err)
	o, err = ev.Eval([]int64{1, 2, 3, 4, 5})
	assert.Nil(t, err)
	assert.NotNil(t, o)

	assert.Equal(t, float64(3), o)

	ev, err = NewEvaluator("sum($., 'min')")
	assert.Nil(t, err)
	o, err = ev.Eval([]int64{1, 2, 3, 4, 5})
	assert.Nil(t, err)
	assert.NotNil(t, o)

	assert.Equal(t, float64(1), o)

	ev, err = NewEvaluator("sum($., 'max')")
	assert.Nil(t, err)
	o, err = ev.Eval([]int64{1, 2, 3, 4, 5})
	assert.Nil(t, err)
	assert.NotNil(t, o)

	assert.Equal(t, float64(5), o)
}

func TestJoinMerge(t *testing.T) {
	joinStr := `
	{
		"type": "MultiPolygon",
		"coordinates": [
		  [
			[
			  [
				101.72236111000007,
				-1.241361110999947
			  ],
			  [
				101.72222060000007,
				-1.241285090999952
			  ],
			  [
				101.72214338500004,
				-1.241276204999963
			  ],
			  [
				101.72203647000003,
				-1.241264359999946
			  ],
			  [
				101.72197111900005,
				-1.241243504999943
			  ],
			  [
				101.72181357800008,
				-1.24108825199994
			  ],
			  [
				101.72177272800008,
				-1.241055303999929
			  ],
			  [
				101.72242112600009,
				-1.240934467999978
			  ],
			  [
				101.722419,
				-1.241011998999966
			  ],
			  [
				101.72236111000007,
				-1.241361110999947
			  ]
			]
		  ]
		]
	  }`

	data := make(map[string]interface{})
	assert.Nil(t, json.Unmarshal([]byte(joinStr), &data))

	ev, err := NewEvaluator("'(' + join(unwrapslice($coordinates.0.0), ',') + ')'")
	assert.Nil(t, err)
	assert.NotNil(t, ev)

	o, err := ev.Eval(data)
	assert.Nil(t, err)
	assert.NotNil(t, o)

	assert.Equal(t, "(101.72236111000007,-1.241361110999947,101.72222060000007,-1.241285090999952,101.72214338500004,-1.241276204999963,101.72203647000003,-1.241264359999946,101.72197111900005,-1.241243504999943,101.72181357800008,-1.24108825199994,101.72177272800008,-1.241055303999929,101.72242112600009,-1.240934467999978,101.722419,-1.241011998999966,101.72236111000007,-1.241361110999947)", o)

	out, err := geoconvert("geojson", "wkt", joinStr)
	assert.Nil(t, err)
	assert.NotNil(t, out)
	assert.Equal(t, "MULTIPOLYGON(((101.72236111000007 -1.241361110999947,101.72222060000007 -1.241285090999952,101.72214338500004 -1.241276204999963,101.72203647000003 -1.241264359999946,101.72197111900005 -1.241243504999943,101.72181357800008 -1.24108825199994,101.72177272800008 -1.241055303999929,101.72242112600009 -1.240934467999978,101.722419 -1.241011998999966,101.72236111000007 -1.241361110999947)))", out)
}

func TestEvalBytesJSON(t *testing.T) {
	data := map[string]interface{}{
		"metadata": map[string]interface{}{
			"timestamp":          time.Now(),
			"record-type":        "control",
			"operation":          "update",
			"partition-key-type": "task-id",
			"schema-name":        "",
			"sequence":           87,
			"table-name":         "awsdms_apply_exceptions",
			"active":             "true",
			"items":              []string{"1", "2", "3"},
			"num":                "-1.23211",
		},
	}

	b, err := json.Marshal(data)
	require.Nil(t, err)
	require.NotNil(t, b)

	ev, err := NewEvaluator("$metadata.record-type == 'control' && ($metadata.operation == 'create' || $metadata.operation == 'update') && $metadata.schema-name == ''")
	assert.Nil(t, err)

	o, err := ev.EvalBytesJSON(b)
	assert.Nil(t, err)
	assert.True(t, o.(bool))

	ev, err = NewEvaluator("match('metadata.record-type', $., 'control')")
	assert.Nil(t, err)

	o, err = ev.EvalBytesJSON(b)
	assert.Nil(t, err)
	fmt.Println(o)
	assert.True(t, o.(bool))

	ev, err = NewEvaluator("eval:int($metadata.sequence)")
	assert.Nil(t, err)

	o, err = ev.EvalBytesJSON(b)
	assert.Nil(t, err)
	assert.Equal(t, 87, o)

	ev, err = NewEvaluator("eval:int($metadata.sequence)")
	assert.Nil(t, err)

	o, err = ev.EvalBytesJSON(b)
	assert.Nil(t, err)
	assert.Equal(t, 87, o)

	ev, err = NewEvaluator("eval:date()")
	assert.Nil(t, err)

	o, err = ev.EvalBytesJSON(b)
	assert.Nil(t, err)
	assert.Equal(t, time.Now().Format("2006/01/02"), o)

	ev, err = NewEvaluator("eval:date($metadata.timestamp)")
	assert.Nil(t, err)

	o, err = ev.EvalBytesJSON(b)
	assert.Nil(t, err)
	assert.Equal(t, time.Now().Format("2006/01/02"), o)

	ev, err = NewEvaluator("eval:date('2021-Mar-29', '2006-Jan-02')")
	assert.Nil(t, err)

	o, err = ev.EvalBytesJSON(b)
	assert.Nil(t, err)
	assert.Equal(t, "2021/03/29", o)

	to := time.Now()
	tb, _ := json.Marshal(to)
	ts := strings.Trim(string(tb), `"`)

	er := "eval:ftime('" + ts + "', '2006-01-02 15:04:05')"
	fmt.Println(er)

	ev, err = NewEvaluator("eval:ftime('" + ts + "', 'T2006-01-02 15:04:05')")
	assert.Nil(t, err)

	o, err = ev.EvalBytesJSON(b)
	assert.Nil(t, err)
	assert.Equal(t, time.Now().Format("2006-01-02 15:04:05"), o)

	ev, err = NewEvaluator("eval:len($metadata.items)")
	assert.Nil(t, err)

	o, err = ev.EvalBytesJSON(b)
	assert.Nil(t, err)
	assert.Equal(t, 3, o)

	ev, err = NewEvaluator("eval:float($metadata.num)")
	assert.Nil(t, err)

	o, err = ev.EvalBytesJSON(b)
	assert.Nil(t, err)
	assert.Equal(t, -1.23211, o)
}
