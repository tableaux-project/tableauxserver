# Tableaux-Server - The backend-server reference implementation for Tableaux

**Note:** This project is WIP. This is not production ready, and for the most part untested. Please help,
and further this project by contributing!

## About

Tableaux-Server is the reference implementation to wrap Tableaux with a simple Go HTTP server.

It contains simple CORS protection, and respons with GZIP compressed responses.

## Getting started

Using the server is easy and straight forward. Get yourself a Tableaux data source connector,
and plug it into the `NewServer` method - and you got yourself a Go HTTP server. Listen and serve!

```go
serverHandler, err := tableauxserver.NewServer(dataSourceConnector, schemaMapper, 8081)
if err != nil {
    log.Fatal(err)
}
log.Fatal(serverHandler.ListenAndServe())
```

## Logging

This library uses [loggers](https://github.com/birkirb/loggers/) for logging abstraction. This means, that is is easy to plug-in your prefered logging library of choice.

If you want to use [Logrus](https://github.com/sirupsen/logrus/) for example, the following code snippet should get you started:

```go
import (
    mapper "github.com/birkirb/loggers-mapper-logrus"
    "github.com/sirupsen/logrus"
    "gopkg.in/birkirb/loggers.v1/log"
)

func init() {
    l := logrus.New()
    l.Out = os.Stdout

    l.Formatter = &logrus.TextFormatter{
        ForceColors:      true,
        DisableTimestamp: false,
        FullTimestamp:    true,
        TimestampFormat:  "2006/01/02 15:04:05",
    }
    l.Level = logrus.DebugLevel

    log.Logger = mapper.NewLogger(l)
}
```

## Dependencies and licensing

Tableaux-Server is licenced via MIT, as specified [here](https://github.com/tableaux-project/tableaux-server/blob/master/LICENSE).

* [Tableaux](https://github.com/tableaux-project/tableaux) - [MIT](https://github.com/tableaux-project/tableaux/blob/master/LICENSE) - Core application
* [loggers](https://github.com/birkirb/loggers/) - [MIT](https://github.com/birkirb/loggers/blob/master/LICENSE.txt) - Abstract logging for Golang projects
* [Gzip Handler](https://github.com/NYTimes/gziphandler/) - [Apache 2.0](https://github.com/NYTimes/gziphandler/blob/master/LICENSE) - Golang middleware to gzip HTTP responses
* [gorilla/mux](https://github.com/gorilla/mux/) - [BSD-3](https://github.com/gorilla/mux/blob/master/LICENSE) - A powerful URL router and dispatcher for golang
* [Go CORS](https://github.com/rs/cors/) - [MIT](https://github.com/rs/cors/blob/master/LICENSE) - Go net/http configurable handler to handle CORS requests

## Versioning

We use [SemVer](http://semver.org/) for versioning.