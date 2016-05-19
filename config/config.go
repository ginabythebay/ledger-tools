package config

import (
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

// Match describes a comment substring match and the associated
// transformations if the transaction matches.
type Match struct {
	// Treat it as a match if we see this as a substring in a comment
	Comment string
	Replace struct {
		// Change the payee to this
		Payee string
		// Treated as a go template.  We replace the posting that is
		// associated with the account with this posting or postings.
		Posting string
	}
}

// Config describes the automated modifications we make for a single account.
type Config struct {
	// Name of the account this applies to
	Account string
	// See above.  We match these in order and if we find a match, we stop
	// looking.
	Match []Match
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
	err = yaml.Unmarshal(bytes, config)

	return config, err
}
