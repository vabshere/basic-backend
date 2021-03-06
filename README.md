# Basic-Backend

This is the backend based on MVC pattern for **Basic**, a chat room with [Angular](https://angular.io) frontend and [Golang](https://golang.org) + [MySQL](https://www.mysql.com/) backend.
This project was generated in [Go](https://golang.org) version 1.14.1.

# Installation

_This app_:<br />
`go get github.com/vabshere/basic-backend`<br />

_External packages_:<br />
`golang.org/x/net/websocket`<br />
`go get github.com/go-sql-driver/mysql`<br />
`go get github.com/gorilla/mux`<br />

_MySQL_:<br />
You will also need to install [MySQL Server](https://www.mysql.com). This project was built using `version 8.0`. Refer the offficial docs for installation.

Update `connectDb()` function in `basic/models/user.go` with your username and password.

# Usage

Run `go run /path/to/main.go`. This should start the local server on any interface `:8080`.
