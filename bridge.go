package cblcgo
/*
#cgo LDFLAGS: -L. -lCouchbaseLiteC
#include <stdlib.h>
#include <stdio.h>
#include "include/CouchbaseLite.h"

void gatewayDatabaseChangeGoCallback(void *context, const CBLDatabase* db _cbl_nonnull, unsigned numDocs, const char **docIDs _cbl_nonnull);
void gatewayDocumentChangeGoCallback(void *context, const CBLDatabase* db _cbl_nonnull, const char *docID _cbl_nonnull);
void gatewayQueryChangeGoCallback(void *context, CBLQuery* query _cbl_nonnull);
void notificationReadyCallback(void *context, CBLDatabase* db _cbl_nonnull);
void gatewayPushFilterCallback(void *context, CBLDocument* doc, bool isDeleted);
void gatewayPullFilterCallback(void *context, CBLDocument* doc, bool isDeleted);
void gatewayReplicatorChangeCallback(void *context, CBLReplicator *replicator _cbl_nonnull, const CBLReplicatorStatus *status _cbl_nonnull);
void gatewayReplicatedDocumentCallback(void *context, CBLReplicator *replicator _cbl_nonnull, bool isPush, unsigned numDocuments, const CBLReplicatedDocument* documents);



FLValue FLArray_AsValue(FLArray);
FLValue FLDict_AsValue(FLDict);

*/
import "C"
import "unsafe"
import "context"
import "reflect"

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

	// v := (*ctx).Value(uuid).(string)
	// (databaseChangeListeners[v])(*ctx, &database, ids)
	callback := (*ctx).Value(callback).(DatabaseChangeListener)
	callback(*ctx, &database, ids)
}
//export documentListenerBridge
func documentListenerBridge(c unsafe.Pointer, db *C.CBLDatabase, c_docID *C.char) {
	ctx := (*context.Context)(c)
	docId := C.GoString(c_docID)
	database := Database{}
	database.db = db
	// v := (*ctx).Value(uuid).(string)
	// (documentChangeListeners[v])(*ctx, &database, docId)
	callback := (*ctx).Value(callback).(DocumentChangeListener)
	callback(*ctx, &database, docId)
}
//export notificationBridge
func notificationBridge(c unsafe.Pointer, db *C.CBLDatabase) {
	ctx := (*context.Context)(c)
	d := Database{}
	d.db = db
	// notificationCallback(*ctx, &d)
	callback := (*ctx).Value(callback).(NotificationsReadyCallback)
	callback(*ctx, &d)
}
//export queryListenerBride
func queryListenerBride(c unsafe.Pointer, query *C.CBLQuery) {
	ctx := (*context.Context)(c)
	q := Query{query}
	// v := (*ctx).Value(uuid).(string)
	// (queryChangeListeners[v])(*ctx, &q)
	callback := (*ctx).Value(callback).(QueryChangeListener)
	callback(*ctx, &q)
}
//export pushFilterBridge
func pushFilterBridge(c unsafe.Pointer, doc *C.CBLDocument, isDeleted C.bool) {
	ctx := (*context.Context)(c)
	d := Document{}
	d.doc = doc
	callback := (*ctx).Value(pushCallback).(ReplicationFilter)
	callback(*ctx, &d, bool(isDeleted))
}
//export pullFilterBridge
func pullFilterBridge(c unsafe.Pointer, doc *C.CBLDocument, isDeleted C.bool) {
	ctx := (*context.Context)(c)
	d := Document{}
	d.doc = doc
	callback := (*ctx).Value(pullCallback).(ReplicationFilter)
	callback(*ctx, &d, bool(isDeleted))
}
//export replicatorChangeBridge
func replicatorChangeBridge(c unsafe.Pointer, replicator *C.CBLReplicator, status *C.CBLReplicatorStatus) {
	ctx := (*context.Context)(c)
	rep := Replicator{replicator}

	e := Error{uint32(status.error.internal_info), uint32(status.error.code), uint32(status.error.domain)}
	activity := ReplicatorActivityLevel(status.activity)
	progress := ReplicatorProgress{float32(status.progress.fractionComplete), uint64(status.progress.documentCount)}
	repStatus := ReplicatorStatus{activity, progress, e}

	callback := (*ctx).Value(callback).(ReplicatorChangeListener)
	callback(*ctx, &rep, &repStatus)
}
//export replicatedDocumentBridge
func replicatedDocumentBridge(c unsafe.Pointer, replicator *C.CBLReplicator, isPush C.bool,
								numDocument C.unsigned, documents *C.CBLReplicatedDocument) {
	ctx := (*context.Context)(c)
	rep := Replicator{replicator}

	e := Error{uint32(documents.error.internal_info), uint32(documents.error.code), uint32(documents.error.domain)}
	id := C.GoString(documents.ID)
	doc_flags := DocumentFlags(documents.flags)
	rep_doc := ReplicatedDocument{id, doc_flags, e}

	callback := (*ctx).Value(callback).(ReplicatedDocumentListener)
	callback(*ctx, &rep, bool(isPush), uint(numDocument), &rep_doc)
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
		break
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
		C.FLArray_Release(fl_arr)
		C.FLMutableArray_Release(mutable_array)
		break
	case reflect.Int:
		value := v.(int)
		C.FLSlot_SetInt(fl_slot, C.int64_t(value))
		break
	case reflect.Int8:
		value := v.(int8)
		C.FLSlot_SetInt(fl_slot, C.int64_t(value))
		break
	case reflect.Int16:
		value := v.(int16)
		C.FLSlot_SetInt(fl_slot, C.int64_t(value))
		break
	case reflect.Int32:
		value := v.(int32)
		C.FLSlot_SetInt(fl_slot, C.int64_t(value))
		break
	case reflect.Int64:
		value := v.(int64)
		C.FLSlot_SetInt(fl_slot, C.int64_t(value))
		break
	case reflect.Uint:
		value := v.(uint)
		C.FLSlot_SetUInt(fl_slot, C.uint64_t(value))
		break
	case reflect.Uintptr:
		value := v.(uintptr)
		C.FLSlot_SetUInt(fl_slot, C.uint64_t(value))
		break
	case reflect.Uint8:
		value := v.(uint8)
		C.FLSlot_SetUInt(fl_slot, C.uint64_t(value))
		break
	case reflect.Uint16:
		value := v.(uint16)
		C.FLSlot_SetUInt(fl_slot, C.uint64_t(value))
		break
	case reflect.Uint32:
		value := v.(uint32)
		C.FLSlot_SetUInt(fl_slot, C.uint64_t(value))
		break
	case reflect.Uint64:
		value := v.(uint64)
		C.FLSlot_SetUInt(fl_slot, C.uint64_t(value))
		break
	case reflect.Float32:
		value := v.(float32)
		C.FLSlot_SetFloat(fl_slot, C.float(value))
		break
	case reflect.Float64:
		value := v.(float64)
		C.FLSlot_SetDouble(fl_slot, C.double(value))
		// double
		break;
	case reflect.Bool:
		value := v.(bool)
		C.FLSlot_SetBool(fl_slot, C.bool(value))
		break
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
			C.FLDict_Release(fl_dict)
			C.FLMutableDict_Release(mutable_dict)
			break
		default:
			return ErrUnsupportedGoType
		}
	default:
		return ErrUnsupportedGoType
	}
	return nil
}