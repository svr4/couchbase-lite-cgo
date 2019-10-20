couchbase-lite-cgo.test: *.go
	go build
	go test -c
	install_name_tool -change @rpath/libCouchbaseLiteC.dylib @loader_path/libCouchbaseLiteC.dylib couchbase-lite-cgo.test
	rm -rf db/*.cblite2