package config

import "regexp"

import "testing"

type verification struct {
	t   *testing.T
	cfg *Config
}

func (v verification) find(account string, comment string) *Replace {
	candidates, ok := v.cfg.PostingAccount[account]
	if !ok {
		v.t.Fatalf("Unable to find account %#q", account)
	}

	for _, r := range candidates {
		if r.Comment == comment {
			return &r
		}
	}
	v.t.Fatalf("Unable to find comment %#q for account %#q", comment, account)
	return nil // we never actually hit this, but the compiler does not realize that
}

func (v verification) verifyReplacement(account string, comment string, payee string, posting *regexp.Regexp) {
	r := v.find(account, comment)
	if r.Payee != payee {
		v.t.Fatalf("Entry %#q/%#q, has payee %#q when we expected %#q ", account, comment, r.Payee, payee)
	}

	if !posting.MatchString(r.Posting) {
		v.t.Fatalf("Entry %#q/%#q posting %#q does not match %#q ",
			account, comment, r.Posting, posting)
	}
}

func (v verification) verifyReplacementCount(account string, expected int) {
	found := len(v.cfg.PostingAccount[account])
	if found != expected {
		v.t.Errorf("Error verifying replacement count for %q.  Expected %d and got %d", account, expected, found)
	}
}

func TestReadConfigScenario(t *testing.T) {
	cfg, err := ParseYamlConfig("testdata/example_config.yaml")
	if err != nil {
		t.Fatal(err)
	}

	if len(cfg.PostingAccount) != 2 {
		t.Errorf("Expected %d entries but found %d entries", 2, len(cfg.PostingAccount))
	}

	v := verification{t, cfg}
	v.verifyReplacementCount("Income:Unknown", 1)
	v.verifyReplacement("Income:Unknown", "ATM fee refund", "ATM fee refund", regexp.MustCompile("\\s*Expenses:Bank Fees   -\\{\\{\\.PostingValue\\}\\}\\s+Expenses:Fix Me  \\$0; adjust bank fees as needed"))

	v.verifyReplacementCount("Expenses:Unknown", 3)
	v.verifyReplacement("Expenses:Unknown", "CO: PACIFIC GAS & EL", "PGE", regexp.MustCompile("\\s*Expenses:Utilities:Gas   \\$0\\s+Expenses:Utilities:Electric  \\$0\\s+Expenses:Fix Me  \\{\\{\\.PostingValue\\}\\}"))
	v.verifyReplacement("Expenses:Unknown", "CO: SAN FRANCISCO WA", "SFPUC", regexp.MustCompile("\\s*Expenses:Utilities:Water   \\$0\\s+Expenses:Utilities:Sewer  \\$0\\s+Expenses:Fix Me  \\{\\{\\.PostingValue\\}\\}"))
	v.verifyReplacement("Expenses:Unknown", "ATM Withdrawal #", "Cash Withdrawal", regexp.MustCompile("\\s*Expenses:Bank Fees   \\$3\\s+Expenses:Fix Me   \\$0; adjust bank fees as needed\\s+Assets:SFFire Checking   -\\{\\{\\.PostingValue\\}\\}\\s+Assets:Cash  \\(\\{\\{\\.Postingvalue\\}\\} - \\$3.00\\)"))
}
