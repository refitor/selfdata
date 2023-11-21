package sdcore

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/smtp"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/gomail.v2"
)

var (
	v_email_from    = ""
	v_email_host    = ""
	v_email_passwd  = ""
)

func InitEmail(host, from, pwd string) {
	v_email_from = from
	v_email_host = host
	v_email_passwd = pwd
}

func SendEmailTLS(data []byte) error {
	dataMap := make(map[string]string, 0)
	if err := json.Unmarshal(data, &dataMap); err != nil {
		return err
	}
	if !IsValidEmail(dataMap["To"]) {
		return fmt.Errorf("invalid email: %s", dataMap["To"])
	}

	m := gomail.NewMessage()
	m.SetHeader("From", v_email_from)
	m.SetHeader("To", dataMap["To"])
	m.SetHeader("Subject", dataMap["Subject"])
	m.SetBody("text/plain", dataMap["Msg"])

	port, _ := strconv.Atoi(strings.Split(v_email_host, ":")[1])
	d := &gomail.Dialer{
		Port:      port,
		SSL:       true,
		Username:  v_email_from,
		Password:  v_email_passwd,
		Host:      strings.Split(v_email_host, ":")[0],
		TLSConfig: &tls.Config{InsecureSkipVerify: true},
		Auth:      smtp.PlainAuth(v_email_passwd, v_email_from, v_email_passwd, strings.Split(v_email_host, ":")[0]),
	}
	return d.DialAndSend(m)
}

func IsValidEmail(email string) bool {
	isValid, _ := regexp.MatchString("^([a-z0-9_\\.-]+)@([\\da-z\\.-]+)\\.([a-z\\.]{2,6})$", email)
	return isValid
}

