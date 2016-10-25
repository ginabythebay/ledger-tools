package gmail

import (
	"context"
	"fmt"
	"io/ioutil"
	"os/user"
	"path"

	"github.com/pkg/errors"

	"golang.org/x/oauth2/google"
	gmail "google.golang.org/api/gmail/v1"
)

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
	user := "me"
	r, err := gm.svc.Users.Labels.List(user).Do()
	if err != nil {
		return nil, errors.Wrap(err, "Labels List")
	}

	var result []string
	for _, l := range r.Labels {
		result = append(result, l.Name)
	}
	return result, nil
}
