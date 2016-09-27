#Mailsender

###Description

Simple library to send mails from go.
It is intended to ease the sending of mails by adding anothe rlayer on top of go smtp library.
You can use it to send mails through an unencrypted connection or through an encrypted connection.
It also provides the posibility to use a self signed certificate for the mail server.

You can send mails using one of the three methods:

* ```SendMail``` - will allow you to send a mail using a connection without Tls security (unencrypted), but with authentication.
* ```SendMailWithoutAth``` - will allow you to send a mail using a connection without Tls security and without authentication. This is a setup that you should not have available if you are using a public mail server.
* ```SendMailTls``` - will allow you to send a mail using a secure connection (with Tls) and with authentication. This is the most secure option. Here you can have three variants:
 * Use a server that has a certificate signed by a known authority (this is the case for Google, Yahoo etc.). Before calling ```SendMailTls``` you will have to obtain the tls config using the ```CreateTlsConfig``` method
 * Use a server that has a self signed certificate or a certificate that is not known by your system. Before calling ```SendMailTls``` you will have to obtain the tls config using the ```CreateTlsConfigWithCA``` method. This method receives the name of the CA file that will be used to validate the server signature.
 * Use a server that provides a secure connection but you do not care to verify the connection (this means that you don't care for a MITM atack). In this case you should obtain the tls config using the ```CreateInsecureTlsConfig``` method.
