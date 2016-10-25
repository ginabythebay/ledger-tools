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
	"strings"

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

// Message represents a single email message
type Message struct {
	Date      string
	To        string
	From      string
	Subject   string
	TextPlain string
}

func decode(msg *gmail.Message) Message {
	var date, to, from, subject string
	payload := msg.Payload
	for _, h := range payload.Headers {
		switch h.Name {
		case "Date":
			date = h.Value
		case "To":
			to = h.Value
		case "From":
			from = h.Value
		case "Subject":
			subject = h.Value
		}
	}
	textPlain := findBody(payload, "text/plain")
	return Message{date, to, from, subject, textPlain}
}

func findBody(part *gmail.MessagePart, mimeType string) string {
	if part.MimeType == mimeType {
		b, _ := base64.StdEncoding.DecodeString(part.Body.Data)
		return string(b)
	}
	for _, child := range part.Parts {
		return findBody(child, mimeType)
	}
	return ""
}

type messageList struct {
	ids           []string
	nextPageToken string
}

func (gm *Gmail) queryPage(query string, nextPageToken string) (*messageList, error) {
	queryCall := gm.svc.Users.Messages.List(me).Q(query)
	if nextPageToken != "" {
		queryCall.PageToken(nextPageToken)
	}
	r, err := queryCall.Do()
	if err != nil {
		return nil, errors.Wrapf(err, "query %q, %q", query, nextPageToken)
	}
	var ids []string
	for _, msg := range r.Messages {
		ids = append(ids, msg.Id)
	}
	return &messageList{ids, r.NextPageToken}, nil
}

// QueryOption represents an option we can use to modify a query for messages.
type QueryOption func() string

// QuerySubject allows us to query for words in a subject.
func QuerySubject(subject string) QueryOption {
	return func() string {
		return fmt.Sprintf(`subject:"%s"`, subject)
	}
}

// QueryFrom allows us to query for a from email address.
func QueryFrom(from string) QueryOption {
	return func() string {
		return fmt.Sprintf("from:%s", from)
	}
}

// QueryNewerThan lets us query for message newer than days.
func QueryNewerThan(days int) QueryOption {
	return func() string {
		return fmt.Sprintf("newer_than:%dd", days)
	}
}

func dump(v interface{}) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", string(b))
	os.Exit(1)
}

// QueryMessages applies the opts to do a query and returns the
// matching messages.
func (gm *Gmail) QueryMessages(opts ...QueryOption) ([]Message, error) {
	var tokens []string
	for _, o := range opts {
		tokens = append(tokens, o())
	}
	query := strings.Join(tokens, " ")

	var result []Message
	var nextPageToken string
	for {
		page, err := gm.queryPage(query, nextPageToken)
		if err != nil {
			return nil, errors.Wrap(err, "query page")
		}
		for _, id := range page.ids {
			msg, err := gm.svc.Users.Messages.Get(me, id).Do()
			if err != nil {
				return nil, errors.Wrapf(err, "Getting msg %q", id)
			}

			result = append(result, decode(msg))
		}
		nextPageToken = page.nextPageToken
		if nextPageToken == "" {
			break
		}
	}
	return result, nil
}