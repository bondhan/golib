package marketing

import (
	"github.com/bondhan/golib/util"
)

type Promo interface {
	GetFilteredAttributes() map[string]interface{}
	GetID() string
}

type DummyPromo struct {
	id         string
	attributes map[string]interface{}
}

func NewDummyPromo(id string) *DummyPromo {
	return &DummyPromo{
		id:         id,
		attributes: make(map[string]interface{}),
	}
}

func (p *DummyPromo) GetID() string {
	return p.id
}

func (p *DummyPromo) GetFilteredAttributes() map[string]interface{} {
	return p.attributes
}

func (p *DummyPromo) SetAttributes(key string, value interface{}) {
	p.attributes[key] = value
}

func Decode(p []interface{ Promo }, v interface{}) error {
	return util.DecodeJSON(p, &v)
}
