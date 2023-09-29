# sqembv

`sqembv` is a Golang embedded web-based SQLite database browser. 
It is a fork of [sqlite-gobroem](https://github.com/bakaoh/sqlite-gobroem) which is a little bit modernized and improved.

Warning: avoid exposing this tool to the public Internet. 
It does not provide any authentication or authorization and 
can be used to read and modify your database.

## Embedded Use 

Use go get to install the latest version of the library:

```bash
$ go get -u github.com/andrianbdn/sqembv
```

Include sqembv in your application:

```go
import "github.com/andrianbdn/sqembv"
```

Initialize the API controller:

```go
api, err := sqembv.NewAPI("path to sqlite db file")
if err != nil {
    log.Fatal("can not open db", err)
}
```

Register the API handler:

```go
http.Handle("/browser/", sqembv.Handler("/browser/"))
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

