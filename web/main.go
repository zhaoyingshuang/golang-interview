package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"strings"
)

//go:embed static/*
var staticFiles embed.FS

type Topic struct {
	ID       string    `json:"id"`
	Title    string    `json:"title"`
	Chapter  string    `json:"chapter"`
	Sections []Section `json:"sections"`
}

type Section struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Code    string `json:"code"`
}

func main() {
	// Static site generation mode
	if len(os.Args) > 1 && os.Args[1] == "generate" {
		outDir := "docs"
		if len(os.Args) > 2 {
			outDir = os.Args[2]
		}
		generateStaticSite(outDir)
		return
	}

	port := "9090"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	staticFS, _ := fs.Sub(staticFiles, "static")
	http.Handle("/", http.FileServer(http.FS(staticFS)))
	http.HandleFunc("/api/topics", handleTopics)
	http.HandleFunc("/api/topic/", handleTopic)

	fmt.Printf("Go 面试知识点学习平台已启动: http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Fprintf(os.Stderr, "启动失败: %v\n", err)
		os.Exit(1)
	}
}

func generateStaticSite(outDir string) {
	topics := getAllTopics()
	listData, _ := json.Marshal(getTopicList())
	detailMap := make(map[string]Topic)
	for _, t := range topics {
		detailMap[t.ID] = t
	}
	detailData, _ := json.Marshal(detailMap)

	tpl, _ := staticFiles.ReadFile("static/index.html")
	html := string(tpl)

	inject := fmt.Sprintf("\nconst TOPICS_LIST = %s;\nconst TOPICS_DETAIL = %s;\n", string(listData), string(detailData))

	// Inject data before first <script>
	html = strings.Replace(html, "<script>\nconst chapters", inject+"<script>\nconst chapters", 1)

	// Replace API fetch with static data
	html = strings.Replace(html, "async function init() {", "function init() {", 1)
	html = strings.Replace(html, "async function loadTopic", "function loadTopic", 1)
	html = strings.Replace(html, "const res = await fetch('/api/topics');\n    topics = await res.json();", "    topics = TOPICS_LIST;", 1)
	html = strings.Replace(html, "const res = await fetch(`/api/topic/${id}`);\n    const data = await res.json();", "    const data = TOPICS_DETAIL[id];", 1)

	// Add base tag for GitHub Pages sub-path
	html = strings.Replace(html, `<meta charset="UTF-8">`,
		`<meta charset="UTF-8">`+"\n    "+`<base href="/golang-interview/">`, 1)

	// Fix hash routing for sub-path
	html = strings.Replace(html, "history.replaceState(null, '', '#' + id);",
		"history.replaceState(null, '', location.pathname + '#' + id);", 1)
	html = strings.Replace(html, "if (location.hash) {\n        loadTopic(location.hash.slice(1));\n    }",
		"if (location.hash) { loadTopic(location.hash.slice(1)); }", 1)

	os.MkdirAll(outDir, 0755)
	os.WriteFile(outDir+"/index.html", []byte(html), 0644)
	fmt.Printf("Static site generated in %s/\n", outDir)
}

func handleTopics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(getTopicList())
}

func handleTopic(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/topic/")
	for _, t := range getAllTopics() {
		if t.ID == id {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(t)
			return
		}
	}
	http.NotFound(w, r)
}

func getTopicList() []map[string]string {
	topics := getAllTopics()
	result := make([]map[string]string, len(topics))
	for i, t := range topics {
		result[i] = map[string]string{
			"id":      t.ID,
			"title":   t.Title,
			"chapter": t.Chapter,
		}
	}
	return result
}

func getAllTopics() []Topic {
	return []Topic{
		sliceTopic(),
		mapTopic(),
		stringTopic(),
		interfaceTopic(),
		deferTopic(),
		pointerTopic(),
		goroutineTopic(),
		channelTopic(),
		syncTopic(),
		contextTopic(),
		patternTopic(),
		gcTopic(),
		escapeTopic(),
		alignmentTopic(),
		profilingTopic(),
		benchmarkTopic(),
	}
}

func s(title, content, code string) Section {
	return Section{Title: title, Content: content, Code: code}
}

// ==================== 01-basics ====================
