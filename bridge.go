package cblcgo
/*
#cgo LDFLAGS: -L. -lCouchbaseLiteC
#include <stdlib.h>
#include <stdio.h>
#include "include/CouchbaseLite.h"

void gatewayGoCallback(void *context, const CBLDatabase* db _cbl_nonnull, unsigned numDocs, const char **docIDs _cbl_nonnull);
void notificationReadyCallback(void *context, CBLDatabase* db _cbl_nonnull);

*/
import "C"
import "unsafe"
import "context"

//export listenerBridge
func listenerBridge(c unsafe.Pointer, db *C.CBLDatabase, numDocs C.unsigned, docIDs **C.char) {
	ctx := (*context.Context)(c)
	ids := make([]string, numDocs)

	var i, count_docs uint
	count_docs = uint(numDocs)
	for i=0; i < count_docs; i++ {
		ids[i] = C.GoString(*docIDs)	
	}

	database := Database{}
	database.db = db

	changeListener(*ctx, &database, ids)
}
//export notificationBridge
func notificationBridge(c unsafe.Pointer, db *C.CBLDatabase) {
	ctx := (*context.Context)(c)
	d := Database{}
	d.db = db
	notificationCallback(*ctx, &d)
}