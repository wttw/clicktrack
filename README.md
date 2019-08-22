# clicktrack
This is a simple implementation of a click tracking webserver,
using cryptographically secured URLs.

It is part of this [series of blog posts](https://wordtothewise.com/2019/08/link-tracking-redirectors/).

## Usage

```
Usage of clicktrack:
  -config string
    	Use this configuration file (default "clicktrack-conf.json")
  -create
    	Create a new url, using key=value parameters or json on stdin
  -init
    	Create a new configuration file
  -parse string
    	Parse a redirector URL
  -rotate
    	Rotate the secret keys
  -serve
    	Run a redirecting webserver
```

The first time it is run clicktrack will create a configuration
file, `clicktrack-conf.json`, with sensible defaults in the
current directory.

### Generate a tracking url
```
$ clicktrack -create url=https://wordtothewise.com/blog/ myid=steve

http://127.0.0.1:3000/?x=2.UBC-ChQ31mqqYCCxnKOkvL4oic
IISkybvO26kL6R2E3PxHS5R2PPEye6zVHG0sEamragEHbY_iSTTuQ
pFbFLM0a3dPH6Xm3_0j8h2gBCybFpLUJW

```
This will take a list of key=value pairs on the command line
and print a tracking URL starting with the `BaseURL` configuration
setting that embeds those values.

It must contain a `url` value, which is the destination that
the link will redirect to. It may contain a `slug` value, which
will be included in the link as readable text.

If no key=value pairs are provided it will read the same data
as json on stdin.

### Decoding a tracking URL

```
$ clicktrack -parse http://127.0.0.1:3000/?x=2.UBC-ChQ31mqqYCC ...
{
  "myid": "steve",
  "url": "https://wordtothewise.com/blog/"
}
```
This takes a tracking URL that had previously been generated
and decodes it, showing the original parameters passed as json.

### Running a redirector server
```
$ clicktrack -serve
Listening on 127.0.0.1:3000
```

This will run a webserver on the address and port in the
`Listen` parameter in the configuration file.

It will answer requests for tracking URLs, returning a
redirect for valid ones and a 404 response in the case
of any error. It will print the parameters as json to
stdout for each request.

You can use `curl -D- http://127.0.0.1:3000/?x=...` from
another window to see exactly how it responds, or enter
a tracking URL into your regular browser to see it work.

### Rotating keys
```
$ clicktrack -rotate
```
This is an example of how to switch to a new signing key
(and potentially an entirely new protocol). It increases
the `Version` field in the configuration file and generates
a matching signing key. It doesn't delete older keys, so
previously generated URLs will continue to work.

## Installation

It's a simple Go application that only uses the standard library

```
git clone https://github.com/wttw/clicktrack.git
cd clicktrack
go build
```
