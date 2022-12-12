package main

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var tmplt *template.Template

func runServer() {
	http.HandleFunc("/player", playPage)
	http.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("./static"))))
	http.Handle("/", http.FileServer(http.Dir(".")))
	http.HandleFunc("/list", homePage)
	http.HandleFunc("/home", handlePage)
	http.HandleFunc("/video", handleStream)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalln("There's an error with the server:", err)
	}
}

func playPage(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL.Path)
	p := "." + r.URL.Path
	if p == "./player" {
		p = "./static/player.html"
	}
	http.ServeFile(w, r, p)
}

func homePage(resp http.ResponseWriter, req *http.Request) {

	dir := req.URL.Query().Get("dir")
	if dir == "" {
		dir = "Movies"
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}
	dir = homeDir + dir
	files := getAllMp4s(dir)
	log.Println("all files: ", files)
	type File struct {
		Name string
	}
	var list []File
	for _, file := range files {
		list = append(list, File{file})
	}
	tmplt, err := template.ParseFiles("home.html")
	if err != nil {
		log.Fatalf("parse home.html err: %v", err)
	}

	if err := tmplt.Execute(resp, list); err != nil {
		return
	}
}

func handlePage(writer http.ResponseWriter, request *http.Request) {
	if request.Method == "GET" {
		tmplt, _ = template.ParseFiles("tutorial.html")
		name := request.URL.Query().Get("name")
		log.Println("home param: ", name)
		if name == "" {
			writer.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		event := News{
			Headline: name,
			Body:     "body?",
			Name:     name,
		}
		err := tmplt.Execute(writer, event)
		if err != nil {
			return
		}
	}
}

type News struct {
	Headline string
	Body     string
	Name     string
}

func main() {
	runServer()
}

func handleStream(writer http.ResponseWriter, request *http.Request) {
	// f, err := os.Open("test.html")
	// f, err := os.Open("video.mp4")
	name := request.URL.Query().Get("name")
	f, err := os.Open(name + ".mp4")
	// f, err := os.Open("video-b.mp4")
	if err != nil {
		writer.WriteHeader(http.StatusNotFound)
		return
	}
	defer f.Close()
	// defer f.Close()
	// _, _ = io.Copy(writer, f)
	if fi, err := f.Stat(); err == nil {
		size := int(fi.Size())
		log.Printf("size: %v range: %v", size, request.Header["Range"])
		if r, ok := request.Header["Range"]; ok {
			s := r[0][6:]
			rs := strings.Split(s, "-")
			start, _ := strconv.Atoi(rs[0])
			end, err := strconv.Atoi(rs[1])
			if err != nil {
				end = size - 1
			}
			chunksize := (end - start) + 1
			log.Printf("start: %v, end: %v, chunkSize: %d", start, end, chunksize)
			writer.Header().Add("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, size))
			writer.Header().Add("Accept-Ranges", "bytes")
			writer.Header().Add("Content-Length", strconv.Itoa(chunksize))
			writer.Header().Add("Content-Type", "video/mp4")
			writer.Header().Add("Connection", "keep-alive")
			writer.Header().Add("Keep-Alive", "timeout=5")
			writer.WriteHeader(http.StatusPartialContent)
			// if chunksize < 10 {
			// 	buf := make([]byte, chunksize)
			// 	n, _ := f.ReadAt(buf, int64(start))
			// 	log.Printf("detect range: start: %v, size: %v, read: %d", start, chunksize, n)
			// 	writer.Write(buf[:n])
			// 	return
			// }
			_, _ = f.Seek(int64(start), 0)
			_, _ = io.CopyN(writer, f, int64(chunksize))
		} else {
			writer.Header().Add("Connection", "keep-alive")
			writer.Header().Add("Keep-Alive", "timeout=5")
			writer.Header().Add("Content-Type", "video/mp4")
			writer.Header().Add("Content-Length", strconv.Itoa(size))
			writer.WriteHeader(http.StatusOK)
			_, _ = io.Copy(writer, f)
			// writer.WriteHeader(206)
			// ioutil.ReadAll(f)
		}
	} else {
		log.Printf("f's stat err: %v", err)
	}
}

func getAllMp4s(root string) []string {
	var all []string
	log.Println("get all mp4s...")
	f := func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			log.Println(err)
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".mp4") {
			fmt.Printf("%v %+v\n", path, info)
			all = append(all, strings.TrimSuffix(path, ".mp4"))
		}
		return nil
	}
	filepath.Walk(root, f)
	return all
}
