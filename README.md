# streamvideo
Put you aa.mp4 files in the root directory, then
```bash
go run main.go
```
Use your broswer, e.g. http://localhost:8080/home?name=aa

# TODO
- [x] list files in the current directory
- [x] click on the files, play the videos
- [ ] refactor: using a Web framework? making the code readable?
- [x] if I want to config the root path?
- [x] to learn: http.ServeFile and http.FileServer, but it is not working if I directly write HTML content to resp writer.
