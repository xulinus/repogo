package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"text/template"
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

func Doc(w http.ResponseWriter, r *http.Request) {
	commitsJson, err := gurl(
		"https://api.github.com/repos/xulinus/policy-docs/commits?path=testdoc.md",
	)
	if err != nil {
		log.Println(err)
	}

	commits, err := commitsJsonToStruct(commitsJson)
	if err != nil {
		log.Println(err)
	}

	dom, err := template.ParseFiles("tmpl/doc.html")
	if err != nil {
		print(err)
	}

	err = dom.Execute(w, struct {
		Commits []C
	}{
		Commits: commits,
	})
}

func gurl(url string) ([]byte, error) {
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

func commitsJsonToStruct(j []byte) ([]C, error) {
	var commits []C
	err := json.Unmarshal(j, &commits)
	if err != nil {
		return nil, err
	}
	return commits, nil
}
