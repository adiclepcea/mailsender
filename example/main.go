package main

import (
	"fmt"
	"github.com/adiclepcea/mailsender"
	"log"
	"net/mail"
)

func main() {

	ms := mailsender.MailStruct{
		From:    mail.Address{"User Name", "user@server.com"},
		To:      mail.Address{"Destination", "destination@destinationserver.com"},
		Subject: "Your subject",
		Body:    "Just a test mail\nOn two lines",
	}

	servername := "mail.server.com:587"

	/*
		//This is if you want to use TLS for sending the mail
		//also import "net"
		host, _, err := net.SplitHostPort(servername)

		if err != nil {
			log.Printf("Error reading host: %s\n", err.Error())
		}

		tlsconfig := mailsender.CreateTlsConfig(host)
		n, err := mailsender.SendMailTls(servername, tlsconfig, "user@server.com", "password", ms)
	*/

	n, err := mailsender.SendMail(servername, "user@server.com", "password", ms)

	if err != nil {
		log.Println(err)
	}

	fmt.Printf("Sent %d bytes\n", n)

}
