package service

import (
	"crypto/tls"
	"encoding/json"
	"net"
	"net/http"
	"os"

	"github.com/adiclepcea/mailsender"
)

//MailSenderService is the type describing the service to send mail
type MailSenderService struct {
	Server         string `json:"server"`
	UserMail       string `json:"usermail"`
	Password       string `json:"password"`
	UseInsecureTLS bool   `json:"insecuretls"`
	ServerCAFile   string `json:"mailservercafile"`
	UseTLS         bool   `json:"usetls"`
	UseAUTH        bool   `json:"useauth"`
	Port           int    `json:"port"`
	CertFile       string `json:"certfile"`
	KeyFile        string `json:"keyfile"`
	CAFile         string `json:"cafile"`
}

//NewMailSenderService initiates MailSenderService struct from a json config file
func NewMailSenderService(configFile string) (*MailSenderService, error) {

	var mss MailSenderService
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		//the cofig file does not exist
		return nil, err
	}

	file, err := os.Open(configFile)
	if err != nil {
		return nil, err
	}

	if err = json.NewDecoder(file).Decode(&mss); err != nil {
		return nil, err
	}

	return &mss, nil
}

//sendMail is the function that performs the actual sending of the mail
func (mss *MailSenderService) sendMail(ms mailsender.MailStruct) error {
	if mss.UseAUTH {
		if mss.UseTLS {
			var tlsconfig *tls.Config
			host, _, err := net.SplitHostPort(mss.Server)
			if err != nil {
				return err
			}
			if mss.UseInsecureTLS {
				//we should not check for certificate validity
				tlsconfig = mailsender.CreateInsecureTlsConfig(host)
			} else if mss.ServerCAFile != "" {
				//we have an own signed certificate
				if _, err = os.Stat(mss.ServerCAFile); os.IsNotExist(err) {
					return err
				}
				tlsconfig, err = mailsender.CreateTlsConfigWithCA(host, mss.ServerCAFile)
				if err != nil {
					return err
				}
			} else {
				//we should use tls with a wel known CA
				tlsconfig = mailsender.CreateTlsConfig(mss.Server)
			}
			_, err = mailsender.SendMailTls(mss.Server, tlsconfig, mss.UserMail, mss.Password, ms)
			if err != nil {
				return err
			}
		} else {
			//we send mail with auth, but without TLS
			_, err := mailsender.SendMail(mss.Server, mss.UserMail, mss.Password, ms)
			if err != nil {
				return nil
			}
		}
	} else {
		//we send the mail without auth and without tls
		_, err := mailsender.SendMailWithoutAuth(mss.Server, ms)
		if err != nil {
			return err
		}
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
