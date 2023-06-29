package mail

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/gomail.v2"
)

func SendMail(receiver string, movies []string) error {
	m := gomail.NewMessage()

	// Set E-Mail sender
	m.SetHeader("From", os.Getenv("EMAIL_SENDER"))

	// Set E-Mail receivers
	m.SetHeader("To", receiver)

	// Set E-Mail subject
	m.SetHeader("Subject", "New movies available to watch")

	// Set E-Mail body. You can set plain text or html with text/html
	m.SetBody("text/plain", "The following movies are available to download:\n"+strings.Join(movies, "\n"))

	// Settings for SMTP server
	port, err := strconv.Atoi(os.Getenv("EMAIL_PORT"))
	if err == nil {
		d := gomail.NewDialer(os.Getenv("EMAIL_SMTP"), port, os.Getenv("EMAIL_USER"), os.Getenv("EMAIL_PASSWORD"))

		// Now send E-Mail
		if err := d.DialAndSend(m); err != nil {
			fmt.Println(err)
			return err
		}
	} else {
		return err
	}

	return nil
}
