package service

import (
	"crypto/tls"
	"log"
	"net/mail"
	"testing"

	"github.com/adiclepcea/mailsender"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MyMailSender struct {
	mock.Mock
	mailsender.MailSender
}

func (m *MyMailSender) SendMailTLS(server string, tlsconfig *tls.Config, usermail string, pass string, ms mailsender.MailStruct) (int, error) {
	args := m.Called(server, tlsconfig, usermail, pass, ms)
	return args.Int(0), args.Error(1)
}

func (m *MyMailSender) SendMail(server string, usermail string, pass string, ms mailsender.MailStruct) (int, error) {
	args := m.Called(server, usermail, pass, ms)
	return args.Int(0), args.Error(1)
}

func (m *MyMailSender) SendMailWithoutAuth(server string, ms mailsender.MailStruct) (int, error) {
	log.Println("Called")
	args := m.Called(server, ms)
	log.Println("Returning")
	return args.Int(0), args.Error(1)
}

func TestServiceSendMail(t *testing.T) {
	//assert := assert.New(t)
	mailSetup := MailSetup{
		Server:          "exampleserver.com:654",
		DefaultMail:     "mail@exampleserver.com",
		DefaultPassword: "secret",
		UseTLS:          false,
		UseAUTH:         false,
		UseInsecureTLS:  true,
	}
	setup := Setup{
		Port: 8080,
	}

	serv := MailSenderService{Setup: setup, Mail: mailSetup}

	ms := mailsender.MailStruct{}
	ms.From = mail.Address{Name: "", Address: "src@server.com"}
	ms.To = mail.Address{Name: "", Address: "dest@server.com"}
	ms.Password = "secret"

	mockMailSender := new(MyMailSender)
	mockMailSender.On("SendMail", "exampleserver.com:654", "src@server.com", "secret", ms).Return(10, nil)
	mockMailSender.On("SendMailTLS", "exampleserver.com:654", mailsender.CreateInsecureTLSConfig("exampleserver.com"), "src@server.com", "secret", ms).Return(11, nil)
	mockMailSender.On("SendMailWithoutAuth", "exampleserver.com:654", ms).Return(12, nil)

	err := serv.SendMail(mockMailSender, ms)

	assert.Nil(t, err, "Err should be nil")
	mockMailSender.AssertCalled(t, "SendMailWithoutAuth", "exampleserver.com:654", ms)
	serv.Mail.UseAUTH = true
	serv.SendMail(mockMailSender, ms)
	mockMailSender.AssertCalled(t, "SendMail", "exampleserver.com:654", "src@server.com", "secret", ms)
	serv.Mail.UseTLS = true
	serv.SendMail(mockMailSender, ms)
	mockMailSender.AssertCalled(t, "SendMailTLS", "exampleserver.com:654", mailsender.CreateInsecureTLSConfig("exampleserver.com"), "src@server.com", "secret", ms)
	mockMailSender.AssertExpectations(t)
}

func TestCreateMailSenderServiceShouldFail(t *testing.T) {
	assert := assert.New(t)
	badJSON := `{"s"}`
	_, err := NewMailSenderService(badJSON)
	assert.NotNil(err, "Error expected while using an invalid json config, got nil\n")
	emptyJSON := `{}`
	_, err = NewMailSenderService(emptyJSON)
	assert.NotNil(err, "Error expected while using an empty json config, got nil\n")
}

func TestCreateMailSenderServiceShouldOK(t *testing.T) {
	assert := assert.New(t)
	goodJSON := `{
    "mailsetup":{
      "server":"exampleserver.com:645",
      "defaultmail":"mail@exampleserver.com",
      "defaultpassword":"secret",
      "insecuretls":false,
      "usetls":true,
      "useauth":true
    },
    "servicesetup":{
      "port":8080,
      "certfile":"",
      "keyfile":""
    }
  }`

	mss, err := NewMailSenderService(goodJSON)
	assert.Nil(err, "No error expected when creating from a valid JSON, got %v\n", err)

	assert.Equal("exampleserver.com:645", mss.Mail.Server, "Expected exampleserver.com:645, got %s\n", mss.Mail.Server)
	assert.Equal("mail@exampleserver.com", mss.Mail.DefaultMail, "Expected mail@exampleserver.com, got %s\n", mss.Mail.DefaultMail)
	assert.Equal("secret", mss.Mail.DefaultPassword, "Expected secret, got %s\n", mss.Mail.DefaultPassword)

	assert.Equal(8080, mss.Setup.Port, "Expected 8080, got %s\n", mss.Setup.Port)

}
