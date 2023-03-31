package errorlib

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type single struct {
	ServiceCode string
}

var lock = &sync.Mutex{}
var singleInstance *single

func SetServiceCode(code string) *single {
	if singleInstance == nil {
		lock.Lock()
		defer lock.Unlock()
		if singleInstance == nil {
			singleInstance = &single{
				ServiceCode: code,
			}
		}
	}

	return singleInstance
}

type BaseError struct {
	ServiceCode string
	Code        string
	Message     string
	GRPCCode    codes.Code
	VarsString  []string
}

func (e *BaseError) Error() error {
	result := ""
	for index := 0; index < len(e.VarsString); index++ {
		vars := fmt.Sprint(e.VarsString[index])
		result += vars + Delimeter
	}
	result = strings.TrimRight(result, Delimeter)
	message := fmt.Sprintf("%s|%s|%s", e.Code, e.Message, result)

	return status.Error(e.GRPCCode, message)
}

func ComposeErr(grpcCode codes.Code, code int, errMsg error, vars ...interface{}) error {
	var str []string
	for _, v := range vars {
		str = append(str, fmt.Sprint(v))
	}

	// result will be AUT + 00001
	errorCode := singleInstance.ServiceCode + fmt.Sprintf("%05d", code)
	err := &BaseError{
		GRPCCode:   grpcCode,
		Code:       errorCode,
		Message:    errMsg.Error(),
		VarsString: str}
	return err.Error()
}

func DecodeErr(errMsg error) (*BaseError, error) {
	// check if errMsg is nil
	if errMsg == nil {
		return nil, errors.New("error message nil")
	}

	result := status.Convert(errMsg)
	if result == nil {
		return nil, errors.New("error conversion failed")
	}

	resultSplit := strings.Split(result.Message(), "|")
	if len(resultSplit) != 3 {
		return nil, errors.New("error conversion failed")
	}

	var code, message string
	var vars []string

	codeInt, err := strconv.Atoi(resultSplit[0][3:])
	if err != nil {
		return nil, err
	}
	code = fmt.Sprintf("%v", codeInt)
	message = resultSplit[1]
	data := resultSplit[2]
	if data != "" {
		vars = strings.Split(data, Delimeter)
	}

	return &BaseError{
		ServiceCode: resultSplit[0],
		Code:        code,
		Message:     message,
		GRPCCode:    result.Code(),
		VarsString:  vars,
	}, nil
}
