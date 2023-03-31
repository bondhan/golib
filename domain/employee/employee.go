package employee

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"

	"github.com/bondhan/golib/util"
)

const EmployeeHeaderKey = "X-Employee"

type ctxKey string

func (c ctxKey) String() string {
	return "employee context " + string(c)
}

const employeeCtx = ctxKey("employeeContext")

func DecodeEmployee(employee interface{}) (*EmployeeContext, error) {
	switch r := employee.(type) {
	case *EmployeeContext:
		return r, nil
	case EmployeeContext:
		return &r, nil
	default:
		var ret EmployeeContext
		if err := util.DecodeJSON(employee, &ret); err != nil {
			return nil, err
		}
		return &ret, nil
	}
}

func DecodeEmployeeFromString(employee string) (*EmployeeContext, error) {
	b, err := base64.StdEncoding.DecodeString(employee)
	if err != nil {
		return nil, err
	}

	var tmp map[string]interface{}
	if err := json.Unmarshal(b, &tmp); err != nil {
		return nil, err
	}

	return DecodeEmployee(tmp)
}

func GetEmployeeFromHeaders(headers map[string]string) (*EmployeeContext, error) {
	employee, ok := headers[EmployeeHeaderKey]
	if !ok {
		return nil, errors.New("[filter/employee] employee not found in headers")
	}

	return DecodeEmployeeFromString(employee)
}

func GetEmployeeFromContext(ctx context.Context) (*EmployeeContext, error) {
	if str, ok := ctx.Value(employeeCtx).(string); ok {
		return DecodeEmployeeFromString(str)
	}
	return nil, errors.New("[filter/employee] employee not found in context")
}

func GetEncodedEmployeeFromContext(ctx context.Context) string {
	if str, ok := ctx.Value(employeeCtx).(string); ok {
		return str
	}

	return ""
}

func GetRawEmployeeFromContext(ctx context.Context) []byte {
	if str, ok := ctx.Value(employeeCtx).(string); ok {
		b, err := base64.StdEncoding.DecodeString(str)
		if err != nil {
			return nil
		}
		return b
	}
	return nil
}

func PropagateEmployee(ctx context.Context, headers map[string]string, key string) (context.Context, error) {
	if key == "" {
		key = EmployeeHeaderKey
	}

	employee, ok := headers[key]
	if !ok {
		return nil, errors.New("[filter/employee] employee not found in headers")
	}

	return context.WithValue(ctx, employeeCtx, employee), nil
}

func WithEmployeeContext(ctx context.Context, ret *EmployeeContext) context.Context {
	str, err := ret.Encode()
	if err != nil {
		return ctx
	}
	return context.WithValue(ctx, employeeCtx, str)
}

func WithEmployeeStringContext(ctx context.Context, ret string) context.Context {
	return context.WithValue(ctx, employeeCtx, ret)
}

type EmployeeContext struct {
	ID              string                   `json:"id" firestore:"id"`
	Department      *EmployeeDepartment      `json:"department,omitempty" firestore:"department"`
	DepartmentLevel *EmployeeDepartmentLevel `json:"departmentLevel,omitempty" firestore:"departmentLevel"`
	Fullname        string                   `json:"fullname,omitempty" firestore:"fullname"`
	User            string                   `json:"user,omitempty" firestore:"user"`
	CreatedAt       int                      `json:"createdAt,omitempty" firestore:"createdAt"`
	UpdatedAt       int                      `json:"updatedAt,omitempty" firestore:"updatedAt"`
}

type EmployeeDepartment struct {
	Name string `json:"name,omitempty" firestore:"name"`
	Slug string `json:"slug,omitempty" firestore:"slug"`
}
type EmployeeDepartmentLevel struct {
	Name string `json:"name,omitempty" firestore:"name"`
	Slug string `json:"slug,omitempty" firestore:"slug"`
}

func (r *EmployeeContext) Encode() (string, error) {
	b, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}
