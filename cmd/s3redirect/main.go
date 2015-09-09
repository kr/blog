/*

Command s3redirect reads a table of redirect rules from stdin
and prints an XML file to stdout in the format used by Amazon S3.

Usage

  s3redirect <redirect |pbcopy

Then paste the XML into the text box on the S3 console.

See $GOPATH/src/github.com/kr/blog/Readme.md
for details of the input format.
See https://docs.aws.amazon.com/AmazonS3/latest/dev/HowDoIWebsiteConfiguration.html
for details of the output format.

*/
package main

import (
	"bufio"
	"bytes"
	"log"
	"net/url"
	"os"
	"text/template"
)

func main() {
	v := map[string]*url.URL{}
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		b := scanner.Bytes()
		if bytes.IndexByte(b, '#') == 0 {
			continue
		}
		p := bytes.IndexByte(b, '\t')
		if p <= 0 || len(b) <= p {
			continue
		}
		u, err := url.Parse(string(b[p+1:]))
		if err != nil {
			log.Println(err)
			continue
		}
		v[string(b[1:p])] = u
	}
	if scanner.Err() != nil {
		log.Fatal(scanner.Err())
	}
	err := tpl.Execute(os.Stdout, v)
	if err != nil {
		log.Fatal(err)
	}
}

var tpl = template.Must(template.New("x").Parse(xml))

const xml = `
<RoutingRules>
	{{range $k, $u := .}}
	<RoutingRule>
		<Condition>
			<KeyPrefixEquals>{{$k}}</KeyPrefixEquals>
		</Condition>
		<Redirect>
			{{if $u.Scheme}}
				<Protocol>{{$u.Scheme}}</Protocol>
			{{end}}
			{{if $u.Host}}
				<HostName>{{$u.Host}}</HostName>
			{{end}}
			<ReplaceKeyWith>{{$u.Path}}</ReplaceKeyWith>
		</Redirect>
	</RoutingRule>
	{{end}}
</RoutingRules>
`
