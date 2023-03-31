package authtoken

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"reflect"
	"strings"
	"testing"
	"time"
)

func Test_sign(t *testing.T) {

	t.Run("should return token with correct expiry time", func(t *testing.T) {
		got, err := sign(60*time.Second, "testIssuer", "testSubject", []string{"employee", "partner"}, []string{"testAud"}, getTestKeyPair(0))
		if err != nil {
			t.Errorf("sign() error = %v", err)
			return
		}

		claims, err := verify(got, &getTestKeyPair(0).PublicKey)
		if err != nil {
			t.Errorf("verify() error = %v", err)
			return
		}

		if !claims.ExpiresAt.After(time.Now()) || !claims.ExpiresAt.Before(time.Now().Add(time.Minute)) {
			t.Errorf("claims.ExpiresAt = %v, want %v", claims.ExpiresAt, time.Now().Add(time.Minute))
		}

	})

	t.Run("verifier should verify expiry time", func(t *testing.T) {
		got, err := sign(0, "testIssuer", "testSubject", []string{"employee", "partner"}, []string{"testAud"}, getTestKeyPair(0))
		if err != nil {
			t.Errorf("sign() error = %v", err)
			return
		}

		t.Log(got)
		_, err = verify(got, &getTestKeyPair(0).PublicKey)
		if err == nil || !strings.Contains(err.Error(), "token is expired") {
			t.Errorf("verify() error = %v", err)
			return
		}

	})

	t.Run("signed token should contain all the fields", func(t *testing.T) {
		got, err := sign(60*time.Second, "testIssuer", "testSubject", []string{"employee", "partner"}, []string{"testAud"}, getTestKeyPair(0))
		if err != nil {
			t.Errorf("sign() error = %v", err)
			return
		}
		claims, err := verify(got, &getTestKeyPair(0).PublicKey)
		if err != nil {
			t.Errorf("verify() error = %v", err)
			return
		}

		if claims.Issuer != "testIssuer" {
			t.Errorf("claims.Issuer = %v, want %v", claims.Issuer, "testIssuer")
		}

		if claims.Subject != "testSubject" {
			t.Errorf("claims.Subject = %v, want %v", claims.Subject, "testSubject")
		}

		if len(claims.Audience) != 1 || claims.Audience[0] != "testAud" {
			t.Errorf("claims.Audience = %v, want %v", claims.Audience, "testAud")
		}

	})

}

func TestSignToken(t *testing.T) {
	type args struct {
		aud      string
		userID   string
		roles    []string
		duration time.Duration
		key      *rsa.PrivateKey
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"should return token with correct expiry time",
			args{
				aud:      "testAud",
				userID:   "testUID",
				duration: time.Minute,
				roles:    []string{"employee", "partner"},
				key:      getTestKeyPair(0),
			},

			false,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SignToken(tt.args.roles, tt.args.aud, tt.args.userID, tt.args.duration, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("SignToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.Token == "" {
				t.Errorf("SignToken() = %v, want %v", got, "not empty")
			}

			claims, err := verify(got.Token, &tt.args.key.PublicKey)
			if err != nil {
				t.Errorf("verify() error = %v", err)
				return
			}

			if claims.Issuer != "auth-service" {
				t.Errorf("claims.Issuer = %v, want %v", claims.Issuer, "auth-service")
			}

			if claims.Subject != tt.args.userID {
				t.Errorf("claims.Subject = %v, want %v", claims.Subject, tt.args.userID)
			}

			if len(claims.Audience) != 1 || claims.Audience[0] != tt.args.aud {
				t.Errorf("claims.Audience = %v, want %v", claims.Audience, tt.args.aud)
			}

			if !claims.ExpiresAt.After(time.Now()) || !claims.ExpiresAt.Before(time.Now().Add(tt.args.duration)) {
				t.Errorf("claims.ExpiresAt = %v, want %v", claims.ExpiresAt, time.Now().Add(tt.args.duration))
			}

			if !reflect.DeepEqual(claims.Roles, tt.args.roles) {
				t.Errorf("claims.Roles = %v, want %v", claims.Roles, tt.args.roles)
			}

		})
	}
}

func getTestKeyPair(n int) *rsa.PrivateKey {
	block, _ := pem.Decode([]byte(testKeyPairs[n]))
	der := block.Bytes
	key, _ := x509.ParsePKCS1PrivateKey(der)
	return key
}

var testKeyPairs = map[int]string{
	0: `-----BEGIN RSA PRIVATE KEY-----
MIIEpQIBAAKCAQEA2ESq09eimKlsV+Euq044z9LrQntn2E4htI+GMs2hp7YaIMRN
boH6jK9eEHqXr/XjjdlPzne3O8N90M2O+u0XSzBI09TNHj/WDrqIa5stNDrjg4B8
ejIENzYv/pty3CcT3S+KiTQaDGNsCMprUzAaOJVwYTlkFIgrl9+o03ul1XUSLBVg
oslzhrn4y50s4pLaCpZpKMsZvL5jL0+XO4y/D97Wl6D10wKWGio5aS0kM9PRlAhW
HQIfFEpj8ycxgADBInmAI9EGcwkEo88ZicQX3fSNp8txmfl2VVTa8ti2NhduBoGP
32+1C7PkIvRUFaBexyYESMT7uCWmk7pa9FzHKQIDAQABAoIBAQCiu7jXQvEkcogF
8HmPiYKSTyGbmwRe7RvLpBvU6opzikFK2qtxNfj0F5luSPEPBLU/rX+x2r9UBxwL
HEI0amcWurRyQTQ/SMWnu5Cfd9qh7JX6A5qm0C//45Rlv07Efdsimn1iFIRtQfqu
+rGbnRx962Tuo6K9GGHFHBULXYVBgj3iH/HGxd6SQiZpuL9ykAazBMZde9/+tx8y
3Ru7vIjka7sYZ+sycokxW/BGgiGEQKIYqlM0yp7xTa8mtcFR/+BUYNhuwtGPh6rq
S/Q0cW+G0p+XLBF0D+Q+i5YodyV3GCyL3unt42R31O2POeD34FgG/bcVE16QiE/4
VGK/TMSJAoGBAPOWelrwJ/s7lFKW4OEueIR3Y2m35PTm5thBDRpmEn6lvNEaK+pj
kzXDf+spULuAUsZLUBCnVVmm7ad3b/OEiOcZyQaIOgLePYiG4EUL2RZPMtiTtXH0
+muy4c+ZuJLNQLcxCfKXOSZSxXlFEBuBeP9pO80NQjxsdyf4WEBT/rhLAoGBAONJ
0Iit01RukXesK08kyUrwHU08P5S7+TaivHfInL/NYB8SW2RGp3l8QRTaXCzdesLJ
fbmyHKyyVGBDqqiPNEVIhmFgjyceWHSTVxkuGX+z+hpG+NoR3A+9jxAaYK0MYolW
dilw4nqFrSLChrnfY3odkNqIq7U2/B8bGzJaGf3bAoGBAJTgWKHx/A2yCWI88u7O
DzyvF2SKz3XbFiYABDkpP46GT5PhkgusllGazjj1RHGE6ZJmf2XeT+z+eGwNNLA6
Rc4xVUsXwZT1Ldiezr2Ek0buWt5B0Pj4SIHAkADpLAUVS8NrRnAtevFwT19iFYkq
JcC9GZ6mxt/VTzJvt8iBTcJnAoGBAMDv1+UuKUZy1WkQ7XKxd231heoaSp6nMlX+
rp2/3c+zNvUpUAs/Lsshft2EvtoW6C6Re/g2CcFPX/CXgDa12Vk2x1vB68L5L31F
1Zm6WErfLF9B/9ydbICwGNFCku5SpRKQIp3rBVWcQ+xN1K/TwU3X6y6W9atOkZaW
G/ASLB4hAoGAMOLT2BiEub67tj2wfavpNXTUEcOrS9kxtUAxthXz5BoFUMPoHWCh
8SaEkt2OQYED10uaC1e238toa9V7WTaqgeQKtmuB3Qw0UaXrN0XdeQUJzppyFw/o
pviFCVK0g4wg8i0au+8CpUWD4ZLx9U93GZfcNawpPYNzrBfeylUIl9k=
-----END RSA PRIVATE KEY-----
`,
	1: `-----BEGIN RSA PRIVATE KEY-----
MIICWgIBAAKBgGvzGgMZ+79+jXKq1Xu8Ti/REUEk+OptSRiYJKId6pCscbZ5JVAy
1xh1qbbkH0U9Hgb8/1WgTPR4q5I4w/KcVdKBgOp2SuajroH+IXuqqmMuUgHlDbAO
uC2b69s2o/yPSuVp/ZoS/qGPbR9md+kHjsS+DFAnSz4A3/d7xxocOuLJAgMBAAEC
gYAQaYmO5yhrWOZQhMCoa1zH0FV7Pg/KNItkfd0z+LBtBorTX1/Y7aHeSiVfdRd8
A2rJTTXU4uZQVPhg5tiDzlkN0F703JvRNeu5oWFRRX4/jvmSjDn63pFRtmBa0ORm
5mHPHeos2Dz3FmKh4NxIYQ6gBAs6Hh9xIXcS42shmHUWcQJBANCfYUHosunqRAal
yjEuNBuPJAaP4C5i46hNAHWIq1i8ATpOvXWZytMiHYbYks3BsuwGvjEW863mJnUm
1TldXucCQQCEdu6MB2jQJXlW4XhdcPR2hZOfetrbznKNuO1tKep5cXXr0llrF8dW
Mr8wWBXbe9qQeVYdbB0VABOOifEjpOrPAkAX83RTAMgpmr+ck8QWyVsqHtDf//yY
1rmUROLcm4gwc8UgUJHnwnRKsQv6wzp3bNmBx3RmZmArgtS/dmncYB/ZAkBSiEuO
4ZrzfTXB5Q96oMMgCY/14LT2GQYUuTDtQB2AdynyuYfPCuy/DzVCKM/NhbijJYZ7
JH5mNDr7J4UgIUPPAkAuWiZ7X7R3P2Jo4v8EAV0wke7YQvYhHXY5f3NHMO8cL+zn
oyb1xYwxiys2aGSSNs+q8Mu6o0s21/knHS+pvKqy
-----END RSA PRIVATE KEY-----`,
}
