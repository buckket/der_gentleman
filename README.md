# der_gentleman [![Go Report Card](https://goreportcard.com/badge/github.com/buckket/der_gentleman)](https://goreportcard.com/report/github.com/buckket/der_gentleman)  [![GoDoc](https://godoc.org/github.com/buckket/der_gentleman?status.svg)](https://godoc.org/github.com/buckket/der_gentleman)

**der_gentleman** is an Instagram comment scraper / Twitter bot _art_ installation.
Nevertheless, the scraper can be used on its own and may provide useful insights.

Instagram removed its Activity tab a while ago, and thus made it nearly impossible to catch up on all the comments a particular account writes and the post he/she likes.

der_gentleman solves this problem by going through every comment of every post of every account the target account is following.
If a comment from the target account is found it is added to the database. While we’re at it we also check if the target account has liked the post.
This is however somewhat unreliable as we parse the "top liker" string to mitigate heavy rate limiting on the "liked by" endpoint.

This code was written with a very specific use case in mind but can be easily adopted to fit one’s needs.

Quick stats: To recheck the last 16 posts (= one result page) of ~1300 accounts takes about one hour. Initial scraping  may take longer, about seven hours or so.

## Installation

### From source

    go get -u github.com/buckket/der_gentleman

## Usage

1) Edit `config.toml`
2) Scrape all the data
3) ???
4) Profit

## Notes

There’s a [bug](https://github.com/ahmdrz/goinsta/pull/306) in the goinsta library, which has not been fixed upstream, thus this code uses my fork which fixes said bug.

## License

 GNU GPLv3+
 