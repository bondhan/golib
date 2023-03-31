package util

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestRequestWrapper(t *testing.T) {

	data := map[string]interface{}{
		"name": "sahalzain",
		"age":  23,
	}

	b, err := json.Marshal(data)
	if err != nil {
		t.Error(err)
	}

	req, err := http.NewRequest("GET", "http://localhost:8080?user=1234", bytes.NewBuffer(b))
	if err != nil {
		t.Error(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rw, err := CopyRequest(req)
	if err != nil {
		t.Error(err)
	}

	if rw == nil {
		t.Error()
	}

	if rw.GetValue("body.name") != "sahalzain" {
		t.Error()
	}

	if rw.GetValue("body.age") != int64(23) {
		t.Error("actual: ", rw.GetValue("body.age"))
	}

	if rw.GetValue("query.user") != "1234" {
		t.Error("actual: ", rw.GetValue("query.user"))
	}

	if rw.GetValue("header.Content-Type") != "application/json" {
		t.Error("actual: ", rw.GetValue("header.Content-Type"))
	}

}

func TestRequestWrapperEval(t *testing.T) {

	data := map[string]interface{}{
		"name": "sahalzain",
		"age":  23,
	}

	b, err := json.Marshal(data)
	if err != nil {
		t.Error(err)
	}

	req, err := http.NewRequest("GET", "http://localhost:8080?user=1234", bytes.NewBuffer(b))
	if err != nil {
		t.Error(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rw, err := CopyRequest(req)
	if err != nil {
		t.Error(err)
	}

	if rw == nil {
		t.Error()
	}

	if rw.GetValue("body.name") != "sahalzain" {
		t.Error()
	}

	if rw.GetValue("body.age") != int64(23) {
		t.Error("actual: ", rw.GetValue("body.age"))
	}

	if rw.GetValue("query.user") != "1234" {
		t.Error("actual: ", rw.GetValue("query.user"))
	}

	if rw.GetValue("header.Content-Type") != "application/json" {
		t.Error("actual: ", rw.GetValue("header.Content-Type"))
	}

	if rw.GetValue(":$body.age > 25 ? 'adult' : 'child'") != "child" {
		t.Error("actual: ", rw.GetValue(":$body.age > 25 ? 'adult' : 'child'"))
	}

}

func TestRequestWrapperScopes(t *testing.T) {

	data := map[string]interface{}{
		"name": "sahalzain",
		"age":  23,
	}

	b, err := json.Marshal(data)
	if err != nil {
		t.Error(err)
	}

	req, err := http.NewRequest("GET", "http://localhost:8080?user=1234", bytes.NewBuffer(b))
	if err != nil {
		t.Error(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rw, err := CopyRequest(req, "header", "query")
	if err != nil {
		t.Error(err)
	}

	if rw == nil {
		t.Error()
	}

	if rw.GetValue("body.name") == "sahalzain" {
		t.Error()
	}

	if rw.GetValue("body.age") == int64(23) {
		t.Error("actual: ", rw.GetValue("body.age"))
	}

	if rw.GetValue("query.user") != "1234" {
		t.Error("actual: ", rw.GetValue("query.user"))
	}

	if rw.GetValue("header.Content-Type") != "application/json" {
		t.Error("actual: ", rw.GetValue("header.Content-Type"))
	}
}

func TestGenerateID(t *testing.T) {
	payload := map[string]interface{}{
		"username":  "sahalzain",
		"name":      "Sahal Zain",
		"age":       35,
		"city":      "Jogjakarta",
		"timestamp": time.Now().UnixNano(),
	}

	b, err := json.Marshal(payload)
	if err != nil {
		t.Error(err)
		return
	}

	req, err := http.NewRequest("POST", "http://localhost:8080/user/12?replace=true", bytes.NewBuffer(b))
	if err != nil {
		t.Error(err)
		return
	}

	rw, err := CopyRequest(req)
	if err != nil {
		t.Error(err)
		return
	}

	id := GenerateRequestID(req, nil, nil)
	if len(id) != 32 {
		t.Error("should be 32 byte ", len(id))
		return
	}

	payload2 := map[string]interface{}{
		"username":  "sahalzain",
		"name":      "Sahal Zain",
		"age":       35,
		"city":      "Jogjakarta",
		"timestamp": time.Now().UnixNano(),
	}

	b2, err := json.Marshal(payload2)
	if err != nil {
		t.Error(err)
		return
	}

	req2, err := http.NewRequest("POST", "http://localhost:8080/user/12?replace=true", bytes.NewBuffer(b2))
	if err != nil {
		t.Error(err)
		return
	}

	rw2, err := CopyRequest(req2)
	if err != nil {
		t.Error(err)
		return
	}

	id2 := GenerateRequestID(req2, nil, nil)
	if len(id2) != 32 {
		t.Error("should be 32 byte ", len(id2))
		return
	}

	if bytes.Equal(id, id2) {
		t.Error("should be different")
	}

	if bytes.Equal(rw.GenerateRequestID(nil, nil, nil), rw2.GenerateRequestID(nil, nil, nil)) {
		t.Error("should be different")
	}

	id = GenerateRequestID(req, nil, []string{"timestamp"})
	id2 = GenerateRequestID(req2, nil, []string{"timestamp"})

	if !bytes.Equal(id, id2) {
		t.Error("should be same")
	}

	if !bytes.Equal(rw.GenerateRequestID(nil, nil, []string{"timestamp"}), rw2.GenerateRequestID(nil, nil, []string{"timestamp"})) {
		t.Error("should be same")
	}

}

func TestRequestWrapperCookieScopes(t *testing.T) {

	data := map[string]interface{}{
		"name": "sahalzain",
		"age":  23,
	}

	b, err := json.Marshal(data)
	if err != nil {
		t.Error(err)
	}

	req, err := http.NewRequest("GET", "http://localhost:8080?user=1234", bytes.NewBuffer(b))
	if err != nil {
		t.Error(err)
	}
	req.AddCookie(&http.Cookie{Name: "username", Value: "sahalzain"})
	req.AddCookie(&http.Cookie{Name: "token", Value: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"})
	req.Header.Set("Content-Type", "application/json")

	rw, err := CopyRequest(req, "*", WithCookieTokenScopeOpt("token"))
	if err != nil {
		t.Error(err)
	}

	if rw == nil {
		t.Error("request wrapper should not nil")
	}

	if rw.GetValue("jwt.sub") != "1234567890" {
		t.Error("actual: ", rw.GetValue("jwt.sub"))
	}

	if rw.GetValue("cookie.username") != "sahalzain" {
		t.Error("actual: ", rw.GetValue("cookie.username"))
	}

	if rw.GetValue("header.Content-Type") != "application/json" {
		t.Error("actual: ", rw.GetValue("header.Content-Type"))
	}

	if rw.GetValue("query.user") != "1234" {
		t.Error("actual: ", rw.GetValue("query.user"))
	}
}
