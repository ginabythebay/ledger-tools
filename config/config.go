package config

import (
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

// Replace describes a comment substring match and the associated
// transformations if the transaction matches.
type Replace struct {
	// Treat it as a match if we see this as a substring in a comment
	Comment string
	// Change the payee to this
	Payee string
	// Treated as a go template.  We replace the posting that is
	// associated with the account with this posting or postings.
	Posting string
}

// Config encapsulates replacements for each posting account
type Config struct {
	// Maps from posting account name to matches
	PostingAccount map[string][]Replace
}

// ParseYamlConfig reads a configuration from a yaml file
func ParseYamlConfig(file string) (config *Config, err error) {
	in, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer in.Close()

	bytes, err := ioutil.ReadAll(in)
	if err != nil {
		return nil, err
	}

	config = &Config{}
	err = yaml.Unmarshal(bytes, &config.PostingAccount)

	return config, err
}
