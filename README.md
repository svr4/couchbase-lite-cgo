# Couchbase Lite for Cgo

This project serves as the Go lang bindings for the couchbase-lite-C library developed by [Couchbase Labs](https://github.com/couchbaselabs).

It currently supports commit https://github.com/couchbaselabs/couchbase-lite-C/commit/e12e4a64911aa97be758e0957338c7d5b7201f5e of the library.

Replication is now working via `ws://` or `wss://`. Make sure to read the `replicator_test.go` file for more details on how to use replication properly with this library.

The package was built and tested on macOS (10.14.6+) using Go version `go1.12.9 darwin/amd64`.

## Example

1. Copy and paste the following code:

```go
package main


import (
	cblcgo "github.com/svr4/couchbase-lite-cgo"
	"fmt"
)

func main() {
	var config cblcgo.DatabaseConfiguration

	// Optional encryption key.
	var encryption_key cblcgo.EncryptionKey
	encryption_key.Algorithm = cblcgo.EncryptionNone
	encryption_key.Bytes = make([]byte, 0)

	config.Directory = "./"
	config.EncryptionKey = encryption_key
	config.Flags = cblcgo.Database_Create

	if db, err := cblcgo.Open("my_db", &config); err == nil {
		// Create a doc
		doc := cblcgo.NewDocumentWithId("test")
		doc.Props["name"] = "Luke"
		doc.Props["lastname"] = "Skywalker"
		doc.Props["age"] = 30
		doc.Props["email"] = "son.of.skywalker@gmail.com"
		doc.Props["action"] = "delete"

		if _, err2 := db.Save(doc, cblcgo.LastWriteWins); err2 == nil {
			// Retrieve the saved doc
			if savedDoc, e := db.GetMutableDocument("test"); e == nil {
				for k, v := range savedDoc.Props {
					if savedDoc.Props[k] != v {
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
	
}

```

2. In the root of your go project, run `go mod init $(pwd)`.

3. Run `go build`, it will fail. You need to copy or symlink the built CouchbaseLiteC binary, using the code up to the latest supported commit described above, into the mod directory under `$GOPATH/pkg/mod/github.com/svr4/couchbase-lite-cgo@vX.X.X/`.

4. For macOS only:
    Make sure the `@rpath` and the `@loader_path` are set to the location of the CouchbaseLiteC binary in your go binary, with the following command:
    `install_name_tool -change @rpath/libCouchbaseLiteC.dylib @loader_path/libCouchbaseLiteC.dylib your_go_binary`

5. Run `go build`.

For more examples look in cblcgo_test.go.

## Testing

If you want to test the package you must build the C library and move or link the necesary files under the `include` directory. The produced `libCouchbaseLiteC` binary from the build should be placed in the root of the package. Once that's in place simply run `make tests` and it should produce the `.test` binaries. To run basic tests use the script `run_basic_tests.sh`. To run replicator tests you need to do the proper configuration of couchbase server and sync gateway. Once that's in place do `cblcgo-replicator.test -test.v`.
