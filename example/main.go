package main

import (
	"fmt"
	"log"
	"net/mail"

	"github.com/adiclepcea/mailsender"
)

func main() {

	ms := mailsender.MailStruct{
		From:    mail.Address{Name: "User Name", Address: "user@server.com"},
		To:      mail.Address{Name: "Destination", Address: "destination@destinationserver.com"},
		Subject: "Your subject",
		Body:    "Just a test mail\nOn two lines",
	}

	servername := "mail.server.com:587"
	impl := mailsender.Impl{}

	/*
		//This is if you want to use TLS for sending the mail
		//also import "net"
		host, _, err := net.SplitHostPort(servername)

		if err != nil {
			log.Printf("Error reading host: %s\n", err.Error())
		}

		tlsconfig := mailsender.CreateTlsConfig(host)
		n, err := impl.SendMailTls(servername, tlsconfig, "user@server.com", "password", ms)
	*/

	n, err := impl.SendMail(servername, "user@server.com", "password", ms)

	if err != nil {
		log.Println(err)
	}

	fmt.Printf("Sent %d bytes\n", n)

}
