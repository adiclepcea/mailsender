package service

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/adiclepcea/mailsender"
)

//MailSenderService is the type describing the service to send mail
type MailSenderService struct {
	Mail  MailSetup `json:"mailsetup"`
	Setup Setup     `json:"servicesetup"`
}

//MailSetup represents the default setup for sending mail
type MailSetup struct {
	Server         string `json:"server"`
	UserMail       string `json:"usermail"`
	Password       string `json:"password"`
	UseInsecureTLS bool   `json:"insecuretls"`
	ServerCAFile   string `json:"mailservercafile"`
	UseTLS         bool   `json:"usetls"`
	UseAUTH        bool   `json:"useauth"`
}

//Setup respresents the setup for the service
type Setup struct {
	Port     int    `json:"port"`
	CertFile string `json:"certfile"`
	KeyFile  string `json:"keyfile"`
	CAFile   string `json:"cafile"`
}

//NewMailSenderService initiates MailSenderService struct from a json config file
func NewMailSenderService(configString string) (*MailSenderService, error) {

	var mss MailSenderService

	if err := json.NewDecoder(strings.NewReader(configString)).Decode(&mss); err != nil {
		return nil, err
	}

	if mss.Mail.Server == "" {
		return nil, fmt.Errorf("Invalid Mail setup")
	}

	if mss.Setup.Port == 0 {
		return nil, fmt.Errorf("Invalid Setup format")
	}

	return &mss, nil
}

//SendMail is the function that performs the actual sending of the mail
func (mss *MailSenderService) SendMail(msender mailsender.MailSender, ms mailsender.MailStruct) error {
	if mss.Mail.UseAUTH {
		if mss.Mail.UseTLS {
			var tlsconfig *tls.Config
			host, _, err := net.SplitHostPort(mss.Mail.Server)
			if err != nil {
				return err
			}
			if mss.Mail.UseInsecureTLS {
				log.Println("Insecure")
				//we should not check for certificate validity
				tlsconfig = mailsender.CreateInsecureTLSConfig(host)
			} else if mss.Mail.ServerCAFile != "" {
				log.Println("CA")
				//we have an own signed certificate
				if _, err = os.Stat(mss.Mail.ServerCAFile); os.IsNotExist(err) {
					return err
				}
				tlsconfig, err = mailsender.CreateTLSConfigWithCA(host, mss.Mail.ServerCAFile)
				if err != nil {
					return err
				}
			} else {
				log.Println("known")
				//we should use tls with a well known CA
				tlsconfig = mailsender.CreateTLSConfig(mss.Mail.Server)
			}
			_, err = msender.SendMailTLS(mss.Mail.Server, tlsconfig, mss.Mail.UserMail, mss.Mail.Password, ms)
			if err != nil {
				return err
			}
		} else {
			//we send mail with auth, but without TLS
			_, err := msender.SendMail(mss.Mail.Server, mss.Mail.UserMail, mss.Mail.Password, ms)
			if err != nil {
				return nil
			}
		}
	} else {
		//we send the mail without auth and without tls
		log.Println("aici")
		_, err := msender.SendMailWithoutAuth(mss.Mail.Server, ms)
		if err != nil {
			log.Printf("err: %v", err)
			return err
		}
		log.Println("muc")
	}
	return nil
}

//SendMailMessage is the method that links the REST call to the sendMail method
func (mss *MailSenderService) SendMailMessage(w http.ResponseWriter, r *http.Request) {

}

//Run starts the service with a REST Api
func (mss *MailSenderService) Run() error {
	http.HandleFunc("/sendmail", mss.SendMailMessage)

	return nil
}
