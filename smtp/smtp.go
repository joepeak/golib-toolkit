package smtp

import (
	"errors"
	"math/rand"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gopkg.in/gomail.v2"
)

var (
	clients []*Client

	config Config
)

const (
	confKey = "smtp"
)

type Client struct {
	dialer *gomail.Dialer
	sender *Sender
}

type Config struct {
	Server  string    `mapstructure:"server"`
	Port    int       `mapstructure:"port"`
	Senders []*Sender `mapstructure:"senders"`
}

type Sender struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Email    string `mapstructure:"email"`
}

func init() {

	if !viper.IsSet(confKey) {
		logrus.WithFields(logrus.Fields{
			"confKey": confKey,
		}).Warn("SMTP configuration not found")

		return
	}

	if err := viper.UnmarshalKey(confKey, &config); err != nil {
		logrus.WithFields(logrus.Fields{
			"confiKey": confKey,
			"error":    err.Error(),
		}).Error("Init smtp error")
		return
	}

	if len(config.Senders) == 0 {
		logrus.WithFields(logrus.Fields{
			"confKey": confKey,
		}).Warn("SMTP senders not found")

		return
	}

	clients = make([]*Client, len(config.Senders))
	for i, sender := range config.Senders {
		clients[i] = &Client{
			dialer: gomail.NewDialer(config.Server, config.Port, sender.Username, sender.Password),
			sender: sender,
		}
	}
}

// Send sends an email using the random dialer.
func Send(toEmail, subject, content string) error {
	m := gomail.NewMessage()
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", content)

	return send(m)
}

// SendWithAttachments sends an email with multiple attachments using the random dialer.
func SendWithAttachments(toEmail, subject, content string, attachmentPaths []string) error {
	m := gomail.NewMessage()
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", content)

	for _, attachmentPath := range attachmentPaths {
		if attachmentPath != "" {
			m.Attach(attachmentPath)
		}
	}

	return send(m)
}

// get random client
func getRandomClient() *Client {
	if len(clients) == 0 {
		return nil
	}

	if len(clients) == 1 {
		return clients[0]
	}

	return clients[rand.Intn(len(clients))]
}

// send sends an email using the random client.
func send(m *gomail.Message) error {
	if m == nil {
		return errors.New("message is nil")
	}

	client := getRandomClient()
	if client == nil {
		return errors.New("no client available")
	}

	// set sender email
	m.SetHeader("From", client.sender.Email)

	return client.dialer.DialAndSend(m)
}
