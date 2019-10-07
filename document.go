package cblcgo
/*
#cgo LDFLAGS: -L. -lCouchbaseLiteC
#include <stdlib.h>
#include <stdio.h>
#include "include/CouchbaseLite.h"

void documentListenerBridge(void *, CBLDatabase *, char *);

void gatewayDocumentChangeGoCallback(void *context, const CBLDatabase* db _cbl_nonnull, const char *docID _cbl_nonnull) {
	documentListenerBridge(context, (CBLDatabase*)db, (char*)docID);
}

FLValue FLArray_AsValue(FLArray arr) {
	if(arr != NULL) {
		return (FLValue)arr;
	}
	return NULL;
}

FLValue FLDict_AsValue(FLDict dict) {
	if(dict != NULL) {
		return (FLValue)dict;
	}
	return NULL;
}

*/
import "C"
import "unsafe"
import "fmt"
import "reflect"
import "context"

/** \defgroup documents   Documents
    @{
    A \ref CBLDocument is essentially a JSON object with an ID string that's unique in its database.
 */

type Document struct {
	doc *C.CBLDocument
	ReadOnly bool
	Props map[string]interface{}
	data []byte
	Keys []string
}

/** \name  Document lifecycle
    @{ */

/** Conflict-handling options when saving or deleting a document. */
type ConcurrencyControl uint8

const (
	/** The current save/delete will overwrite a conflicting revision if there is a conflict. */
	ConcurrencyControlLastWriteWins ConcurrencyControl = iota
	/** The current save/delete will fail if there is a conflict. */
	ConcurrencyControlFailOnConflict
)


var documentChangeListeners map[string]DocumentChangeListener

/** Reads a document from the database, creating a new (immutable) \ref CBLDocument object.
    Each call to this function creates a new object (which must later be released.)
    @note  If you are reading the document in order to make changes to it, call
            \ref CBLDatabase_GetMutableDocument instead.
    @param database  The database.
    @param docID  The ID of the document.
    @return  A new \ref CBLDocument instance, or NULL if no document with that ID exists. */
// _cbl_warn_unused
// const CBLDocument* CBLDatabase_GetDocument(const CBLDatabase* database _cbl_nonnull,
//                                            const char* _cbl_nonnull docID) CBLAPI;
func (db *Database) GetReadOnlyDocument(docId string) (*Document, error){
	c_docId := C.CString(docId)
	document := C.CBLDatabase_GetDocument(db.db, c_docId)
	C.free(unsafe.Pointer(c_docId))
	if document == nil {
		return nil, ErrProblemReadingDocument
	}
	doc := Document{}
	doc.doc = document
	doc.ReadOnly = true
	documentProperties(&doc)
	return &doc, nil
}

/** Saves a (mutable) document to the database.
    @param db  The database to save to.
    @param doc  The mutable document to save.
    @param concurrency  Conflict-handling strategy.
    @param error  On failure, the error will be written here.
    @return  An updated document reflecting the saved changes, or NULL on failure. */
// _cbl_warn_unused
// const CBLDocument* CBLDatabase_SaveDocument(CBLDatabase* db _cbl_nonnull,
//                                             CBLDocument* doc _cbl_nonnull,
//                                             CBLConcurrencyControl concurrency,
//                                             CBLError* error) CBLAPI;
func (db *Database) Save(doc *Document, concurrency ConcurrencyControl) (*Document, error) {
	if doc.ReadOnly {
		return nil, ErrDocumentIsNotReadOnly
	}
	err := C.CBLError{}
	saved_doc := C.CBLDatabase_SaveDocument(db.db, doc.doc, C.CBLConcurrencyControl(concurrency), &err)
	if err.code == 0 {
		doc.doc = saved_doc
		return doc, nil
	}
	return nil, ErrProblemSavingDocument
}


/** Deletes a document from the database. Deletions are replicated.
    @warning  You are still responsible for releasing the CBLDocument.
    @param document  The document to delete.
    @param concurrency  Conflict-handling strategy.
    @param error  On failure, the error will be written here.
    @return  True if the document was deleted, false if an error occurred. */
// bool CBLDocument_Delete(const CBLDocument* document _cbl_nonnull,
//                         CBLConcurrencyControl concurrency,
//                         CBLError* error) CBLAPI;
func (db *Database) DeleteDocument(doc *Document, concurrency ConcurrencyControl) error {
	err := C.CBLError{}
	result := bool(C.CBLDocument_Delete(doc.doc, C.CBLConcurrencyControl(concurrency), &err))
	if result && err.code == 0{
		C.free(unsafe.Pointer(doc.doc))
		return nil
	}
	ErrCBLInternalError = fmt.Errorf("CBL: Problem Deleting Document. Domain: %d Code: %d", err.domain, err.code)
	return ErrCBLInternalError
}

/** Purges a document. This removes all traces of the document from the database.
    Purges are _not_ replicated. If the document is changed on a server, it will be re-created
    when pulled.
    @warning  You are still responsible for releasing the \ref CBLDocument reference.
    @note If you don't have the document in memory already, \ref CBLDatabase_PurgeDocumentByID is a
          simpler shortcut.
    @param document  The document to delete.
    @param error  On failure, the error will be written here.
    @return  True if the document was purged, false if it doesn't exist or the purge failed. */
// bool CBLDocument_Purge(const CBLDocument* document _cbl_nonnull,
//                        CBLError* error) CBLAPI;
func (db *Database) Purge(doc *Document) error {
	err := C.CBLError{}
	result := bool(C.CBLDocument_Purge(doc.doc, &err))
	if result && err.code == 0{
		C.free(unsafe.Pointer(doc.doc))
		return nil
	}
	ErrCBLInternalError = fmt.Errorf("CBL: Problem Purging Document. Domain: %d Code: %d", err.domain, err.code)
	return ErrCBLInternalError
}

/** Purges a document, given only its ID.
    @note  If no document with that ID exists, this function will return false but the error
            code will be zero.
    @param database  The database.
    @param docID  The document ID to purge.
    @param error  On failure, the error will be written here.
    @return  True if the document was purged, false if it doesn't exist or the purge failed.
 */
// bool CBLDatabase_PurgeDocumentByID(CBLDatabase* database _cbl_nonnull,
//                                   const char* docID _cbl_nonnull,
//                                   CBLError* error) CBLAPI;
func (db *Database) PurgeById(docId string) error {
	err := C.CBLError{}
	c_docId := C.CString(docId)
	result := bool(C.CBLDatabase_PurgeDocumentByID(db.db, c_docId, &err))
	if result && err.code == 0{
		C.free(unsafe.Pointer(c_docId))
		return nil
	}
	ErrCBLInternalError = fmt.Errorf("CBL: Problem Purging Document. Domain: %d Code: %d", err.domain, err.code)
	return ErrCBLInternalError
}

/** @} */



/** \name  Mutable documents
    @{
    The type `CBLDocument*` without a `const` qualifier refers to a _mutable_ document instance.
    A mutable document exposes its properties as a mutable dictionary, so you can change them
    in place and then call \ref CBLDatabase_SaveDocument to persist the changes.
 */

/** Reads a document from the database, in mutable form that can be updated and saved.
    (This function is otherwise identical to \ref CBLDatabase_GetDocument.)
    @note  You must release the document when you're done with it.
    @param database  The database.
    @param docID  The ID of the document.
    @return  A new mutable CBLDocument instance, or NULL if no document with that ID exists. */
// _cbl_warn_unused
// CBLDocument* CBLDatabase_GetMutableDocument(CBLDatabase* database _cbl_nonnull,
//                                             const char* docID _cbl_nonnull) CBLAPI;
func (db *Database) GetMutableDocument(docId string) (*Document, error) {
	c_docId := C.CString(docId)
	c_doc := C.CBLDatabase_GetMutableDocument(db.db, c_docId)
	C.free(unsafe.Pointer(c_docId))
	if c_doc == nil {
		return nil, ErrProblemReadingDocument
	}
	document := Document{}
	document.doc = c_doc
	document.ReadOnly = false
	documentProperties(&document)
	return &document, nil
}

/** Creates a new, empty document in memory. It will not be added to a database until saved.
    @param docID  The ID of the new document, or NULL to assign a new unique ID.
    @return  The mutable document instance. */
//CBLDocument* CBLDocument_New(const char *docID) CBLAPI _cbl_warn_unused _cbl_returns_nonnull;
func NewDocument() *Document {
	doc := C.CBLDocument_New(nil)
	document := Document{}
	document.doc = doc
	document.ReadOnly = false
	document.Keys = make([]string, 0)
	document.Props = make(map[string]interface{})
	return &document
}

func NewDocumentWithId(docId string) *Document {
	c_docId := C.CString(docId)
	doc := C.CBLDocument_New(c_docId)
	C.free(unsafe.Pointer(c_docId))
	document := Document{}
	document.doc = doc
	document.ReadOnly = false
	document.Keys = make([]string, 0)
	document.Props = make(map[string]interface{})
	return &document
}

/** Creates a new mutable CBLDocument instance that refers to the same document as the original.
    If the original document has unsaved changes, the new one will also start out with the same
    changes; but mutating one document thereafter will not affect the other.
    @note  You must release the new reference when you're done with it. */
// CBLDocument* CBLDocument_MutableCopy(const CBLDocument* original _cbl_nonnull) CBLAPI
//     _cbl_warn_unused _cbl_returns_nonnull;
func DocumentMutableCopy(original *Document) *Document {
	copy := Document{}
	copy.ReadOnly = false
	copy.doc = C.CBLDocument_MutableCopy(original.doc)
	copy.Keys = original.Keys
	copy.Props = original.Props
	return &copy
}

/** @} */



/** \name  Document properties and metadata
    @{
    A document's body is essentially a JSON object. The properties are accessed in memory
    using the Fleece API, with the body itself being a \ref FLDict "dictionary").
 */

/** Returns a document's ID. */
// const char* CBLDocument_ID(const CBLDocument* _cbl_nonnull) CBLAPI _cbl_returns_nonnull;
func (doc *Document) Id() string {
	c_id := C.CBLDocument_ID(doc.doc)
	id := C.GoString(c_id)
	C.free(unsafe.Pointer(c_id))
	return id
}

/** Returns a document's current sequence in the local database.
    This number increases every time the document is saved, and a more recently saved document
    will have a greater sequence number than one saved earlier, so sequences may be used as an
    abstract 'clock' to tell relative modification times. */
// uint64_t CBLDocument_Sequence(const CBLDocument* _cbl_nonnull) CBLAPI;
func (doc *Document) DocumentSequence() uint64 {
	return uint64(C.CBLDocument_Sequence(doc.doc))
}

/** Returns a document's properties as a dictionary.
    @note  The dictionary object is owned by the document; you do not need to release it.
    @warning  This dictionary _reference_ is immutable, but if the document is mutable the
           underlying dictionary itself is mutable and could be modified through a mutable
           reference obtained via \ref CBLDocument_MutableProperties. If you need to preserve the
           properties, call \ref FLDict_MutableCopy to make a deep copy. */
// FLDict CBLDocument_Properties(const CBLDocument* _cbl_nonnull) CBLAPI;
func documentProperties(doc *Document) error {
	if !doc.ReadOnly {
		return ErrDocumentIsNotReadOnly
	}
	var err error
	fl_dict := C.CBLDocument_Properties(doc.doc)
	// Got the props, now I need to move them to the Props property
	doc.Keys = getDocumentKeys(doc)
	doc.Props, err = getKeyValuePropMap(fl_dict)
	if err != nil {
		return err
	}
	return nil
}

func getDocumentKeys(doc *Document) []string {
	fl_dict := C.CBLDocument_Properties(doc.doc)
	return getDocumentKeysHelper(fl_dict)
}

func getKeyValuePropMap(fl_dict C.FLDict) (map[string]interface{}, error) {

	iter := C.FLDictIterator{}
	C.FLDictIterator_Begin(fl_dict, &iter)
	var value C.FLValue

	props := make(map[string]interface{})

	for value = C.FLDictIterator_GetValue(&iter); value != nil; value = C.FLDictIterator_GetValue(&iter) {
		// FLString
		key := C.FLDictIterator_GetKeyString(&iter)
		str_key := C.GoStringN((*C.char)(key.buf), C.int(key.size))
		
		i, e := getFLValueToGoValue(value)

		if e == nil {
			props[str_key] = i
		}

		C.FLDictIterator_Next(&iter)
	}

	return props, nil
}

func getDocumentKeysHelper(fl_dict C.FLDict) []string {
	iter := C.FLDictIterator{}
	C.FLDictIterator_Begin(fl_dict, &iter)
	var value C.FLValue

	keys := make([]string, int(C.FLDictIterator_GetCount(&iter)))
	i := 0
	for value = C.FLDictIterator_GetValue(&iter); value != nil; value = C.FLDictIterator_GetValue(&iter) {
		// FLString
		key := C.FLDictIterator_GetKeyString(&iter);
		keys[i] = C.GoStringN((*C.char)(key.buf), C.int(key.size))
		i++
		C.FLDictIterator_Next(&iter)
	}

	return keys
}

func getFLValueToGoValue(fl_val C.FLValue) (interface{}, error) {
	var val interface{}
	switch C.FLValue_GetType(fl_val) {
		///< Type of a NULL pointer, i.e. no such value, like JSON `undefined`. Also the type of a value created by FLEncoder_WriteUndefined().
		case C.kFLUndefined:
		///< Equivalent to a JSON 'null'
		case C.kFLNull:
			val = nil
			return val, nil
		case C.kFLBoolean:
			val = bool(C.FLValue_AsBool(fl_val))
			return val, nil
		///< A numeric value, either integer or floating-point
		case C.kFLNumber:
			if C.FLValue_IsInteger(fl_val) {
				val = int64(C.FLValue_AsInt(fl_val))
			} else if C.FLValue_IsUnsigned(fl_val) {
				val = uint64(C.FLValue_AsUnsigned(fl_val))
			} else if C.FLValue_IsDouble(fl_val) {
				val = float64(C.FLValue_AsDouble(fl_val))
			} else {
				val = float32(C.FLValue_AsFloat(fl_val))
			}
			return val, nil
		case C.kFLString:
			fl_str := C.FLValue_AsString(fl_val)
			val = C.GoString((*C.char)(fl_str.buf))
			return val, nil
		case C.kFLData:
			fl_data_slice := C.FLValue_AsData(fl_val)
			val = C.GoBytes(fl_data_slice.buf, C.int(fl_data_slice.size))
			return val, nil
		case C.kFLArray:
			// This could be a homogenous array or a hetero one.
			// Return the bytes and let the developer decide.
			// Hetero arrays can be processed by converting []byte to []interface{}
			fl_array, err := C.FLValue_AsArray(fl_val)
			if err == nil {
				is_empty := bool(C.FLArray_IsEmpty(fl_array))
				if !is_empty {
					val = C.GoBytes(unsafe.Pointer(fl_array), C.int(C.FLArray_Count(fl_array)))
				} else {
					val = nil
				}
			}
			return val, nil
		case C.kFLDict:
			// Determine if dictionary is a Blob
			if isBlob(C.FLValue_AsDict(fl_val)) {
				if blob, err := getBlob(C.FLValue_AsDict(fl_val)); err == nil {
					return blob, nil
				}
				return nil, ErrProblemGettingBlobWithData
			}
			// Deep iterate over the value which is a map
			iter := C.FLDeepIterator_New(fl_val)
			var value C.FLValue

			props := make(map[string]interface{})

			for value = C.FLDeepIterator_GetValue(iter); value != nil; value = C.FLDeepIterator_GetValue(iter) {
				// FLString
				key := C.FLDeepIterator_GetKey(iter)
				str_key := C.GoStringN((*C.char)(key.buf), C.int(key.size))
				
				i, e := getFLValueToGoValue(value)

				if e == nil {
					props[str_key] = i
				}

				C.FLDeepIterator_Next(iter)
			}
			C.FLDeepIterator_Free(iter)
			return props, nil
		default:
			return nil, ErrInvalidCBLType

		}
		return nil, ErrInvalidCBLType
}

func storeGoValueInSlot(fl_slot C.FLSlot, v interface{}) error {

	switch val := reflect.TypeOf(v); val.Kind() {
	case reflect.String:
		value := v.(string)
		s := C.CString(value)
		// Create a key in the dict. Returns an FLSlot, and set the slot with the value
		C.FLSlot_SetString(fl_slot, C.FLStr(s))
		C.free(unsafe.Pointer(s))
		break;
	case reflect.Array:
		return ErrUnsupportedGoType
	case reflect.Slice:
	//case []byte:
		// We have to iterate through the array.
		mutable_array := C.FLMutableArray_New()
		v_arr := v.([]interface{})
		for i:=0; i < len(v_arr); i++ {
			v_slot := C.FLMutableArray_Append(mutable_array)
			storeGoValueInSlot(v_slot, v_arr[i]);
		}
		fl_arr := C.FLMutableArray_GetSource(mutable_array)
		C.FLSlot_SetValue(fl_slot, C.FLArray_AsValue(fl_arr))
		break;
	case reflect.Int:
		value := v.(int)
		C.FLSlot_SetInt(fl_slot, C.int64_t(value))
		break;
	case reflect.Int8:
		value := v.(int8)
		C.FLSlot_SetInt(fl_slot, C.int64_t(value))
		break;
	case reflect.Int16:
		value := v.(int16)
		C.FLSlot_SetInt(fl_slot, C.int64_t(value))
		break;
	case reflect.Int32:
		value := v.(int32)
		C.FLSlot_SetInt(fl_slot, C.int64_t(value))
		break;
	case reflect.Int64:
		value := v.(int64)
		C.FLSlot_SetInt(fl_slot, C.int64_t(value))
		break;
	case reflect.Uint:
		value := v.(uint)
		C.FLSlot_SetUInt(fl_slot, C.uint64_t(value))
		break;
	case reflect.Uintptr:
		value := v.(uintptr)
		C.FLSlot_SetUInt(fl_slot, C.uint64_t(value))
		break;
	case reflect.Uint8:
		value := v.(uint8)
		C.FLSlot_SetUInt(fl_slot, C.uint64_t(value))
		break;
	case reflect.Uint16:
		value := v.(uint16)
		C.FLSlot_SetUInt(fl_slot, C.uint64_t(value))
		break;
	case reflect.Uint32:
		value := v.(uint32)
		C.FLSlot_SetUInt(fl_slot, C.uint64_t(value))
		break;
	case reflect.Uint64:
		value := v.(uint64)
		C.FLSlot_SetUInt(fl_slot, C.uint64_t(value))
		break;
	case reflect.Float32:
		value := v.(float32)
		C.FLSlot_SetFloat(fl_slot, C.float(value))
		break;
	case reflect.Float64:
		value := v.(float64)
		C.FLSlot_SetDouble(fl_slot, C.double(value))
		// double
		break;
	case reflect.Bool:
		value := v.(bool)
		C.FLSlot_SetBool(fl_slot, C.bool(value))
		break;
	case reflect.Map:
		switch v.(type) {
		case map[string]interface{}:
			v_map := v.(map[string]interface{})
			mutable_dict := C.FLMutableDict_New()

			for key, val := range v_map {
				c_key := C.CString(key)
				v_slot := C.FLMutableDict_Set(mutable_dict, C.FLStr(c_key))
				storeGoValueInSlot(v_slot, val)
				C.free(unsafe.Pointer(c_key))
			}

			fl_dict := C.FLMutableDict_GetSource(mutable_dict)
			C.FLSlot_SetValue(fl_slot, C.FLDict_AsValue(fl_dict))
			break;
		default:
			return ErrUnsupportedGoType
		}
	default:
		return ErrUnsupportedGoType
	}
	return nil
}

/** Returns a mutable document's properties as a mutable dictionary.
    You may modify this dictionary and then call \ref CBLDatabase_SaveDocument to persist the changes.
    @note  The dictionary object is owned by the document; you do not need to release it.
    @note  Every call to this function returns the same mutable collection. This is the
           same collection returned by \ref CBLDocument_Properties. */
// FLMutableDict CBLDocument_MutableProperties(CBLDocument* _cbl_nonnull) CBLAPI _cbl_returns_nonnull;

/** Sets a mutable document's properties.
    Call \ref CBLDatabase_SaveDocument to persist the changes.
    @note  The dictionary object will be retained by the document. You are responsible for
           releasing your own reference(s) to it. */
// void CBLDocument_SetProperties(CBLDocument* _cbl_nonnull,
                            //    FLMutableDict properties _cbl_nonnull) CBLAPI;
func setProperties(doc *Document) {

}
// FLDoc CBLDocument_CreateFleeceDoc(const CBLDocument* _cbl_nonnull) CBLAPI;

/** Returns a document's properties as a null-terminated JSON string.
    @note You are responsible for calling `free()` on the returned string. */
// char* CBLDocument_PropertiesAsJSON(const CBLDocument* _cbl_nonnull) CBLAPI _cbl_returns_nonnull; 
func (doc *Document) ToJSONString() string {
	// We need to sync up the map with the underlying FLDict
	syncMapToUnderlyingDict(doc)
	c_json := C.CBLDocument_PropertiesAsJSON(doc.doc)
	json := C.GoString(c_json)
	C.free(unsafe.Pointer(c_json))
	return json
}

func syncMapToUnderlyingDict(doc *Document) bool {
	if doc.ReadOnly {
		return false
	}

	mutableDict := C.FLMutableDict_New()

	for k, v := range doc.Props {
		c_key := C.CString(k)

		switch v.(type) {
		case (*Blob):
			v_blob := v.(*Blob)
			C.FLMutableDict_SetBlob(mutableDict, C.FLStr(c_key), v_blob.blob)
			C.free(unsafe.Pointer(c_key))
			continue;
		}

		fl_slot := C.FLMutableDict_Set(mutableDict, C.FLStr(c_key))
		storeGoValueInSlot(fl_slot, v)
		C.free(unsafe.Pointer(c_key))
	}
	
	C.CBLDocument_SetProperties(doc.doc, mutableDict)
	C.free(unsafe.Pointer(mutableDict))
	return true
}

/** Sets a mutable document's properties from a JSON string. */
// bool CBLDocument_SetPropertiesAsJSON(CBLDocument* _cbl_nonnull,
                                    //  const char *json _cbl_nonnull,
									//  CBLError*) CBLAPI;
func (doc *Document) SetPropertiesAsJSON(json string) bool {
	err := C.CBLError{}
	c_json := C.CString(json)
	result := bool(C.CBLDocument_SetPropertiesAsJSON(doc.doc, c_json, &err))
	if result && err.code == 0 {
		C.free(unsafe.Pointer(c_json))
		documentProperties(doc)
		return result
	}
	return result
}

/** Returns the time, if any, at which a given document will expire and be purged.
    Documents don't normally expire; you have to call \ref CBLDatabase_SetDocumentExpiration
    to set a document's expiration time.
    @param db  The database.
    @param docID  The ID of the document.
    @param error  On failure, an error is written here.
    @return  The expiration time as a CBLTimestamp (milliseconds since Unix epoch),
             or 0 if the document does not have an expiration,
             or -1 if the call failed. */
// CBLTimestamp CBLDatabase_GetDocumentExpiration(CBLDatabase* db _cbl_nonnull,
//                                                const char *docID _cbl_nonnull,
//                                                CBLError* error) CBLAPI;
func (db *Database) GetDocumentExpiration(docId string) (int64, error) {
	err := C.CBLError{}
	c_docId := C.CString(docId)
	timestamp := C.CBLDatabase_GetDocumentExpiration(db.db, c_docId, &err)
	if err.code == 0 {
		return int64(timestamp), nil
	}
	ErrCBLInternalError = fmt.Errorf("CBL: Problem Retrieving Document Timestamp. Domain: %d Code: %d", err.domain, err.code)
	return -1, ErrCBLInternalError
}
/** Sets or clears the expiration time of a document.
    @note  The purging of expired documents is not yet automatic; you will need to call
            \ref CBLDatabase_PurgeExpiredDocuments when the time comes, to make it happen.
    @param db  The database.
    @param docID  The ID of the document.
    @param expiration  The expiration time as a CBLTimestamp (milliseconds since Unix epoch),
                        or 0 if the document should never expire.
    @param error  On failure, an error is written here.
    @return  True on success, false on failure. */
// bool CBLDatabase_SetDocumentExpiration(CBLDatabase* db _cbl_nonnull,
//                                        const char *docID _cbl_nonnull,
//                                        CBLTimestamp expiration,
//                                        CBLError* error) CBLAPI;

/** @} */
func (db *Database) SetDocumentExpiration(docId string, timestamp int64) bool {
	err := C.CBLError{}
	c_docId := C.CString(docId)
	result := bool(C.CBLDatabase_SetDocumentExpiration(db.db, c_docId, C.CBLTimestamp(timestamp), &err))
	if result && err.code == 0 {
		C.free(unsafe.Pointer(c_docId))
		return result
	}
	return result
}


/** \name  Document listeners
    @{
    A document change listener lets you detect changes made to a specific document after they
    are persisted to the database.
    @note If there are multiple CBLDatabase instances on the same database file, each one's
    document listeners will be notified of changes made by other database instances.
 */

/** A document change listener callback, invoked after a specific document is changed on disk.
    @warning  By default, this listener may be called on arbitrary threads. If your code isn't
                    prepared for that, you may want to use \ref CBLDatabase_BufferNotifications
                    so that listeners will be called in a safe context.
    @param context  An arbitrary value given when the callback was registered.
    @param db  The database containing the document.
    @param docID  The document's ID. */
// typedef void (*CBLDocumentChangeListener)(void *context,
//                                           const CBLDatabase* db _cbl_nonnull,
//                                           const char *docID _cbl_nonnull);
	type DocumentChangeListener func(ctx context.Context, db *Database, docId string)
/** Registers a document change listener callback. It will be called after a specific document
    is changed on disk.
    @param db  The database to observe.
    @param docID  The ID of the document to observe.
    @param listener  The callback to be invoked.
    @param context  An opaque value that will be passed to the callback.
    @return  A token to be passed to \ref CBLListener_Remove when it's time to remove the
            listener.*/
// _cbl_warn_unused
// CBLListenerToken* CBLDatabase_AddDocumentChangeListener(const CBLDatabase* db _cbl_nonnull,
//                                                         const char* docID _cbl_nonnull,
//                                                         CBLDocumentChangeListener listener _cbl_nonnull,
//                                                         void *context) CBLAPI;
func (db *Database) AddDocumentChangeListener(listener DocumentChangeListener, docId string, ctx context.Context) error {
	if v := ctx.Value(uuid); v != nil {
		key, ok := v.(string)
		if ok {
			documentChangeListeners[key] = listener
			c_docId := C.CString(docId)
			token := C.CBLDatabase_AddDocumentChangeListener(db.db, c_docId,
				(C.CBLDocumentChangeListener)(C.gatewayDocumentChangeGoCallback), unsafe.Pointer(&ctx))
			listenerTokens[key] = token
			return nil
		}
	}
	ErrCBLInternalError = fmt.Errorf("CBL: No UUID present in context.")
	return ErrCBLInternalError
}
/** @} */
/** @} */
