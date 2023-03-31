package aclpath

type AclPath struct {
	ID       string     `json:"id"`
	Method   string     `json:"method"`
	Path     string     `json:"path"`
	Request  *Modifiers `json:"request,omitempty"`
	Response *Modifiers `json:"response,omitempty"`
}

type Action struct {
	Kind  string      `json:"kind"`
	Key   string      `json:"key"`
	Value interface{} `json:"value,omitempty"`
	Type  string      `json:"type,omitempty"`
}

type Modifiers struct {
	Query   *Modifier `json:"query"`
	Headers *Modifier `json:"header"`
	Body    *Modifier `json:"body"`
}

type Modifier struct {
	Actions []*Action `json:"actions"`
}
