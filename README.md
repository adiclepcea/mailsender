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

### Simple service implementation

 In example_service folder you find a possible implementation for a service that will listen for a REST request that can send mails.
 For example, you could have a ```config.json``` file near your executable that could have a content like:

 ```
 {
   "mailsetup":{
     "server":"mail.yourmailserver.net:465",
     "defaultmail":"admin@yourmailserver.net",
     "defaultpassword":"verysecretpass",
     "insecuretls":false,
     "usetls":true,
     "useauth":true
   },
   "servicesetup":{
     "port":8080,
     "certfile":"",
     "keyfile":"",
     "cafile": ""
   }
 }
```

This tells the program to use the ```mail.yourmailserver.net``` mail server, to connect to the port ```465```, to use the ```admin@yourmailserver.net``` as the default mail address for authentication, and the ```verysecretpass``` password. This configuration would use a secure connection to the mail server to send mails.

The service would listen for requests on port 8080.

A possible request to send a mail using the service running on the localhost and curl is:

```
curl -X POST http://localhost:8080/sendmail -d '{"To":{"Name":"","Address":"user@somemailserver.com"},"Subject":"test","Body":"From service"}'
```

This would send a mail comming from ```admin@yourmailserver.net``` with the subject and body specified in the curl request.

```
curl -X POST http://localhost:8080/sendmail -d '{"To":{"Name":"","Address":"user@somemailserver.com"},"Subject":"test","Body":"From service", "From":{"Name":"","Address":"user@yourmailserver.net"},"Password":"otherpass"}'
```
This would send a mail comming from ```user@yourmailserver.net``` with the subject and body specified in the curl request as long as the password of this user is valid for your mailserver.

To make the service use a secure connection (https), you should provide a key and a cert file.

To restrict access only to the clients with a valid certificate, use the corresponding CA file in the configuration file.
