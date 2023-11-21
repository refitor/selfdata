package sdwork

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"github.com/dgryski/dgoogauth"
	"github.com/manifoldco/promptui"
)

func AuthByGoogle(secret, code string) (string, error) {
	if code == "" {
		var promptCode promptui.Prompt
		promptCode = promptui.Prompt{
			Label:    "Authorization code",
			Validate: nil,
		}
		code, _ = promptCode.Run()
	}

	if ok, err := NewGoogleAuth().VerifyCode(secret, code); !ok {
		return "", fmt.Errorf("invalid code: %+v", err)
	}
	return code, nil
}

// Google Authenticator
type GoogleAuth struct {
}

func NewGoogleAuth() *GoogleAuth {
	return &GoogleAuth{}
}

func (this *GoogleAuth) un() int64 {
	return time.Now().UnixNano() / 1000 / 30
}

func (this *GoogleAuth) hmacSha1(key, data []byte) []byte {
	h := hmac.New(sha1.New, key)
	if total := len(data); total > 0 {
		h.Write(data)
	}
	return h.Sum(nil)
}

func (this *GoogleAuth) base32encode(src []byte) string {
	return base32.StdEncoding.EncodeToString(src)
}

func (this *GoogleAuth) GetSecret() string {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, this.un())
	return strings.ToUpper(this.base32encode(this.hmacSha1(buf.Bytes(), nil)))
}

func (this *GoogleAuth) GetQrcode(user, secret string) string {
	return fmt.Sprintf("otpauth://totp/selfdata:%s?secret=%s", user, secret)
}

func (this *GoogleAuth) VerifyCode(secret, code string) (bool, error) {
	otpConfig := &dgoogauth.OTPConfig{
		Secret:      strings.TrimSpace(secret),
		WindowSize:  3,
		HotpCounter: 0,
	}
	return otpConfig.Authenticate(strings.TrimSpace(code))
}
