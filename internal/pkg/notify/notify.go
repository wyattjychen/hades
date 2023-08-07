package notify

import (
	"strings"
	"time"

	"github.com/wyattjychen/hades/internal/pkg/utils"
)

type Noticer interface {
	SendMsg(*Message)
}

var msgQueue chan *Message

func Init(mail *Mail) {
	defaultMail = &Mail{
		Port:     mail.Port,
		From:     mail.From,
		Host:     mail.Host,
		Secret:   mail.Secret,
		Nickname: mail.Nickname,
	}
	msgQueue = make(chan *Message, 64)
}

type Message struct {
	Type      int
	IP        string
	Subject   string
	Body      string
	To        []string
	OccurTime string
}

func Send(msg *Message) {
	msgQueue <- msg
}

func Serve() {
	for {
		select {
		case msg := <-msgQueue:
			if msg == nil {

			}
			switch msg.Type {
			case 1:
				//Mail
				msg.Check()
				defaultMail.SendMsg(msg)
			}
		}
	}
}

func (m *Message) Check() {
	if m.OccurTime == "" {
		m.OccurTime = time.Now().Format(utils.TimeFormatSecond)
	}
	m.Body = strings.Replace(m.Body, "\"", "'", -1)
	m.Body = strings.Replace(m.Body, "\n", "", -1)
}
