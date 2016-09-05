package mailgun

import (
	"fmt"
	"log"

	"time"

	"gopkg.in/mailgun/mailgun-go.v1"
)

func Newpochtalion(domain, apiKey, publicApiKey string) *Pochtalion {
	return &Pochtalion{
		mailgun: mailgun.NewMailgun(domain, apiKey, publicApiKey),
	}
}

type Pochtalion struct {
	mailgun mailgun.Mailgun
}

func (p *Pochtalion) Send(from, title, to, body string) chan error {
	done := make(chan error, 1)
	go func() {
		message := p.mailgun.NewMessage(from, title, body, to)
		message.SetTracking(true)
		message.SetReplyTo(from)
		msg, id, err := p.mailgun.Send(message)

		log.Println("[SEND]", "[ID]", id, "msg", msg, "[ERR]", err)
		done <- err
	}()

	return done
}

func (p *Pochtalion) SendMailling(from, title, body string, emails ...string) chan error {
	done := make(chan error, 1)
	go func() {
		var countErr int
		var err error

		for _, to := range emails {
			select {
			case err := <-p.Send(from, title, to, body):
				if err != nil {
					countErr++
				}
			case <-time.After(time.Second * 3):
				log.Println("[SEND]", "[ERR]", "sending timeout")
				countErr++
			}
		}

		if countErr > 0 {
			err = fmt.Errorf("[WARNING] not sending %d emails, see log", countErr)
		}
		done <- err

	}()

	return done
}
