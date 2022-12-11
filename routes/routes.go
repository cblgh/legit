package routes

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/alexedwards/flow"
	"icyphox.sh/legit/config"
	"icyphox.sh/legit/git"
)

type deps struct {
	c *config.Config
}

func (d *deps) RepoIndex(w http.ResponseWriter, r *http.Request) {
	name := flow.Param(r.Context(), "name")
	name = filepath.Clean(name)
	// TODO: remove .git
	path := filepath.Join(d.c.Git.ScanPath, name+".git")
	gr, err := git.Open(path, "")
	if err != nil {
		Write404(w, *d.c)
		return
	}

	files, err := gr.FileTree("")
	if err != nil {
		Write500(w, *d.c)
		log.Println(err)
		return
	}

	data := make(map[string]any)
	data["name"] = name
	// TODO: make this configurable
	data["ref"] = "master"

	d.listFiles(files, data, w)
	return
}

func (d *deps) RepoTree(w http.ResponseWriter, r *http.Request) {
	name := flow.Param(r.Context(), "name")
	treePath := flow.Param(r.Context(), "...")
	ref := flow.Param(r.Context(), "ref")

	name = filepath.Clean(name)
	// TODO: remove .git
	path := filepath.Join(d.c.Git.ScanPath, name+".git")
	gr, err := git.Open(path, ref)
	if err != nil {
		Write404(w, *d.c)
		return
	}

	files, err := gr.FileTree(treePath)
	if err != nil {
		Write500(w, *d.c)
		log.Println(err)
		return
	}

	data := make(map[string]any)
	data["name"] = name
	data["ref"] = ref
	data["parent"] = treePath

	d.listFiles(files, data, w)
	return
}

func (d *deps) FileContent(w http.ResponseWriter, r *http.Request) {
	name := flow.Param(r.Context(), "name")
	treePath := flow.Param(r.Context(), "...")
	ref := flow.Param(r.Context(), "ref")

	name = filepath.Clean(name)
	// TODO: remove .git
	path := filepath.Join(d.c.Git.ScanPath, name+".git")
	gr, err := git.Open(path, ref)
	if err != nil {
		Write404(w, *d.c)
		return
	}

	contents, err := gr.FileContent(treePath)
	data := make(map[string]any)
	data["name"] = name
	data["ref"] = ref

	d.showFile(contents, data, w)
	return
}

func (d *deps) Log(w http.ResponseWriter, r *http.Request) {
	name := flow.Param(r.Context(), "name")
	ref := flow.Param(r.Context(), "ref")

	path := filepath.Join(d.c.Git.ScanPath, name+".git")
	gr, err := git.Open(path, ref)
	if err != nil {
		Write404(w, *d.c)
		return
	}

	commits, err := gr.Commits()
	if err != nil {
		Write500(w, *d.c)
		log.Println(err)
		return
	}

	tpath := filepath.Join(d.c.Template.Dir, "*")
	t := template.Must(template.ParseGlob(tpath))

	data := make(map[string]interface{})
	data["commits"] = commits
	data["meta"] = d.c.Meta
	data["name"] = name
	data["ref"] = ref

	if err := t.ExecuteTemplate(w, "log", data); err != nil {
		log.Println(err)
		return
	}
}
