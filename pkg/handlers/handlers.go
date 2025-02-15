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
	Files []struct {
		Sha         string `json:"sha"`
		Filename    string `json:"filename"`
		Status      string `json:"status"`
		Additions   int64  `json:"additions"`
		Deletions   int64  `json:"deletions"`
		Changes     int64  `json:"changes"`
		Blob_url    string `json:"blob_url"`
		Raw_url     string `json:"raw_url"`
		Content_url string `json:"content_url"`
		Patch       string `json:"patch"`
	} `json:"files"`
}

type Changelog struct {
	Date, Revision, Whom, Change, FullSha string
}

type Revisionlog struct {
	Additions, Deletions, Changes, Message, Patch string
}

func Doc(w http.ResponseWriter, r *http.Request) {
	doc := mux.Vars(r)["doc"]
	sha := mux.Vars(r)["sha"]

	var mdUrl string
	var changelog []Changelog
	var revisionlog Revisionlog

	// if no sha is provided, we serve the lastes version of the document
	if sha == "" {
		mdUrl = global.GH_RAW_URL + global.REPO + "refs/heads/" + global.BRANCH + "/" + doc
		commitsUrl := global.GH_API_REPO_URL + global.REPO + "commits?path=" + doc

		commitsJson, err := getHttpBodyInBytes(commitsUrl)
		if err != nil {
			log.Println(err)
		}

		commits, err := GhApiJsonToStruct(commitsJson)
		if err != nil {
			log.Println(err)
		}

		changelog = changelogFromCommits(commits)

		// if sha is provided, we show information for the document at that version
	} else {
		mdUrl = global.GH_RAW_URL + global.REPO + sha + "/" + doc
		previousCommitUrl := global.GH_API_REPO_URL + global.REPO + "commits/" + sha

		previousCommitJson, err := getHttpBodyInBytes(previousCommitUrl)
		if err != nil {
			log.Println(err)
		}

		revisionData, err := GhApiJsonToStructMap(previousCommitJson)
		if err != nil {
			log.Println(err)
		}

		revisionlog = revisionlogFromRevisionData(revisionData, doc)

	}

	// we always want to print the document itself
	md, err := getHttpBodyInBytes(mdUrl)
	if err != nil {
		log.Println(err)
	}

	dom, err := template.ParseFiles("tmpl/doc.html")
	if err != nil {
		print(err)
	}

	err = dom.Execute(w, struct {
		Changelog   []Changelog
		Revisionlog Revisionlog
		Document    string
		Filename    string
	}{
		Changelog:   changelog,
		Revisionlog: revisionlog,
		Document:    string(mdToHTML(md)),
		Filename:    doc,
	})
}

func Main(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello World")
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

func GhApiJsonToStruct(j []byte) ([]C, error) {
	var commits []C
	err := json.Unmarshal(j, &commits)
	if err != nil {
		return nil, err
	}
	return commits, nil
}

func GhApiJsonToStructMap(j []byte) (C, error) {
	var c C
	err := json.Unmarshal(j, &c)
	if err != nil {
		return C{}, err
	}
	return c, nil
}

func changelogFromCommits(commits []C) []Changelog {
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
			FullSha:  v.Sha,
		})
	}

	return changelog
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

func revisionlogFromRevisionData(data C, doc string) Revisionlog {
	var revisionlog Revisionlog

	revisionlog.Message = data.Commit.Message

	for _, f := range data.Files {
		if f.Filename == doc {
			revisionlog.Additions = string(f.Additions)
			revisionlog.Deletions = string(f.Deletions)
			revisionlog.Changes = string(f.Changes)
			revisionlog.Patch = f.Patch
		}
	}

	return revisionlog
}
