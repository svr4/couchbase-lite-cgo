tests: cblcgo.test cblcgo-replication.test

cblcgo.test: *.go
	#rm -rf db/*.cblite2
	go build
	go test -c -o cblcgo.test
	install_name_tool -change @rpath/libCouchbaseLiteC.dylib @loader_path/libCouchbaseLiteC.dylib cblcgo.test

cblcgo-replication.test: *.go
	#rm -rf db/*.cblite2
	go build
	go test -tags replication -c -o cblcgo-replication.test
	install_name_tool -change @rpath/libCouchbaseLiteC.dylib @loader_path/libCouchbaseLiteC.dylib cblcgo-replication.test