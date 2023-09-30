# sqembv

`sqembv` is a Golang embedded web-based SQLite database "browser". 
It is a fork of [sqlite-gobroem](https://github.com/bakaoh/sqlite-gobroem) which is a little bit modernized and improved.

Main reason for the fork was extremely slow compilation speed of the original project.
Also, this fork switched to CGO-optional SQLite driver and fixes some bugs.

**Warning**: avoid exposing this tool to the public Internet. 
It does not provide any authentication or authorization and 
can be used to read and modify your database.

**Note**: While the original project (and this fork) claims to be a "browser", it allows to execute arbitrary SQL queries.


## Embedded Use

Include `sqembv` in your application:

```go
import "github.com/andrianbdn/sqembv"
```

Initialize the API controller:

```go
api, err := sqembv.NewAPI("test.sqlite3")
if err != nil {
    log.Fatalln("can not open db:", err)
}
```

Register the API handler:
```go
http.Handle("/browser/", api.Handler("/browser/"))
```

or, even better — register it with HTTP basic auth (minimal implementation below)

```go
h := api.Handler("/browser/")

u, p := "sql-user", "PasSw0rd" // Username and Password - CHANGE the password!
golden := "Basic " + base64.StdEncoding.EncodeToString([]byte(u+":"+p))

http.HandleFunc("/browser/", func(w http.ResponseWriter, r *http.Request) {
  if a := r.Header.Get("Authorization"); subtle.ConstantTimeCompare([]byte(a), []byte(golden)) != 1 {
    w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
    http.Error(w, "Unauthorized", http.StatusUnauthorized)
    return
  }
  h.ServeHTTP(w, r)
})
```

## Standalone

Use go install to install the latest version of the program:

```bash
$ go install github.com/andrianbdn/sqembv/cmd/sqembv@latest
```

Run sqembv:

```bash
$ ./sqembv -h

sqembv, v0.1.1
Usage of ./sqembv:
  -bind string
        HTTP server host (default "localhost")
  -db string
        SQLite database file (default "test.sqlite3")
  -license
        Print program license and exit
  -listen uint
        HTTP server listen port (default 8000)


$ ./sqembv
```

Open browser http://localhost:8000/

