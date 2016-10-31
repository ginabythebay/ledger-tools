
# Ledger Tools

[![Build Status](https://travis-ci.org/ginabythebay/ledger-tools.svg?branch=master)](https://travis-ci.org/ginabythebay/ledger-tools)


## importing tasks still to be done

* instead of converting csv to csv, convert it directly into ledger format
* recognize existing entries that make for duplicates.  Prefer email import to csv import
* Some non-amazonsmile orders have email subjects that look like: 'Your Amazon.com order of "book title" has shipped!'.  Handle those too.
* handle Patreon emails.  from bingo@patreon.com.  subject 'Thank you for supporting your creators!'
* handle apple itunes emails
* handle long tall sally emails.  from: 'customerservices@longtallsally.com'.  Subject like: 'About your order...'.  need to parse html.  They don't send text/plain parts!
* handle asos orders.  from: 'order_confirm@asos.com'.  subject like: 'Thanks for your order!'.  looks like I need to parse html again.  possibly double-base64 encoded.
