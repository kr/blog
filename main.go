package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/kr/smartypants"
	"github.com/russross/blackfriday"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	ttemplate "text/template"
	"time"
)

const (
	timeSlug = "2006-01-02"
	timePath = "2006/01/02"
)

type Page struct {
	Title    template.HTML
	Summary  template.HTML
	Path     string
	Id       string
	Content  template.HTML
	Date     time.Time
	Articles []Page
}

var dstDir string

var (
	baseLayout *template.Template
	htmlfuncs  = template.FuncMap{
		"rev":     rev,
		"datauri": datauri,
	}
	textfuncs = ttemplate.FuncMap(htmlfuncs)
)

func mustGetOutput(c *exec.Cmd) []byte {
	b, err := c.Output()
	if err != nil {
		panic(err)
	}
	return b
}

func lines(b []byte) []string {
	a := bytes.Split(b, []byte("\n"))
	s := make([]string, len(a))
	for i := range a {
		s[i] = string(a[i])
	}
	return s
}

func makeSet(a []string) map[string]bool {
	m := make(map[string]bool)
	for _, s := range a {
		m[s] = true
	}
	return m
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: blog src dst")
		os.Exit(1)
	}
	translate(os.Args[1], os.Args[2])
	serve(os.Args[2])
}

func translate(src, dst string) {
	if err := os.Chdir(src); err != nil {
		panic(err)
	}
	dstDir = dst

	if err := os.RemoveAll(dst); err != nil {
		panic(err)
	}

	baseLayout = template.Must(template.New("base.layout").Funcs(htmlfuncs).ParseFiles("base.layout"))

	skipDir := map[string]bool{
		".git":    true,
		"article": true,
		"draft":   true,
	}

	a := renderArticles(filepath.Glob("article/*.md"))
	for _, article := range a {
		article.renderHTML("article.layout")
	}
	filepath.Walk(".", func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			panic(err)
		}
		if skipDir[fi.Name()] {
			return filepath.SkipDir
		}
		if !fi.IsDir() && fi.Name()[0] != '.' {
			handle(path, a)
		}
		return nil
	})
}

func renderArticles(paths []string, err error) []Page {
	if err != nil {
		panic(err)
	}
	a := make([]Page, len(paths))
	for i, p := range paths {
		title, summary, body := readArticle(p)
		a[i] = Page{
			Title:    template.HTML(educate(title)),
			Summary:  template.HTML(educate(summary)),
			Date:     dateFromName(filepath.Base(p)),
			Articles: a,
			Content:  template.HTML(blackfriday.MarkdownCommon(body)),
		}
		a[i].Id = articleId(a[i].Date, filepath.Base(p))
		a[i].Path = a[i].Id + ".html"
	}
	return a
}

func educate(b []byte) []byte {
	buf := new(bytes.Buffer)
	smartypants.New(buf, 0).Write(b)
	return buf.Bytes()
}

func readArticle(p string) (title, summary, body []byte) {
	b, err := ioutil.ReadFile(p)
	if err != nil {
		panic(err)
	}
	title, body, err = splitPara(b)
	if err != nil || title[0] != '#' || title[1] != ' ' {
		panic("bad title format in " + p)
	}
	title = title[2:] // omit leading "# "
	if s, b, err := splitPara(body); err == nil && s[0] == '/' && s[1] == ' ' {
		summary, body = s[2:], b // omit leading "/ "
	}
	return
}

func splitPara(b []byte) (head, rest []byte, err error) {
	if i := bytes.Index(b, []byte("\n\n")); i < 0 {
		err = errors.New("no paragraph")
	} else {
		head, rest = b[0:i], b[i+2:]
	}
	return
}

func handle(p string, a []Page) {
	ext := path.Ext(p)
	if strings.HasSuffix(ext, "~") {
		return
	}
	base := p[:len(p)-len(ext)]
	page := Page{
		Articles: a,
	}
	for _, a := range a {
		if a.Date.After(page.Date) {
			page.Date = a.Date
		}
	}
	switch ext {
	case ".t":
		page.Path = base
		page.renderText(p)
	case ".md":
		title, summary, body := readArticle(p)
		page.Path = base + ".html"
		page.Title = template.HTML(educate(title))
		page.Summary = template.HTML(educate(summary))
		page.Content = template.HTML(blackfriday.MarkdownCommon(body))
		page.renderHTML("page.layout")
	case ".p":
		page.Path = base + ".html"
		page.renderHTML(p)
	case ".layout":
	default:
		copyFile(p)
	}
}

func copyFile(p string) {
	if r, err := os.Open(p); err != nil {
		panic(err)
	} else {
		if _, err = io.Copy(createDst(p), r); err != nil {
			panic(err)
		}
	}
}

func createDst(urlPath string) io.Writer {
	dstPath := dstDir + "/" + urlPath
	err := os.MkdirAll(filepath.Dir(dstPath), 0777)
	if err != nil {
		panic(err)
	}
	w, err := os.Create(dstPath)
	if err != nil {
		panic(err)
	}
	return w
}

func (p Page) renderText(path string) {
	w := createDst(p.Path)
	t := ttemplate.Must(ttemplate.New(path).Funcs(textfuncs).ParseFiles(path))
	if err := t.Execute(w, p); err != nil {
		panic(err)
	}
}

func (p Page) renderHTML(tpath string) {
	w := createDst(p.Path)
	t := template.Must(baseLayout.Clone())
	content, err := ioutil.ReadFile(tpath)
	if err != nil {
		panic(err)
	}
	template.Must(t.New("page").Parse(string(content)))
	err = t.ExecuteTemplate(w, "base.layout", p)
	if err != nil {
		panic(err)
	}
}

func dateFromName(s string) time.Time {
	q := s[:len(s)-len(".md")]
	t, err := time.Parse(timeSlug, q[0:len(timeSlug)])
	if err != nil {
		panic(err)
	}
	return t
}

func articleId(t time.Time, s string) string {
	return t.Format(timePath) + "/" + s[len(timeSlug)+1:len(s)-len(".md")]
}

func datauri(contentType, path string) string {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return "data:" + contentType + ",notfound"
	}
	return "data:" + contentType + ";base64," + base64.StdEncoding.EncodeToString(b)
}

func rev(a []Page) []Page {
	b := make([]Page, len(a))
	for i := range a {
		b[len(b)-1-i] = a[i]
	}
	return b
}

func markdownPath(p string) []byte {
	b, err := ioutil.ReadFile(p)
	if err != nil {
		panic(err)
	}
	fmt.Println("render article", p)
	return blackfriday.MarkdownCommon(b)
}

func serve(dir string) {
	addr := ":" + os.Getenv("PORT")
	if addr == ":" {
		addr = ":8000"
	}
	for k, v := range readTable("redirect") {
		http.Handle(k, http.RedirectHandler(v, http.StatusMovedPermanently))
	}
	http.Handle("/", http.FileServer(http.Dir(dir)))
	log.Println("listen", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func readTable(path string) map[string]string {
	m := make(map[string]string)
	body, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	for _, line := range strings.Split(string(body), "\n") {
		if strings.HasPrefix(line, "#") {
			continue
		}
		if part := strings.SplitN(line, "\t", 2); len(part) == 2 {
			m[part[0]] = part[1]
		}
	}
	return m
}
