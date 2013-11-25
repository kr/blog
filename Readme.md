### Blog Generator

Command `blog` generates HTML files for a blog based on Markdown
input and web page templates.

#### Install

    $ go get github.com/kr/blog

#### Usage

    $ blog [output]

Files are read from `.`. If output contains a colon,
it is treated as an address and the site is served
there. Otherwise, it is treated as a path, it will
be removed, and rendered files will be written there.

Missing output argument is treated like `:8000`.

#### Special Files

The following files in `src` produce special output:

- `articles` - directory of Markdown articles
- `*.md` - Markdown page or article
- `*.p` - HTML template
- `*.t` - text template
- `*.layout` - HTML wrapper template (see below)
- others - copied unaltered to the destination

These files are required:

- `page.layout` - HTML template wrapping `.md` files
- `article.layout` - HTML template wrapping articles (`.md` files
in directory `articles`)
- `base.layout` - wraps page layout and article layout
- `redirect` - table of permanent redirects

##### Markdown Files

Files with extension `.md` are translated as follows:

- An [html/template](http://golang.org/pkg/html/template/) set
is created with two templates:
  1. Template `base.layout` from file `base.layout`
  2. Template `page` from file `page.layout`
- The Markdown contents of the file are rendered to HTML
- The template set is evaluated with the rendered page HTML
as template field `.Content`.
- Output path is the path and basename of the file + `.html`

##### Articles

Articles are `.md` files inside directory `articles`. They are
rendered like other Markdown pages, except:

- Filename must be of the form `2006-01-02-the-article-slug.md`.
- Its first line must be a level-one-heading `#` Markdown title.
- If the second paragraph begins with "/ ", its remainder is used
as the article summary.
- File `article.layout` is used instead of `page.layout`.
- Output path is of the form `/2006/01/02/the-article-slug.html`.

##### HTML Template Files

Files with extension `.p` are translated as follows:

- An [html/template](http://golang.org/pkg/html/template/) set
is created with two templates:
  1. Template `base.layout` from file `base.layout`
  2. Template `page` from the `.p` file itself
- Output path is the path and basename of the file + `.html`

##### Text Template Files

Files ending in `.t` are translated as a text/template template
with the same fields as a `.p` file.

Output path is the path and basename of the file.

##### Template Fields

The struct being passed to the template is defined at the top of
main.go.

##### Redirection

File `redirect` is a table of permanent redirects. It may be
empty, but it must be present. Each line is a record of two
tab-separated fields: request path and destination. It redirects
each request path to the corresponding destination URL, with
status 301 Moved Permanently.

The request path should begin with a slash `/`. The destination
URL is a full URL, but most components are optional; a relative
path destination will be made absolute by combining it with the
request path.

The redirection table works only with the built-in web server;
other files can be served by any web server.
