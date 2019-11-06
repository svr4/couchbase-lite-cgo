all: couchbase-lite-cgo.test cblgo-replication.test

couchbase-lite-cgo.test: *.go
	rm -rf db/*.cblite2
	go build
	go test -c
	install_name_tool -change @rpath/libCouchbaseLiteC.dylib @loader_path/libCouchbaseLiteC.dylib couchbase-lite-cgo.test

cblgo-replication.test: *.go
	rm -rf db/*.cblite2
	go build
	go test -tags replication -c -o cblcgo-replication.test
	install_name_tool -change @rpath/libCouchbaseLiteC.dylib @loader_path/libCouchbaseLiteC.dylib cblcgo-replication.test