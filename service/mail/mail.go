package mail

import (
	"bytes"
	"crypto/tls"
	"encoding/gob"
	"encoding/json"
	gomail "github.com/go-mail/mail"
	"github.com/hackform/governor"
	"github.com/hackform/governor/service/msgqueue"
	"github.com/hackform/governor/service/template"
	"github.com/labstack/echo"
	"net/http"
	"strconv"
	"time"
)

const (
	govmailqueueid     = "gov-mail"
	govmailqueueworker = "gov-mail-worker"
)

type (
	// Mail is a service wrapper around a mailer instance
	Mail interface {
		governor.Service
		Send(to, subjecttpl, bodytpl string, emdata interface{}) *governor.Error
	}

	goMail struct {
		logger      governor.Logger
		tpl         template.Template
		queue       msgqueue.Msgqueue
		host        string
		port        int
		username    string
		password    string
		insecure    bool
		bufferSize  int
		workerSize  int
		connMsgCap  int
		fromAddress string
		fromName    string
		msgc        chan *gomail.Message
	}

	mailmsg struct {
		To         string
		Subjecttpl string
		Bodytpl    string
		Emdata     string
	}
)

const (
	moduleID = "mail"
)

// New creates a new mailer service
func New(c governor.Config, l governor.Logger, tpl template.Template, queue msgqueue.Msgqueue) (Mail, error) {
	v := c.Conf()
	rconf := v.GetStringMapString("mail")

	l.Info("initialized mail service", moduleID, "initialize mail service", 0, map[string]string{
		"smtp server host": rconf["host"],
		"smtp server port": rconf["port"],
		"buffer_size":      strconv.Itoa(v.GetInt("mail.buffer_size")),
		"worker_size":      strconv.Itoa(v.GetInt("mail.worker_size")),
		"conn_msg_cap":     strconv.Itoa(v.GetInt("mail.conn_msg_cap")),
		"sender name":      rconf["from_name"],
		"sender address":   rconf["from_address"],
	})

	gm := &goMail{
		logger:      l,
		tpl:         tpl,
		queue:       queue,
		host:        rconf["host"],
		port:        v.GetInt("mail.port"),
		username:    rconf["username"],
		password:    rconf["password"],
		insecure:    v.GetBool("mail.insecure"),
		bufferSize:  v.GetInt("mail.buffer_size"),
		workerSize:  v.GetInt("mail.worker_size"),
		connMsgCap:  v.GetInt("mail.conn_msg_cap"),
		fromAddress: rconf["from_address"],
		fromName:    rconf["from_name"],
		msgc:        make(chan *gomail.Message, v.GetInt("mail.buffer_size")),
	}

	if err := gm.startWorkers(); err != nil {
		return nil, err
	}

	return gm, nil
}

func (m *goMail) dialer() *gomail.Dialer {
	d := gomail.NewDialer(m.host, m.port, m.username, m.password)

	if m.insecure {
		d.TLSConfig = &tls.Config{
			ServerName:         m.host,
			InsecureSkipVerify: true,
		}
	}

	return d
}

const (
	moduleIDmailWorker = moduleID + ".mailWorker"
)

func (m *goMail) mailWorker() {
	cap := m.connMsgCap
	d := m.dialer()
	var sender gomail.SendCloser
	mailSent := 0

	for {
		select {
		case msg, ok := <-m.msgc:
			if !ok {
				return
			}
			if sender == nil || mailSent >= cap && cap > 0 {
				if s, err := d.Dial(); err == nil {
					sender = s
					mailSent = 0
				} else {
					m.logger.Error(err.Error(), moduleIDmailWorker, "fail dial smtp server", 0, nil)
				}
			}
			if sender != nil {
				if err := gomail.Send(sender, msg); err != nil {
					m.logger.Error(err.Error(), moduleIDmailWorker, "fail send smtp server", 0, nil)
				}
				mailSent++
			}

		case <-time.After(30 * time.Second):
			if sender != nil {
				if err := sender.Close(); err != nil {
					m.logger.Error(err.Error(), moduleIDmailWorker, "fail close smtp client", 0, nil)
				}
				sender = nil
			}
		}
	}
}

const (
	moduleIDmailSubscriber = moduleID + ".mailSubscriber"
)

func (m *goMail) mailSubscriber(msgdata []byte) {
	emmsg := mailmsg{}
	b := bytes.NewBuffer(msgdata)
	if err := gob.NewDecoder(b).Decode(&emmsg); err != nil {
		m.logger.Error(err.Error(), moduleIDmailSubscriber, "fail decode mailmsg", 0, nil)
		return
	}

	emdata := map[string]string{}
	b1 := bytes.NewBufferString(emmsg.Emdata)
	if err := json.NewDecoder(b1).Decode(&emdata); err != nil {
		m.logger.Error(err.Error(), moduleIDmailSubscriber, "fail decode emdata", 0, nil)
		return
	}
	if err := m.enqueue(emmsg.To, emmsg.Subjecttpl, emmsg.Bodytpl, emdata); err != nil {
		m.logger.Error(err.Error(), moduleIDmailSubscriber, "fail enqueue mail", 0, nil)
		return
	}
}

const (
	moduleIDstartWorkers = moduleID + ".startWorkers"
)

func (m *goMail) startWorkers() *governor.Error {
	for i := 0; i < m.workerSize; i++ {
		go m.mailWorker()
	}
	if _, err := m.queue.SubscribeQueue(govmailqueueid, govmailqueueworker, m.mailSubscriber); err != nil {
		err.AddTrace(moduleIDstartWorkers)
		return err
	}
	return nil
}

const (
	moduleIDenqueue = moduleID + ".enqueue"
)

func (m *goMail) enqueue(to, subjecttpl, bodytpl string, emdata interface{}) *governor.Error {
	body, err := m.tpl.ExecuteHTML(bodytpl, emdata)
	if err != nil {
		err.AddTrace(moduleIDenqueue)
		return err
	}
	subject, err := m.tpl.ExecuteHTML(subjecttpl, emdata)
	if err != nil {
		err.AddTrace(moduleIDenqueue)
		return err
	}

	msg := gomail.NewMessage()
	if len(m.fromName) > 0 {
		msg.SetAddressHeader("From", m.fromAddress, m.fromName)
	} else {
		msg.SetHeader("From", m.fromAddress)
	}
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", body)

	select {
	case m.msgc <- msg:
		return nil
	case <-time.After(30 * time.Second):
		return governor.NewError(moduleIDenqueue, "email service experiencing load", 0, http.StatusInternalServerError)
	}
}

// Mount is a place to mount routes to satisfy the Service interface
func (m *goMail) Mount(conf governor.Config, l governor.Logger, r *echo.Group) error {
	l.Info("mounted mail service", moduleID, "mount mail service", 0, nil)
	return nil
}

// Health is a health check for the service
func (m *goMail) Health() *governor.Error {
	return nil
}

// Setup is run on service setup
func (m *goMail) Setup(conf governor.Config, l governor.Logger, rsetup governor.ReqSetupPost) *governor.Error {
	return nil
}

const (
	moduleIDSend = moduleID + ".Send"
)

// Send creates and enqueues a new message to be sent
func (m *goMail) Send(to, subjecttpl, bodytpl string, emdata interface{}) *governor.Error {
	datastring := bytes.Buffer{}
	if err := json.NewEncoder(&datastring).Encode(emdata); err != nil {
		return governor.NewError(moduleIDSend, err.Error(), 0, http.StatusInternalServerError)
	}

	b := bytes.Buffer{}
	if err := gob.NewEncoder(&b).Encode(mailmsg{
		To:         to,
		Subjecttpl: subjecttpl,
		Bodytpl:    bodytpl,
		Emdata:     datastring.String(),
	}); err != nil {
		return governor.NewError(moduleIDSend, err.Error(), 0, http.StatusInternalServerError)
	}
	if err := m.queue.Publish(govmailqueueid, b.Bytes()); err != nil {
		err.AddTrace(moduleIDSend)
		return err
	}
	return nil
}
