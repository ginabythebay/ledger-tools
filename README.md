
# Ledger Tools

[![Build Status](https://travis-ci.org/ginabythebay/ledger-tools.svg?branch=master)](https://travis-ci.org/ginabythebay/ledger-tools)


## importing tasks still to be done

* automated and benchmark tests for register import
* See if we can solve the notes problem for register import
* Take a look at converting register import to a streaming system
* instead of converting csv to csv, convert it directly into ledger format
* recognize existing entries that make for duplicates.  Prefer email import to csv import
* handle Patreon emails.  from bingo@patreon.com.  subject 'Thank you for supporting your creators!'
* handle apple itunes emails
* handle long tall sally emails.  from: 'customerservices@longtallsally.com'.  Subject like: 'About your order...'.  need to parse html.  They don't send text/plain parts!
* handle asos orders.  from: 'order_confirm@asos.com'.  subject like: 'Thanks for your order!'.  looks like I need to parse html again.  possibly double-base64 encoded.
