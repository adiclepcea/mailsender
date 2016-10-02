package service

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/mail"
	"os"
	"regexp"
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
	Server          string `json:"server"`
	DefaultMail     string `json:"defaultmail"`
	DefaultPassword string `json:"defaultpassword"`
	UseInsecureTLS  bool   `json:"insecuretls"`
	ServerCAFile    string `json:"mailservercafile"`
	UseTLS          bool   `json:"usetls"`
	UseAUTH         bool   `json:"useauth"`
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
				tlsconfig = mailsender.CreateTLSConfig(host)
			}
			log.Printf("%s, user=%s, password=%s", mss.Mail.Server, ms.From.Address, ms.Password)
			_, err = msender.SendMailTLS(mss.Mail.Server, tlsconfig, ms.From.Address, ms.Password, ms)
			if err != nil {
				return err
			}
		} else {
			//we send mail with auth, but without TLS
			_, err := msender.SendMail(mss.Mail.Server, ms.From.Address, ms.Password, ms)
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

func validateEmail(email string) bool {
	Re := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return Re.MatchString(email)
}

//ValidateMailStruct will validate the mail struct received as json against the rules
func (mss *MailSenderService) ValidateMailStruct(ms *mailsender.MailStruct) (
	*mailsender.MailStruct, error) {

	if ms.To.Address == "" {
		return nil, fmt.Errorf("No destination address provided")
	} else if !validateEmail(ms.To.Address) {
		return nil, fmt.Errorf("%s is not a valid destination address", ms.To.Address)
	}
	if ms.From.Address == "" {
		ms.From = mail.Address{Name: ms.From.Name, Address: mss.Mail.DefaultMail}
		ms.Password = mss.Mail.DefaultPassword
	} else if !validateEmail(ms.From.Address) {
		return nil, fmt.Errorf("%s is not a valid mail address", ms.From.String())
	} else if ms.Password == "" {
		return nil, fmt.Errorf("No password provided for this address")
	}

	return ms, nil

}

//SendMailMessage is the method that links the REST call to the sendMail method
func (mss *MailSenderService) SendMailMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Sorry, only POST allowed!"))
		return
	}
	decoder := json.NewDecoder(r.Body)

	var ms mailsender.MailStruct

	err := decoder.Decode(&ms)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	_, err = mss.ValidateMailStruct(&ms)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	msender := &mailsender.Impl{}
	err = mss.SendMail(msender, ms)
	log.Println(ms.From)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write([]byte("OK"))
}

func fileCanBeUsed(fileName string) (bool, error) {
	var err error
	if _, err = os.Stat("config.json"); err != nil {
		return false, err
	}
	return true, nil
}

func loadCA(fileName string) ([]byte, error) {
	if exists, err := fileCanBeUsed(fileName); !exists {
		return nil, err
	}
	caCert, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	return caCert, nil
}

//Run starts the service with a REST Api
func (mss *MailSenderService) Run() error {
	var caCert []byte
	var caCertPool *x509.CertPool
	var err error
	var server *http.Server
	var tlsConfig *tls.Config

	http.HandleFunc("/sendmail", mss.SendMailMessage)

	//load the CAFile to authenticate the clients if needed
	if mss.Setup.CAFile != "" {
		caCert, err = loadCA(mss.Setup.CAFile)
		if err != nil {
			log.Fatalf("CA file %s can't be used: %s", mss.Setup.CAFile, err.Error())
		}
		caCertPool = x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
	}

	//load the CertFile and  KeyFile to use TLS if needed
	if mss.Setup.CertFile != "" {
		if ok, err := fileCanBeUsed(mss.Setup.CertFile); !ok {
			log.Fatalf("CertFile %s can't be used: %s", mss.Setup.CAFile, err)
		}
		if ok, err := fileCanBeUsed(mss.Setup.KeyFile); !ok {
			log.Fatalf("KeyFile %s can't be used: %s", mss.Setup.KeyFile, err)
		}
		if caCert != nil {
			tlsConfig = &tls.Config{
				ClientCAs:  caCertPool,
				ClientAuth: tls.RequireAndVerifyClientCert,
			}
		} else {
			tlsConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
		}
		tlsConfig.BuildNameToCertificate()
		server = &http.Server{
			TLSConfig: tlsConfig,
			Addr:      fmt.Sprintf(":%d", mss.Setup.Port),
		}
		server.ListenAndServeTLS(mss.Setup.CertFile, mss.Setup.KeyFile)
	} else {
		//start the server without tls
		server = &http.Server{
			Addr: fmt.Sprintf(":%d", mss.Setup.Port),
		}
		server.ListenAndServe()
	}

	return nil
}
