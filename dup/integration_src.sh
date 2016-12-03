#!/bin/bash

set -e
set pipefail

ledger csv -f integration_src.ledger --csv-format '%(quoted(filename)),%(quoted(xact.beg_line)),%(quoted(join(xact.note))),%(quoted(date)),%(quoted(code)),%(quoted(payee)),%(quoted(beg_line)),%(quoted(display_account)),%(quoted(commodity(scrub(display_amount)))),%(quoted(quantity(scrub(display_amount)))),%(quoted(cleared ? "*" : (pending ? "!" : ""))),%(quoted(join(note)))\n' | perl -p  -e 's{\".*integration_src.ledger\"}{\"integration_src.ledger\"}'
