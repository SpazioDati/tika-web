tika-web is an HTTP frontend to Apache Tika.

Apache Tika gives to the developers a (rich) socket interface with support for XML, HTML, plain text, ... outputs, and an HTTP interface with only plain text support.

This projects is an HTTP interface to the socket based Tika Server, and it adds support for URL input (the HTTP version of Tika requires you to POST the file content instead).

## Installation

You need go (>= 1.0, but 1.1 is preferred because the "net" package is more reliable) to build tika-web:

```
git clone git://github.com/SpazioDati/tika-web.git
cd tika-web
go build
```

## Usage

You need tika server listening on port 9876:

```
java -jar tika-app.jar --server --port 9876
```

And then run tika-web:
```
./tika-web
```

This will start tika-web on 9875. If you need to change this port, or you bound the tika-app to another address, read the help with:

```
./tika-web --help
```
