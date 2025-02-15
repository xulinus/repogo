package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"text/template"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/gorilla/mux"

	"github.com/xulinus/repogo/pkg/global"
)

type C struct {
	Sha    string `json:"sha"`
	Commit struct {
		Author struct {
			Name  string `json:"name"`
			Email string `json:"email"`
			Date  string `json:"date"`
		} `json:"author"`
		Committer struct {
			Name  string `json:"name"`
			Email string `json:"email"`
			Date  string `json:"date"`
		} `json:"committer"`
		Message string `json:"message"`
		Tree    struct {
			Sha string `json:"sha"`
			Url string `json:"url"`
		} `json:"tree"`
		Url           string `json:"url"`
		Comment_count int64  `json:"comment_count"`
		Verification  struct {
			Verified    bool   `json:"verified"`
			Reason      string `json:"reason"`
			Signature   string `json:"signature"`
			Payload     string `json:"payload"`
			Verified_at string `json:"verified_at"`
		} `json:"verification"`
	} `json:"commit"`
}

type Changelog struct {
	Date, Revision, Whom, Change string
}

func Doc(w http.ResponseWriter, r *http.Request) {
	doc := mux.Vars(r)["doc"]
	url := global.GH_API_REPO_URL + global.REPO + "commits?path=" + doc

	commitsJson, err := getHttpBodyInBytes(url)
	if err != nil {
		log.Println(err)
	}

	commits, err := commitsJsonToStruct(commitsJson)
	if err != nil {
		log.Println(err)
	}

	var changelog []Changelog
	for _, v := range commits {

		revision := v.Sha[:7]
		whom := fmt.Sprintf("%s (%s)", v.Commit.Author.Name, v.Commit.Author.Email)
		message := strings.Split(v.Commit.Message, "\n\n")[0]

		changelog = append(changelog, Changelog{
			Date:     v.Commit.Author.Date,
			Revision: revision,
			Whom:     whom,
			Change:   message,
		})
	}

	dom, err := template.ParseFiles("tmpl/doc.html")
	if err != nil {
		print(err)
	}

	err = dom.Execute(w, struct {
		Changelog []Changelog
	}{
		Changelog: changelog,
	})
}

func NonListFileServer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func commitsJsonToStruct(j []byte) ([]C, error) {
	var commits []C
	err := json.Unmarshal(j, &commits)
	if err != nil {
		return nil, err
	}
	return commits, nil
}

func getHttpBodyInBytes(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, err
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return bodyBytes, nil
}

func mdToHTML(md []byte) []byte {
	// create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return markdown.Render(doc, renderer)
}
