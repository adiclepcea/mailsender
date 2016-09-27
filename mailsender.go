package mailsender

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"net/mail"
	"net/smtp"
)

//MailStruct holds the basic mail fields
type MailStruct struct {
	From    mail.Address
	To      mail.Address
	Subject string
	Body    string
}

func messageFromMailStruct(ms MailStruct) []byte {
	headers := make(map[string]string)
	headers["From"] = ms.From.String()
	headers["To"] = ms.To.String()
	headers["Subject"] = ms.Subject

	var message string

	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + ms.Body

	return []byte(message)

}

//CreateTlsConfigWithCA will create a tls configuration that will
//check for the validity of the server certificate against a specified CA
//This is most likely to happen when you own sign your certificates.
func CreateTlsConfigWithCA(serverName string, caFile string) (*tls.Config, error) {

	caCert, err := ioutil.ReadFile(caFile)
	if err != nil {
		return nil, err
	}

	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM(caCert)
	if !ok {
		return nil, fmt.Errorf("Failed to parse CA")
	}

	return &tls.Config{
		RootCAs:    roots,
		ServerName: serverName,
	}, nil
}

//createInsecureTlsConfig will create a tls configuration that does
//not check for the validity of the server certificate
func CreateInsecureTlsConfig(serverName string) *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         serverName,
	}
}

//createTlsConfig will create a tls configuration that will
//check for tha validity of the server certificate against
//a known authority signed CA
func CreateTlsConfig(serverName string) *tls.Config {
	return &tls.Config{
		ServerName: serverName,
	}
}

//SendMailTls send the mail after the tls params have been set
func SendMailTls(server string, tlsconfig *tls.Config, usermail string, pass string, ms MailStruct) (int, error) {
	host, _, err := net.SplitHostPort(server)
	if err != nil {
		return 0, err
	}

	auth := smtp.PlainAuth("", usermail, pass, host)

	conn, err := tls.Dial("tcp", server, tlsconfig)

	if err != nil {
		return 0, err
	}

	defer conn.Close()

	return sendMailWithConnAndAuth(conn, host, ms, auth)

}

//SendMail will send a mail using authentication without encryption
func SendMail(server string, usermail string, pass string, ms MailStruct) (int, error) {

	host, _, err := net.SplitHostPort(server)
	if err != nil {
		return 0, err
	}

	auth := smtp.PlainAuth("", usermail, pass, host)

	conn, err := net.Dial("tcp", server)

	if err != nil {
		return 0, nil
	}

	return sendMailWithConnAndAuth(conn, host, ms, auth)
}

//SendMailWithoutAuth sends a mail without using authentication
func SendMailWithoutAuth(server string, ms MailStruct) (int, error) {

	host, _, err := net.SplitHostPort(server)
	if err != nil {
		return 0, err
	}
	conn, err := net.Dial("tcp", server)
	if err != nil {
		return 0, nil
	}

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return 0, err
	}

	return sendMailWithClient(client, ms)
}

//SendMailWithClient sends the mail after the client has been set up
func sendMailWithClient(client *smtp.Client, ms MailStruct) (int, error) {

	defer client.Quit()

	if err := client.Mail(ms.From.Address); err != nil {
		return 0, err
	}

	if err := client.Rcpt(ms.To.Address); err != nil {
		return 0, err
	}

	writer, err := client.Data()

	if err != nil {
		return 0, err
	}

	defer writer.Close()

	n, err := writer.Write(messageFromMailStruct(ms))

	if err != nil {
		return 0, err
	}

	return n, nil

}

//SendMailWithConnAndAuth sends the mail using a setup conn and an authentication
func sendMailWithConnAndAuth(conn net.Conn, host string, ms MailStruct, auth smtp.Auth) (int, error) {
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return 0, err
	}

	if err = client.Auth(auth); err != nil {
		return 0, err
	}

	return sendMailWithClient(client, ms)
}
