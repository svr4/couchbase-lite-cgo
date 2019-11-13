# Couchbase Lite for Cgo

This project serves as a Go lang wrapper for the couchbase-lite-C library developed by [Couchbase Labs](https://github.com/couchbaselabs).

It currently supports commit https://github.com/couchbaselabs/couchbase-lite-C/commit/e12e4a64911aa97be758e0957338c7d5b7201f5e of the library.

Replication is now working via HTTP. Support for HTTPS is coming down the pipeline.

The package was built and tested on macOS Mojave (10.14.6) using Go version `go1.12.9 darwin/amd64`.

## Example

```go
var config DatabaseConfiguration

// Optional encryption key.
var encryption_key EncryptionKey
encryption_key.algorithm = EncryptionNone
encryption_key.bytes = make([]byte, 0)

config.directory = "./db"
config.encryptionKey = encryption_key
config.flags = Database_Create

if db, err := Open("my_db", &config); err == nil {
    // Create a doc
    doc := NewDocumentWithId("test")
    doc.Props["name"] = "Luke"
    doc.Props["lastname"] = "Skywalker"
    doc.Props["age"] = 30
    doc.Props["email"] = "son.of.skywalker@gmail.com"
    doc.Props["action"] = "delete"

    if _, err2 := db.Save(doc, LastWriteWins); err2 == nil {
        // Retrieve the saved doc
        if savedDoc, e := db.GetMutableDocument("test"); e == nil {
            for k, v := range savedDoc.Props {
                if _doc.Props[k] != v {
                    fmt.Println("Saved document and retrieved document are different.")
                }
            }
            savedDoc.Release()
        } else {
            fmt.Println(e)
        }
    } else {
        fmt.Println(err2)
    }

    if !db.Close() {
		fmt.Println("Couldn't close the database.")
	}

} else {
    fmt.Println(err)
}

```
For more examples look in cblcgo_test.go.

## Testing

If you want to test the package you must build the C library and move or link the necesary files under the `include` directory. The produced `libCouchbaseLiteC` binary from the build should be placed in the root of the package. Once that's in place simply run `make tests` and it should produce the `.test` binaries. To run basic tests use the script `run_basic_tests.sh`. To run replicator tests you need to do the proper configuration of couchbase server and sync gateway. Once that's in place do `cblcgo-replicator.test -test.v`.
