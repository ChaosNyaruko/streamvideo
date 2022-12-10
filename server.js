const http=require('http');
const fs=require("fs");
const querystring=require("querystring");
const path=require("path");
const server=http.createServer((req,res)=>{
    // return res.end(req.url+req.method);
    if(req.method==='GET' && req.url==="/"){
        console.log("get index")
        res.writeHead(200);
        fs.createReadStream(path.resolve("index.html")).pipe(res);
        return;
    }
    console.log("url: ", req.url)
    if(req.method==='GET' && req.url.startsWith("/video?")){
        const params = (req.url.slice(7).split("="))[1]
        // const queryObject = querystring.parse(req.url);
        // console.log(queryObject)
        const filepath = path.resolve(params + ".mp4");
        const stat = fs.statSync(filepath)
        const fileSize = stat.size
        const range = req.headers.range
        if (range) {
            const parts = range.replace(/bytes=/, "").split("-")
            const start = parseInt(parts[0], 10)
            console.log("[raw] with range", start, "-", parts[1])
            const end = parts[1] ?parseInt(parts[1], 10) :fileSize - 1
            const chunksize = (end - start) + 1
            console.log("with range", start, "-", end, " = ", chunksize)
            const file = fs.createReadStream(filepath, {start,end})
            const head = {
                'Content-Range': `bytes ${start}-${end}/${fileSize}`,
                'Accept-Ranges': 'bytes',
                'Content-Length': chunksize,
                'Content-Type': 'video/mp4',
            }
            res.writeHead(206, head);
          file.pipe(res);
        }else{
           console.log("without range", fileSize)
           const head = {
               'Content-Length': fileSize,
               'Content-Type': 'video/mp4',
           }
           res.writeHead(200, head);
           fs.createReadStream(filepath).pipe(res);
        }
    }
    else{
        res.writeHead(400);
        res.end("bad request");
    }
})


const PORT = process.env.PORT || 3000;
server.listen(PORT,() => {
  console.log(`server listening on port:${PORT}`);
})
