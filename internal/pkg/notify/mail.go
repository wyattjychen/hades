package notify

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/go-gomail/gomail"
	"github.com/wyattjychen/hades/internal/pkg/logger"
)

const (
	NotifyTypeMail = 1
)

var mailTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>
<title></title>
<meta charset="utf-8"/>

</head>
<body>
<div class="cap" style="
        height: 500px"
>
<div class="content" style="
        font-size: 13px;
        padding: 25px 25px;
        margin: 25px 25px;
    ">
    <div class="hello" style="text-align: center;font-size: 18px;font-weight: bolder">
        {{.Subject}}
    </div>
    <br>
    <div>
        <table cellspacing="0px" cellpadding="4px" style="margin: 0 auto;">
            <tr >
                <td>Node IP</td>
                <td >{{.IP}}</td>
            </tr>

            <tr>
                <td>Exception Time</td>
                <td>{{.OccurTime}}</td>
            </tr>

            <tr>
                <td>Exception Message</td>
                <td>{{.Body}}</td>
            </tr>

        </table>
    </div>
    <br><br>
</div>
</div>
<br>

</body>
</html>
`

var defaultMail *Mail

type Mail struct {
	Port     int
	From     string
	Host     string
	Secret   string
	Nickname string
}

func (mail *Mail) SendMsg(msg *Message) {
	m := gomail.NewMessage()

	m.SetHeader("From", m.FormatAddress(defaultMail.From, defaultMail.Nickname))
	m.SetHeader("To", msg.To...)
	m.SetHeader("Subject", msg.Subject)
	msgData := parseMailTemplate(msg)
	m.SetBody("text/html", msgData)

	d := gomail.NewDialer(defaultMail.Host, defaultMail.Port, defaultMail.From, defaultMail.Secret)
	if err := d.DialAndSend(m); err != nil {
		logger.GetLogger().Warn(fmt.Sprintf("smtp send msg[%+v] err: %s", msg, err.Error()))
	}
}

func parseMailTemplate(msg *Message) string {
	tmpl, err := template.New("notify").Parse(mailTemplate)
	if err != nil {
		return fmt.Sprintf("Failed to parse the notification template error: %s", err.Error())
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, msg)
	if err != nil {
		return fmt.Sprintf("Failed to parse the notification template execute error: %s", err.Error())
	}
	return buf.String()
}
