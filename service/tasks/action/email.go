package action

import (
	"fmt"
	"github.com/jordan-wright/email"
	"net/smtp"
)

func SendEmail() {

	e := email.NewEmail()
	e.From = "admin  <@gmail.com>"
	e.To = []string{"ap@nube-io.com"}
	e.Subject = "Awesome Subject"
	e.Text = []byte("Text Body is, of course, supported!")
	e.HTML = []byte("<h1>Fancy HTML is supported, too!</h1>")
	err := e.Send("smtp.gmail.com:587", smtp.PlainAuth("", "@gmail.com", "", "smtp.gmail.com"))
	fmt.Println(err)

}
