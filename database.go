package cblcgo
/*
#cgo LDFLAGS: -L. -lCouchbaseLiteC
#include <stdlib.h>
#include <stdio.h>
#include "include/CouchbaseLite.h"

void databaseListenerBridge(void *, CBLDatabase*, unsigned, char **);
void notificationBridge(void *);

void gatewayDatabaseChangeGoCallback(void *context, const CBLDatabase* db _cbl_nonnull, unsigned numDocs, const char **docIDs _cbl_nonnull) {
	databaseListenerBridge(context, (CBLDatabase*)db, numDocs, (char**)docIDs);
}

void notificationReadyCallback(void *context, CBLDatabase* db _cbl_nonnull) {
	notificationBridge(context);
}
*/
import "C"
import "unsafe"
import "context"
import "fmt"

type EncryptionAlgorithm uint32
type DatabaseFlags uint32

const (
	EncryptionNone EncryptionAlgorithm = iota
	EncryptionAES256
)

/** Flags for how to open a database. */
const (
	Database_Create DatabaseFlags = (2 ^ iota) ///< Create the file if it doesn't exist
	Databse_ReadOnly ///< Open file read-only
	Database_NoUpgrade ///< Disable upgrading an older-version database
)

type Database struct {
	db *C.CBLDatabase
	config *C.CBLDatabaseConfiguration
	name string
}

type ListenerToken struct {
	token *C.CBLListenerToken
}


var uuid string = "UUID"
var callback string = "CALLBACK"
var pushCallback = "PUSHCALLBACK"
var pullCallback = "PULLCALLBACK"

/** Encryption key specified in a \ref CBLDatabaseConfiguration. */
type EncryptionKey struct {
    algorithm EncryptionAlgorithm ///< Encryption algorithm
    bytes []byte ///< Raw key data
}

/** Database configuration options. */
type DatabaseConfiguration struct {
    directory string                  ///< The parent directory of the database
    flags DatabaseFlags                 ///< Options for opening the database
    encryptionKey EncryptionKey;         ///< The database's encryption key (if any)
}

/** \name  Database file operations
    @{
    These functions operate on database files without opening them.
 */

/** Returns true if a database with the given name exists in the given directory.
    @param name  The database name (without the ".cblite2" extension.)
    @param inDirectory  The directory containing the database. If NULL, `name` must be an
                        absolute or relative path to the database. */
//bool CBL_DatabaseExists(const char* _cbl_nonnull name, const char *inDirectory) CBLAPI;
func DatabaseExists(name, inDirectory string) bool {
	c_name := C.CString(name)
	c_inDirectory := C.CString(inDirectory)
	result := C.CBL_DatabaseExists(c_name, c_inDirectory)
	C.free(unsafe.Pointer(c_name))
	C.free(unsafe.Pointer(c_inDirectory))
	return bool(result)
}

/** Copies a database file to a new location, and assigns it a new internal UUID to distinguish
	it from the original database when replicating.
	@param fromPath  The full filesystem path to the original database (including extension).
	@param toName  The new database name (without the ".cblite2" extension.)
	@param config  The database configuration (directory and encryption option.) */

// bool CBL_CopyDatabase(const char* _cbl_nonnull fromPath,
// 						const char* _cbl_nonnull toName,
// 						const CBLDatabaseConfiguration* config,
// 						CBLError*) CBLAPI;

func CopyDatabase(fromPath, toName string, config *DatabaseConfiguration) bool {
	c_fromPath := C.CString(fromPath)
	c_toName := C.CString(toName)
	c_dir := C.CString(config.directory)

	// need to check length and return false if diff or less than 32
	var key_data [32]C.uint8_t
	for i:=0; i < len(config.encryptionKey.bytes); i++ {
		key_data[i] = C.uint8_t(config.encryptionKey.bytes[i])
	}

	encryption_key := C.CBLEncryptionKey{C.uint32_t(config.encryptionKey.algorithm), key_data}

	c_config := (*C.CBLDatabaseConfiguration)(C.malloc(C.sizeof_CBLDatabaseConfiguration))
	c_config.directory = c_dir
	c_config.flags = C.uint32_t(config.flags)
	c_config.encryptionKey = encryption_key
	
	err := (*C.CBLError)(C.malloc(C.sizeof_CBLError))

	result := C.CBL_CopyDatabase(c_fromPath, c_toName, c_config, err)

	C.free(unsafe.Pointer(c_fromPath))
	C.free(unsafe.Pointer(c_toName))
	C.free(unsafe.Pointer(c_dir))
	C.free(unsafe.Pointer(err))

	return bool(result)
}

/** Deletes a database file. If the database file is open, an error is returned.
	@param name  The database name (without the ".cblite2" extension.)
	@param inDirectory  The directory containing the database. If NULL, `name` must be an
						absolute or relative path to the database.
	@param outError  On return, will be set to the error that occurred, or a 0 code if no error.
		@return  True if the database was deleted, false if it doesn't exist or deletion failed.
				(You can tell the last two cases apart by looking at \ref outError.)*/

// bool CBL_DeleteDatabase(const char _cbl_nonnull *name, 
// 						const char *inDirectory,
// 						CBLError *outError) CBLAPI;

func DeleteDatabase(name, inDirectory string) bool {

	c_name := C.CString(name)
	c_inDirectory := C.CString(inDirectory)
	err := (*C.CBLError)(C.malloc(C.sizeof_CBLError))


	result := C.CBL_DeleteDatabase(c_name, c_inDirectory, err)
	C.free(unsafe.Pointer(c_name))
	C.free(unsafe.Pointer(c_inDirectory))
	C.free(unsafe.Pointer(err))
	return bool(result)
}

/** \name  Database lifecycle
    @{
    Opening, closing, and managing open databases.
 */

/** Opens a database, or creates it if it doesn't exist yet, returning a new \ref CBLDatabase
    instance.
    It's OK to open the same database file multiple times. Each \ref CBLDatabase instance is
    independent of the others (and must be separately closed and released.)
    @param name  The database name (without the ".cblite2" extension.)
    @param config  The database configuration (directory and encryption option.)
    @param error  On failure, the error will be written here.
	@return  The new database object, or NULL on failure. */

	// _cbl_warn_unused
	// CBLDatabase* CBLDatabase_Open(const char *name _cbl_nonnull,
	// 							  const CBLDatabaseConfiguration* config,
	// 							  CBLError* error) CBLAPI;

func Open(name string, config *DatabaseConfiguration) (*Database, error) {

	c_name := C.CString(name)
	defer C.free(unsafe.Pointer(c_name))
	// Convert to C array
	var key_data [32]C.uint8_t
	for i:=0; i < len(config.encryptionKey.bytes); i++ {
		key_data[i] = C.uint8_t(config.encryptionKey.bytes[i])
	}
	// Create C key
	c_key := C.CBLEncryptionKey{C.uint32_t(config.encryptionKey.algorithm), key_data}
	// Create C config
	c_config := (*C.CBLDatabaseConfiguration)(C.malloc(C.sizeof_CBLDatabaseConfiguration))

	c_dir := C.CString(config.directory)
	defer C.free(unsafe.Pointer(c_dir))

	c_config.directory = c_dir
	c_config.flags = C.uint32_t(config.flags)
	c_config.encryptionKey = c_key
	
	err := (*C.CBLError)(C.malloc(C.sizeof_CBLError))
	defer C.free(unsafe.Pointer(err))

	// Open Database
	c_db := C.CBLDatabase_Open(c_name, c_config, err)

	if (*err).code == 0 {
		database := Database{}
		database.db = c_db
		database.config = c_config
		database.name = name
		return &database, nil
	}

	ErrCBLInternalError = fmt.Errorf("CBL: Problem Opening Database. Domain: %d Code: %d", (*err).domain, (*err).code)
	return nil, ErrCBLInternalError
}


/** Closes an open database. */
// bool CBLDatabase_Close(CBLDatabase*, CBLError*) CBLAPI;
func (db *Database) Close() bool {
	err := (*C.CBLError)(C.malloc(C.sizeof_CBLError))
	defer C.free(unsafe.Pointer(err))
	result := C.CBLDatabase_Close(db.db, err)
	C.free(unsafe.Pointer(db.config))
	return bool(result)
}


/** Closes and deletes a database. If there are any other connections to the database,
	an error is returned. */
// bool CBLDatabase_Delete(CBLDatabase* _cbl_nonnull, CBLError*) CBLAPI;
func (db *Database) Delete() bool {
	err := (*C.CBLError)(C.malloc(C.sizeof_CBLError))
	defer C.free(unsafe.Pointer(err))
	result := C.CBLDatabase_Delete(db.db, err)
	if (*err).code == 0 {
		return bool(result)
	}
	return false
}

/** Compacts a database file. */
// bool CBLDatabase_Compact(CBLDatabase* _cbl_nonnull, CBLError*) CBLAPI;
func (db *Database) Compact() bool {
	err := (*C.CBLError)(C.malloc(C.sizeof_CBLError))
	defer C.free(unsafe.Pointer(err))
	result := C.CBLDatabase_Compact(db.db, err)
	if (*err).code == 0 {
		return bool(result)
	}
	return false
}

/** Begins a batch operation, similar to a transaction. You **must** later call \ref
	CBLDatabase_EndBatch to end (commit) the batch.
	@note  Multiple writes are much faster when grouped inside a single batch.
	@note  Changes will not be visible to other CBLDatabase instances on the same database until
			the batch operation ends.
	@note  Batch operations can nest. Changes are not committed until the outer batch ends. */
// bool CBLDatabase_BeginBatch(CBLDatabase* _cbl_nonnull, CBLError*) CBLAPI;
func (db *Database) BeginBatch() bool {
	err := (*C.CBLError)(C.malloc(C.sizeof_CBLError))
	defer C.free(unsafe.Pointer(err))
	result := C.CBLDatabase_BeginBatch(db.db, err)
	if (*err).code == 0 {
		return bool(result)
	}
	return false
}

/** Ends a batch operation. This **must** be called after \ref CBLDatabase_BeginBatch. */
// bool CBLDatabase_EndBatch(CBLDatabase* _cbl_nonnull, CBLError*) CBLAPI;
func (db *Database) EndBatch() bool {
	err := (*C.CBLError)(C.malloc(C.sizeof_CBLError))
	defer C.free(unsafe.Pointer(err))
	result := C.CBLDatabase_EndBatch(db.db, err)
	if (*err).code == 0 {
		return bool(result)
	}
	return false
}


/** Returns the nearest future time at which a document in this database will expire,
	or 0 if no documents will expire. */

// CBLTimestamp CBLDatabase_NextDocExpiration(CBLDatabase* _cbl_nonnull) CBLAPI;
// int64_t
func (db *Database) NextDocExpiration() int64 {
	timestamp := C.CBLDatabase_NextDocExpiration(db.db)
	return int64(timestamp)
}

/** Purges all documents whose expiration time has passed.
	@param db  The database to purge
	@param error  On failure, the error will be written here.
	@return  The number of documents purged, or -1 on error. */

// int64_t CBLDatabase_PurgeExpiredDocuments(CBLDatabase* db _cbl_nonnull,
// 										  CBLError* error) CBLAPI;
func (db *Database) PurgeExpiredDocuments() int64 {
	err := (*C.CBLError)(C.malloc(C.sizeof_CBLError))
	defer C.free(unsafe.Pointer(err))
	result := C.CBLDatabase_PurgeExpiredDocuments(db.db, err)
	if (*err).code == 0 {
		return int64(result)
	}
	return -1
}
/** @} */

/** \name  Database accessors
    @{
    Getting information about a database.
 */

/** Returns the database's name. */
// const char* CBLDatabase_Name(const CBLDatabase* _cbl_nonnull) CBLAPI _cbl_returns_nonnull;
func (db *Database) DatabaseName() string {
	c_name := C.CBLDatabase_Name(db.db)
	name := C.GoString(c_name)
	C.free(unsafe.Pointer(c_name))
	return name
}

/** Returns the database's full filesystem path. */
// const char* CBLDatabase_Path(const CBLDatabase* _cbl_nonnull) CBLAPI _cbl_returns_nonnull;
func (db *Database) Path() string {
	c_path := C.CBLDatabase_Path(db.db)
	path := C.GoString(c_path)
	C.free(unsafe.Pointer(c_path))
	return path
}

/** Returns the number of documents in the database. */
// uint64_t CBLDatabase_Count(const CBLDatabase* _cbl_nonnull) CBLAPI;
func (db *Database) Count() uint64 {
	c_count := C.CBLDatabase_Count(db.db)
	return uint64(c_count)
}

/** Returns the database's configuration, as given when it was opened.
    @note  The encryption key is not filled in, for security reasons. */
// const CBLDatabaseConfiguration CBLDatabase_Config(const CBLDatabase* _cbl_nonnull) CBLAPI;
func (db *Database) DatabaseConfig() *DatabaseConfiguration {
	c_config := C.CBLDatabase_Config(db.db)
	config := DatabaseConfiguration{}
	key := EncryptionKey{}

	dir := C.GoString(c_config.directory)
	flags := DatabaseFlags(c_config.flags)
	key.algorithm = EncryptionAlgorithm(c_config.encryptionKey.algorithm)
	key.bytes = make([]byte, 0)
	
	// setup go config
	config.directory = dir
	config.encryptionKey = key
	config.flags = flags

	return &config
}

/** \name  Database listeners
    @{
    A database change listener lets you detect changes made to all documents in a database.
    (If you only want to observe specific documents, use a \ref CBLDocumentChangeListener instead.)
    @note If there are multiple \ref CBLDatabase instances on the same database file, each one's
    listeners will be notified of changes made by other database instances.
    @warning  Changes made to the database file by other processes will _not_ be notified. */

/** A database change listener callback, invoked after one or more documents are changed on disk.
    @warning  By default, this listener may be called on arbitrary threads. If your code isn't
                    prepared for that, you may want to use \ref CBLDatabase_BufferNotifications
                    so that listeners will be called in a safe context.
    @param context  An arbitrary value given when the callback was registered.
    @param db  The database that changed.
    @param numDocs  The number of documents that changed (size of the `docIDs` array)
	@param docIDs  The IDs of the documents that changed, as a C array of `numDocs` C strings. */
	
    // typedef void (*CBLDatabaseChangeListener)(void *context,
	// 	const CBLDatabase* db _cbl_nonnull,
	// 	unsigned numDocs,
	// 	const char **docIDs _cbl_nonnull);
type DatabaseChangeListener func(ctx context.Context, db *Database, docIDs []string)

/** Registers a database change listener callback. It will be called after one or more
documents are changed on disk.
@param db  The database to observe.
@param listener  The callback to be invoked.
@param context  An opaque value that will be passed to the callback.
@return  A token to be passed to \ref CBLListener_Remove when it's time to remove the
listener.*/

// _cbl_warn_unused
// CBLListenerToken* CBLDatabase_AddChangeListener(const CBLDatabase* db _cbl_nonnull,
// 		  CBLDatabaseChangeListener listener _cbl_nonnull,
// 		  void *context) CBLAPI;
func (db *Database) AddDatabaseChangeListener(listener DatabaseChangeListener, ctx context.Context) *ListenerToken {
	ctx = context.WithValue(ctx, callback, listener)
	token := C.CBLDatabase_AddChangeListener(db.db, (C.CBLDatabaseChangeListener)(C.gatewayDatabaseChangeGoCallback), unsafe.Pointer(&ctx))
	listener_token := ListenerToken{token}
	return &listener_token
}
/** @} */
/** @} */    // end of outer \defgroup

/** \defgroup listeners   Listeners
    @{ */
/** \name  Scheduling notifications
    @{
    Applications may want control over when Couchbase Lite notifications (listener callbacks)
    happen. They may want them called on a specific thread, or at certain times during an event
    loop. This behavior may vary by database, if for instance each database is associated with a
    separate thread.

    The API calls here enable this. When notifications are "buffered" for a database, calls to
    listeners will be deferred until the application explicitly allows them. Instead, a single
    callback will be issued when the first notification becomes available; this gives the app a
    chance to schedule a time when the notifications should be sent and callbacks called.
 */

/** Callback indicating that the database (or an object belonging to it) is ready to call one
    or more listeners. You should call \ref CBLDatabase_SendNotifications at your earliest
    convenience, in the context (thread, dispatch queue, etc.) you want them to run.
    @note  This callback is called _only once_ until the next time \ref CBLDatabase_SendNotifications
            is called. If you don't respond by (sooner or later) calling that function,
            you will not be informed that any listeners are ready.
    @warning  This can be called from arbitrary threads. It should do as little work as
			  possible, just scheduling a future call to \ref CBLDatabase_SendNotifications. */
			  
			//   typedef void (*CBLNotificationsReadyCallback)(void *context,
			// 	CBLDatabase* db _cbl_nonnull);
type NotificationsReadyCallback func (ctx context.Context, db *Database)

/** Switches the database to buffered-notification mode. Notifications for objects belonging
to this database (documents, queries, replicators, and of course the database) will not be
called immediately; your \ref CBLNotificationsReadyCallback will be called instead.
@param db  The database whose notifications are to be buffered.
@param callback  The function to be called when a notification is available.
@param context  An arbitrary value that will be passed to the callback. */

// void CBLDatabase_BufferNotifications(CBLDatabase *db _cbl_nonnull,
// 	   CBLNotificationsReadyCallback callback _cbl_nonnull,
// 	   void *context) CBLAPI;
func (db *Database) DatabaseBufferNotifications(callback NotificationsReadyCallback, ctx context.Context) {
	ctx = context.WithValue(ctx, callback, callback)
	C.CBLDatabase_BufferNotifications(db.db, 
		(C.CBLNotificationsReadyCallback)(C.notificationReadyCallback),
		unsafe.Pointer(&ctx))
}

/** Immediately issues all pending notifications for this database, by calling their listener
callbacks. */

// void CBLDatabase_SendNotifications(CBLDatabase *db _cbl_nonnull) CBLAPI;
func (db *Database) SendNotifications() {
	C.CBLDatabase_SendNotifications(db.db)
}
/*
	Removes a listener callback, given the token that was returned when it was added.
*/
func RemoveListener(token *ListenerToken) {
	C.CBLListener_Remove(token.token)
}
	   
/** @} */
/** @} */    // end of outer \defgroup