package gmail

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path"

	"github.com/pkg/errors"

	"golang.org/x/oauth2/google"
	gmail "google.golang.org/api/gmail/v1"
)

const me = "me"

type Gmail struct {
	svc *gmail.Service
}

// GetService returns a Gmail service.
func GetService() (*Gmail, error) {
	ctx := context.Background()

	usr, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("Unable to determine home directory: %v", err)
	}

	secretFile := path.Join(usr.HomeDir, ".config", "ledger-tools", "gmail_client_id.json")
	b, err := ioutil.ReadFile(secretFile)
	if err != nil {
		return nil, errors.Wrapf(err, "Reading %s", secretFile)
	}

	config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
	if err != nil {
		return nil, errors.Wrapf(err, "Parsing %s", secretFile)
	}
	client, err := getClient(ctx, config)
	if err != nil {
		return nil, errors.Wrap(err, "getClient")
	}

	srv, err := gmail.New(client)
	if err != nil {
		return nil, errors.Wrap(err, "getting new client")
	}
	return &Gmail{srv}, nil
}

// LabelNames returns the names of known labels.
func (gm *Gmail) LabelNames() ([]string, error) {
	r, err := gm.svc.Users.Labels.List(me).Do()
	if err != nil {
		return nil, errors.Wrap(err, "Labels List")
	}

	var result []string
	for _, l := range r.Labels {
		result = append(result, l.Name)
	}
	return result, nil
}

// Message represents a single email message
type Message struct {
	To        string
	From      string
	Subject   string
	TextPlain string
}

func decode(msg *gmail.Message) Message {
	var to, from, subject string
	payload := msg.Payload
	for _, h := range payload.Headers {
		switch h.Name {
		case "To":
			to = h.Value
		case "From":
			from = h.Value
		case "Subject":
			subject = h.Value
		}
	}
	textPlain := getTextPlain(payload)
	return Message{to, from, subject, textPlain}
}

func headerValue(name string, headers []*gmail.MessagePartHeader) string {
	for _, h := range headers {
		if h.Name == name {
			return h.Value
		}
	}
	return ""
}

func getTextPlain(part *gmail.MessagePart) string {
	if part.MimeType == "text/plain" && headerValue("Content-Transfer-Encoding", part.Headers) == "quoted-printable" {
		b, _ := base64.StdEncoding.DecodeString(part.Body.Data)
		return string(b)
	}
	for _, child := range part.Parts {
		return getTextPlain(child)
	}
	return ""
}

func dump(v interface{}) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", string(b))
	os.Exit(1)
}

func (gm *Gmail) LyftMessages() ([]Message, error) {
	r, err := gm.svc.Users.Messages.
		List(me).
		Q(`subject:"Your ride with" from:no-reply@lyftmail.com newer_than:30d`).
		Do()
	if err != nil {
		return nil, errors.Wrap(err, "Messages List")
	}
	// TODO(gina) handle paging

	result := make([]Message, len(r.Messages), len(r.Messages))
	for i, m := range r.Messages {
		msg, err := gm.svc.Users.Messages.Get(me, m.Id).Do()
		if err != nil {
			return nil, errors.Wrapf(err, "Getting msg %q", m.Id)
		}

		result[i] = decode(msg)
	}
	return result, nil
}
