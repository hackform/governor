package mail

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"io"
	"mime/multipart"
	"mime/quotedprintable"
	"net/http"
	"net/smtp"
	"net/textproto"
	"xorkevin.dev/governor"
	"xorkevin.dev/governor/service/msgqueue"
	"xorkevin.dev/governor/service/template"
)

const (
	govmailchannelid = "gov.mail"
	govmailworker    = "gov.mail.worker"
)

type (
	// Mail is a service wrapper around a mailer instance
	Mail interface {
		Send(from, name, to, subjecttpl, bodytpl string, emdata interface{}) error
	}

	Service interface {
		governor.Service
		Mail
	}

	mailOp struct {
		from string
		to   []string
		msg  []byte
		res  chan<- error
	}

	mailmsg struct {
		From       string `json:"from"`
		FromName   string `json:"fromname"`
		To         string `json:"to"`
		Subjecttpl string `json:"subjecttpl"`
		Bodytpl    string `json:"bodytpl"`
		Emdata     string `json:"emdata"`
	}

	msgbuilder struct {
		headers  []string
		body     []byte
		htmlbody []byte
	}

	service struct {
		tpl         template.Template
		queue       msgqueue.Msgqueue
		config      governor.SecretReader
		logger      governor.Logger
		host        string
		addr        string
		fromAddress string
		fromName    string
		insecure    bool
		outbox      chan mailOp
		done        <-chan struct{}
	}
)

// New creates a new mailer service
func New(tpl template.Template, queue msgqueue.Msgqueue) Service {
	return &service{
		tpl:    tpl,
		queue:  queue,
		outbox: make(chan mailOp),
	}
}

func (s *service) Register(r governor.ConfigRegistrar, jr governor.JobRegistrar) {
	r.SetDefault("auth", "")
	r.SetDefault("host", "localhost")
	r.SetDefault("port", "587")
	r.SetDefault("fromaddress", "")
	r.SetDefault("fromname", "")
	r.SetDefault("insecure", false)
}

func (s *service) Init(ctx context.Context, c governor.Config, r governor.ConfigReader, l governor.Logger, g *echo.Group) error {
	s.logger = l
	l = s.logger.WithData(map[string]string{
		"phase": "init",
	})

	s.config = r

	conf := r.GetStrMap("")
	s.host = conf["host"]
	s.addr = fmt.Sprintf("%s:%s", conf["host"], conf["port"])
	s.fromAddress = conf["fromaddress"]
	s.fromName = conf["fromname"]
	s.insecure = r.GetBool("insecure")

	done := make(chan struct{})
	go s.execute(ctx, done)
	s.done = done

	l.Info("initialize mail options", map[string]string{
		"smtp server addr": s.addr,
		"sender address":   s.fromAddress,
		"sender name":      s.fromName,
	})
	return nil
}

func (s *service) Setup(req governor.ReqSetup) error {
	return nil
}

func (s *service) Start(ctx context.Context) error {
	return nil
}

func (s *service) Stop(ctx context.Context) {
	l := s.logger.WithData(map[string]string{
		"phase": "stop",
	})
	select {
	case <-s.done:
		return
	case <-ctx.Done():
		l.Warn("failed to stop", nil)
	}
}

func (s *service) Health() error {
	return nil
}

func (s *service) execute(ctx context.Context, done chan<- struct{}) {
	defer close(done)
	for {
		select {
		case <-ctx.Done():
			return
		case op := <-s.outbox:
			op.res <- s.handleSendMail(op.from, op.to, op.msg)
			close(op.res)
		}
	}
}

func (s *service) handleSendMail(from string, to []string, msg []byte) error {
	authsecret, err := s.config.GetSecret("auth")
	if err != nil {
		return err
	}

	smtpauth := smtp.PlainAuth("", authsecret["username"].(string), authsecret["password"].(string), s.host)
	if err := smtp.SendMail(s.addr, smtpauth, from, to, msg); err != nil {
		return err
	}
	return nil
}

func (s *service) mailSubscriber(msgdata []byte) error {
	emmsg := mailmsg{}
	if err := json.NewDecoder(bytes.NewBuffer(msgdata)).Decode(&emmsg); err != nil {
		return governor.NewError("Failed to decode mail message", http.StatusInternalServerError, err)
	}
	emdata := map[string]string{}
	if err := json.NewDecoder(bytes.NewBufferString(emmsg.Emdata)).Decode(&emdata); err != nil {
		return governor.NewError("Failed to decode mail data", http.StatusInternalServerError, err)
	}

	body, err := s.tpl.ExecuteHTML(emmsg.Bodytpl, emdata)
	if err != nil {
		return governor.NewError("Failed to execute mail body template", http.StatusInternalServerError, err)
	}
	subject, err := s.tpl.ExecuteHTML(emmsg.Subjecttpl, emdata)
	if err != nil {
		return governor.NewError("Failed to execute mail subject template", http.StatusInternalServerError, err)
	}

	msg := newMsgBuilder()
	if emmsg.FromName == "" {
		msg.addHeader("From", emmsg.From)
	} else {
		msg.addAddrHeader("From", emmsg.FromName, emmsg.From)
	}
	msg.addHeader("To", emmsg.To)
	msg.addHeader("Subject", string(subject))
	msg.addHtmlBody(body)
	buf, err := msg.build()
	if err != nil {
		return governor.NewError("Failed to write mail", http.StatusInternalServerError, err)
	}

	res := make(chan error)
	op := mailOp{
		from: emmsg.From,
		to:   []string{emmsg.To},
		msg:  buf.Bytes(),
		res:  res,
	}
	select {
	case <-s.done:
		return governor.NewError("Mail service shutdown", http.StatusInternalServerError, err)
	case s.outbox <- op:
		return <-res
	}
}

func newMsgBuilder() *msgbuilder {
	return &msgbuilder{
		headers: []string{},
		body:    nil,
	}
}

func (b *msgbuilder) addHeader(key, val string) {
	b.headers = append(b.headers, fmt.Sprintf("%s: %s", key, val))
}

func (b *msgbuilder) addAddrHeader(key, name, addr string) {
	b.headers = append(b.headers, fmt.Sprintf("%s: %s <%s>", key, name, addr))
}

func (b *msgbuilder) addBody(body []byte) {
	b.body = body
}

func (b *msgbuilder) addHtmlBody(body []byte) {
	b.htmlbody = body
}

func (b *msgbuilder) writeHeaders(buf *bytes.Buffer) {
	for _, h := range b.headers {
		buf.WriteString(h)
		buf.WriteString("\r\n")
	}
}

func (b *msgbuilder) writePart(w io.Writer, data []byte) error {
	qw := quotedprintable.NewWriter(w)
	defer qw.Close()
	if _, err := qw.Write(data); err != nil {
		return err
	}
	return nil
}

func (b *msgbuilder) writeBody(w *multipart.Writer) error {
	defer w.Close()
	if len(b.body) != 0 {
		w, err := w.CreatePart(textproto.MIMEHeader{"Content-Type": {"text/plain; charset=UTF-8"}, "Content-Transfer-Encoding": {"quoted-printable"}})
		if err != nil {
			return err
		}
		if err := b.writePart(w, b.body); err != nil {
			return err
		}
	}
	if len(b.htmlbody) != 0 {
		w, err := w.CreatePart(textproto.MIMEHeader{"Content-Type": {"text/html; charset=UTF-8"}, "Content-Transfer-Encoding": {"quoted-printable"}})
		if err != nil {
			return err
		}
		if err := b.writePart(w, b.htmlbody); err != nil {
			return err
		}
	}
	return nil
}

func genBoundary() string {
	return multipart.NewWriter(&bytes.Buffer{}).Boundary()
}

func createPart(m *multipart.Writer, contenttype string) (*multipart.Writer, error) {
	boundary := genBoundary()
	part, err := m.CreatePart(textproto.MIMEHeader{"Content-Type": {fmt.Sprintf("%s; boundary=\"%s\"", contenttype, boundary)}})
	if err != nil {
		return nil, err
	}
	w := multipart.NewWriter(part)
	if err := w.SetBoundary(boundary); err != nil {
		return nil, err
	}
	return w, nil
}

func (b *msgbuilder) build() (*bytes.Buffer, error) {
	buf := &bytes.Buffer{}
	b.writeHeaders(buf)

	buf.WriteString("Mime-Version: 1.0\r\n")
	m := multipart.NewWriter(buf)
	defer m.Close()
	fmt.Fprintf(buf, "Content-Type: multipart/mixed; boundary=\"%s\"; charset=UTF-8\r\n\r\n", m.Boundary())

	part, err := createPart(m, "multipart/alternative")
	if err != nil {
		return nil, err
	}
	if err := b.writeBody(part); err != nil {
		return nil, err
	}

	return buf, nil
}

// Send creates and enqueues a new message to be sent
func (s *service) Send(from, fromname, to, subjecttpl, bodytpl string, emdata interface{}) error {
	datastring := &bytes.Buffer{}
	if err := json.NewEncoder(datastring).Encode(emdata); err != nil {
		return governor.NewError("Failed to encode email data to JSON", http.StatusInternalServerError, err)
	}

	msg := mailmsg{
		From:       from,
		FromName:   fromname,
		To:         to,
		Subjecttpl: subjecttpl,
		Bodytpl:    bodytpl,
		Emdata:     datastring.String(),
	}
	if msg.From == "" {
		msg.From = s.fromAddress
	}
	if msg.FromName == "" {
		msg.FromName = s.fromName
	}

	b := &bytes.Buffer{}
	if err := json.NewEncoder(b).Encode(msg); err != nil {
		return governor.NewError("Failed to encode email to json", http.StatusInternalServerError, err)
	}
	if err := s.queue.Publish(govmailchannelid, b.Bytes()); err != nil {
		return governor.NewError("Failed to publish new email to message queue", http.StatusInternalServerError, err)
	}
	return nil
}
