package services

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"mime/multipart"
	"mime/quotedprintable"
	"net/smtp"
	"net/textproto"

	"github.com/DIMO-Network/trips-api/internal/config"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

//go:embed email_template.html
var rawDataDownloadEmail string

type EmailService struct {
	emailTemplate *template.Template
	ClientConn    *grpc.ClientConn
	username      string
	pw            string
	host          string
	port          string
	emailFrom     string
	log           *zerolog.Logger
	usersGRPCAddr string
}

func NewEmailService(settings *config.Settings, log *zerolog.Logger) (*EmailService, error) {
	t := template.Must(template.New("email_template").Parse(rawDataDownloadEmail))
	conn, err := grpc.Dial(settings.UsersAPIGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &EmailService{emailTemplate: t,
		ClientConn:    conn,
		username:      settings.EmailUsername,
		pw:            settings.EmailPassword,
		host:          settings.EmailHost,
		port:          settings.EmailPort,
		emailFrom:     settings.EmailFrom,
		log:           log,
		usersGRPCAddr: settings.UsersAPIGRPCAddr}, nil
}

func (es *EmailService) SendEmail(user string, summary string) error {

	userEmail, err := es.getVerifiedEmailAddress(user)
	if err != nil {
		return err
	}

	auth := smtp.PlainAuth("", es.username, es.pw, es.host)
	addr := fmt.Sprintf("%s:%s", es.host, es.port)

	var partsBuffer bytes.Buffer
	w := multipart.NewWriter(&partsBuffer)
	defer w.Close() //nolint

	p, err := w.CreatePart(textproto.MIMEHeader{"Content-Type": {"text/plain"}, "Content-Transfer-Encoding": {"quoted-printable"}})
	if err != nil {
		return err
	}

	ptMessage := fmt.Sprintf("Hi,\r\n\r\nSee below for request summary data for your trip:\n%s\n\n", summary)

	pw := quotedprintable.NewWriter(p)
	if _, err := pw.Write([]byte(ptMessage)); err != nil {
		return err
	}
	pw.Close()
	h, err := w.CreatePart(textproto.MIMEHeader{"Content-Type": {"text/html"}, "Content-Transfer-Encoding": {"quoted-printable"}})
	if err != nil {
		return err
	}

	hw := quotedprintable.NewWriter(h)
	var htmlMessage string

	if err := es.emailTemplate.Execute(hw, template.HTML(htmlMessage)); err != nil {
		return err
	}
	hw.Close()
	var buffer bytes.Buffer
	buffer.WriteString("From: DIMO <" + es.emailFrom + ">\r\n" +
		"To: " + userEmail + "\r\n" +
		"Subject: [DIMO] Vehicle Segment Export\r\n" +
		"Content-Type: multipart/alternative; boundary=\"" + w.Boundary() + "\"\r\n" +
		"\r\n")
	if _, err := partsBuffer.WriteTo(&buffer); err != nil {
		return err
	}

	return smtp.SendMail(addr, auth, es.emailFrom, []string{userEmail}, buffer.Bytes())
}

func (es *EmailService) getVerifiedEmailAddress(userID string) (string, error) {

	// usersClient := pb_users.NewUserServiceClient(es.ClientConn)
	// user, err := usersClient.GetUser(context.Background(), &pb_users.GetUserRequest{Id: userID})
	// if err != nil {
	// 	if s, ok := status.FromError(err); ok && s.Code() == codes.NotFound {
	// 		es.log.Debug().Str("userId", userID).Msg("user not found.")
	// 		return "", err
	// 	}
	// 	return "", err
	// }

	// addr := user.GetEmailAddress()
	// if addr == "" {
	// 	es.log.Debug().Str("userId", userID).Msg("user does not have confirmed email address")
	// 	return "", errors.New("user does not have confirmed email address")
	// }

	addr := userID + "@gmail.com"

	return addr, nil
}
