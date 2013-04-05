# Keith Rarick's Blog Generator

blog generates html files for a blog based on
markdown input and web page templates

## Usage

Run it with `go run main.go src dst`

## Configuration

Content is read from the `src` directory and output is written into `dst`. The
destination directory is removed on startup.

The `src` directory must have the following files `base.layout`, `page.layout`,
`article.layout` and `redirects`.

The `.layout` files are [html/template](http://golang.org/pkg/html/template/)
files, the `redirects` file is a tab separated file with "file" and "redirect"
on each line. The `redirects` file is only working with the built-in web server,
the other files can be served by any web server.
