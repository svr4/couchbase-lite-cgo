package cblcgo
/*
#cgo LDFLAGS: -L. -lCouchbaseLiteC
#include <stdlib.h>
#include <stdio.h>
#include "include/CouchbaseLite.h"

void gatewayDatabaseChangeGoCallback(void *context, const CBLDatabase* db _cbl_nonnull, unsigned numDocs, const char **docIDs _cbl_nonnull);
void gatewayDocumentChangeGoCallback(void *context, const CBLDatabase* db _cbl_nonnull, const char *docID _cbl_nonnull);
void notificationReadyCallback(void *context, CBLDatabase* db _cbl_nonnull);
FLValue FLArray_AsValue(FLArray);
FLValue FLDict_AsValue(FLDict);

*/
import "C"
import "unsafe"
import "context"

//export databaseListenerBridge
func databaseListenerBridge(c unsafe.Pointer, db *C.CBLDatabase, numDocs C.unsigned, docIDs **C.char) {
	ctx := (*context.Context)(c)
	ids := make([]string, numDocs)

	var i, count_docs uint
	count_docs = uint(numDocs)
	for i=0; i < count_docs; i++ {
		ids[i] = C.GoString(*docIDs)	
	}

	database := Database{}
	database.db = db

	v := (*ctx).Value(uuid).(string)
	(databaseChangeListeners[v])(*ctx, &database, ids)
}
//export documentListenerBridge
func documentListenerBridge(c unsafe.Pointer, db *C.CBLDatabase, c_docID *C.char) {
	ctx := (*context.Context)(c)
	docId := C.GoString(c_docID)
	database := Database{}
	database.db = db
	v := (*ctx).Value(uuid).(string)
	(documentChangeListeners[v])(*ctx, &database, docId)
}
//export notificationBridge
func notificationBridge(c unsafe.Pointer, db *C.CBLDatabase) {
	ctx := (*context.Context)(c)
	d := Database{}
	d.db = db
	notificationCallback(*ctx, &d)
}