package cblcgo
/*
#cgo LDFLAGS: -L. -lCouchbaseLiteC
#include <stdlib.h>
#include <stdio.h>
#include "include/CouchbaseLite.h"

*/
import "C"
import "unsafe"
import "fmt"

/** \defgroup blobs Blobs
    @{
    A \ref CBLBlob is a binary data blob associated with a document.

    The content of the blob is not stored in the document, but externally in the database.
    It is loaded only on demand, and can be streamed. Blobs can be arbitrarily large, although
    Sync Gateway will only accept blobs under 20MB.

    The document contains only a blob reference: a dictionary with the special marker property
    `"@type":"blob"`, and another property `digest` whose value is a hex SHA-1 digest of the
    blob's data. This digest is used as the key to retrieve the blob data.
    The dictionary usually also has the property `length`, containing the blob's length in bytes,
    and it may have the property `content_type`, containing a MIME type.

    A \ref CBLBlob object acts as a proxy for such a dictionary in a \ref CBLDocument. Once
    you've loaded a document and located the \ref FLDict holding the blob reference, call
    \ref CBLBlob_Get on it to create a \ref CBLBlob object you can call.
    The object has accessors for the blob's metadata and for loading the data itself.

    To create a new blob from in-memory data, call \ref CBLBlob_CreateWithData, then call
    \ref FLMutableDict_SetBlob or \ref FLMutableArray_SetBlob to add the \ref CBLBlob to the
    document (or to a dictionary or array property of the document.)

    To create a new blob from a stream, call \ref CBLBlobWriter_New to create a
    \ref CBLBlobWriteStream, then make one or more calls to \ref CBLBlobWriter_Write to write
    data to the blob, then finally call \ref CBLBlob_CreateWithStream to create the blob.
    To store the blob into a document, do as in the previous paragraph.

 */

//  CBL_CORE_API extern const FLSlice kCBLTypeProperty;             ///< `"@type"`
//  CBL_CORE_API extern const FLSlice kCBLBlobType;                 ///< `"blob"`

//  CBL_CORE_API extern const FLSlice kCBLBlobDigestProperty;       ///< `"digest"`
//  CBL_CORE_API extern const FLSlice kCBLBlobLengthProperty;       ///< `"length"`
//  CBL_CORE_API extern const FLSlice kCBLBlobContentTypeProperty;  ///< `"content_type"`

type Blob struct {
	blob *C.CBLBlob
	Props map[string]interface{}
}

 /** Returns true if a dictionary in a document is a blob reference.
	 If so, you can call \ref CBLBlob_Get to access it.
	 @note This function tests whether the dictionary has a `@type` property,
			 whose value is `"blob"`. */
//  bool CBL_IsBlob(FLDict) CBLAPI;
func isBlob(fl_dict C.FLDict) bool {
	return bool(C.CBL_IsBlob(fl_dict))
}

 
//  CBL_REFCOUNTED(CBLBlob*, Blob);

 /** Returns a CBLBlob object corresponding to a blob dictionary in a document.
	 @param blobDict  A dictionary in a document.
	 @return  A CBLBlob instance for this blob, or NULL if the dictionary is not a blob. */
//  const CBLBlob* CBLBlob_Get(FLDict blobDict) CBLAPI;
func getBlob(fl_dict C.FLDict) (*Blob, error) {
	c_blob := C.CBLBlob_Get(fl_dict)
	if props, err := getKeyValuePropMap(getBlobPoperties(c_blob)); err == nil {
		blob := Blob{c_blob, props}
		return &blob, nil
	}
	return nil, ErrProblemGettingBlobWithData
}

// #pragma mark - BLOB METADATA:

 /** Returns the length in bytes of a blob's content (from its `length` property). */
//  uint64_t CBLBlob_Length(const CBLBlob* _cbl_nonnull) CBLAPI;
func (blob *Blob) Length() uint64 {
	return uint64(C.CBLBlob_Length(blob.blob))
}

 /** Returns the cryptographic digest of a blob's content (from its `digest` property). */
//  const char* CBLBlob_Digest(const CBLBlob* _cbl_nonnull) CBLAPI;
func (blob *Blob) Digest() string {
	c_dig := C.CBLBlob_Digest(blob.blob)
	dig := C.GoString(c_dig)
	C.free(unsafe.Pointer(c_dig))
	return dig
}

 /** Returns a blob's MIME type, if its metadata has a `content_type` property. */
//  const char* CBLBlob_ContentType(const CBLBlob* _cbl_nonnull) CBLAPI;
func (blob *Blob) ContentType() string {
	c_type := C.CBLBlob_ContentType(blob.blob)
	t := C.GoString(c_type)
	C.free(unsafe.Pointer(c_type))
	return t
}
 /** Returns a blob's metadata. This includes the `digest`, `length` and `content_type`
	 properties, as well as any custom ones that may have been added. */
//  FLDict CBLBlob_Properties(const CBLBlob* _cbl_nonnull) CBLAPI;
func getBlobPoperties(blob *C.CBLBlob) C.FLDict {
	return C.CBLBlob_Properties(blob)
}

// #pragma mark - READING:

 /** Reads the blob's contents into memory and returns them.
	 You are responsible for calling \ref FLSliceResult_Free on the returned data when done.
	 @warning  This can potentially allocate a very large heap block! */
//  FLSliceResult CBLBlob_LoadContent(const CBLBlob* _cbl_nonnull, CBLError *outError) CBLAPI;
func blobLoadContent(blob *C.CBLBlob) C.FLSliceResult {
	err := (*C.CBLError)(C.malloc(C.sizeof_CBLError))
	defer C.free(unsafe.Pointer(err))
	result := C.CBLBlob_LoadContent(blob, err)
	if (*err).code == 0 {
		return result
	}
	return C.FLSliceResult{}
}

 /** A stream for reading a blob's content. */
//  typedef struct CBLBlobReadStream CBLBlobReadStream;
type BlobReadStream struct {
	rs *C.CBLBlobReadStream
}

 /** Opens a stream for reading a blob's content. */
//  CBLBlobReadStream* CBLBlob_OpenContentStream(const CBLBlob* _cbl_nonnull, CBLError *outError) CBLAPI;
func (blob *Blob) NewReadStream() *BlobReadStream {
	err := (*C.CBLError)(C.malloc(C.sizeof_CBLError))
	defer C.free(unsafe.Pointer(err))
	brs := C.CBLBlob_OpenContentStream(blob.blob, err)
	if (*err).code == 0 {
		rs := BlobReadStream{brs}
		return &rs
	}
	return nil
}

 /** Reads data from a blob.
	 @param stream  The stream to read from.
	 @param dst  The address to copy the read data to.
	 @param maxLength  The maximum number of bytes to read.
	 @param outError  On failure, an error will be stored here if non-NULL.
	 @return  The actual number of bytes read; 0 if at EOF, -1 on error. */
//  int CBLBlobReader_Read(CBLBlobReadStream* stream _cbl_nonnull,
// 						void *dst _cbl_nonnull,
// 						size_t maxLength,
// 						CBLError *outError) CBLAPI;
func (blob *Blob) Read(res *BlobReadStream, dst []byte) (int, error) {
	err := (*C.CBLError)(C.malloc(C.sizeof_CBLError))
	defer C.free(unsafe.Pointer(err))
	maxLength := len(dst)
	c_dst := C.CBytes(dst)
	bytesRead := C.CBLBlobReader_Read(res.rs, c_dst, C.size_t(maxLength), err)
	if (*err).code == 0 {
		readData := C.GoBytes(c_dst, bytesRead)
		copy(dst, readData)
		C.free(c_dst)
		return int(bytesRead), nil
	}
	C.free(c_dst)
	ErrCBLInternalError = fmt.Errorf("CBL: Problem Reading Blob. Domain: %d Code: %d", (*err).domain, (*err).code)
	return -1, ErrCBLInternalError
}
 /** Closes a CBLBlobReadStream. */
//  void CBLBlobReader_Close(CBLBlobReadStream*) CBLAPI;
func (blob *Blob) CloseReader(rs *BlobReadStream) {
	C.CBLBlobReader_Close(rs.rs)
}

// #pragma mark - CREATING:

 /** Creates a new blob given its contents as a single block of data.
	 @note  You are responsible for releasing the \ref CBLBlob, but not until after its document
			 has been saved.
	 @param contentType  The MIME type (optional).
	 @param contents  The data's address and length.
	 @return  A new CBLBlob instance. */
//  CBLBlob* CBLBlob_CreateWithData(const char *contentType,
// 								 FLSlice contents) CBLAPI;
func NewBlobWithData(contentType string, contents []byte) (*Blob, error) {
	c_ct := C.CString(contentType)
	size := C.size_t(len(contents))
	c_contents := C.CBytes(contents)
	fl_slice := C.FLSlice{c_contents, size}
	c_blob := C.CBLBlob_CreateWithData(c_ct, fl_slice)
	C.free(unsafe.Pointer(c_ct))
	C.free(c_contents)
	if props, err := getKeyValuePropMap(getBlobPoperties(c_blob)); err == nil {
		blob := Blob{c_blob, props}
		return &blob, nil
	}
	return nil, ErrProblemCreatingBlobWithData
}



 /** A stream for writing a new blob to the database. */
//  typedef struct CBLBlobWriteStream CBLBlobWriteStream;
type BlobWriteStream struct {
	wrs *C.CBLBlobWriteStream
}

 /** Opens a stream for writing a new blob.
	 You should next call \ref CBLBlobWriter_Write one or more times to write the data,
	 then \ref CBLBlob_CreateWithStream to create the blob.

	 If for some reason you need to abort, just call \ref CBLBlobWriter_Close. */
//  CBLBlobWriteStream* CBLBlobWriter_New(CBLDatabase *db _cbl_nonnull,
// 									   CBLError *outError) CBLAPI;
func (db *Database) NewBlobWriter() (*BlobWriteStream, error) {
	err := (*C.CBLError)(C.malloc(C.sizeof_CBLError))
	defer C.free(unsafe.Pointer(err))
	c_wrs := C.CBLBlobWriter_New(db.db, err)
	if (*err).code == 0 {
		bwrs := BlobWriteStream{c_wrs}
		return &bwrs, nil
	}
	ErrCBLInternalError = fmt.Errorf("CBL: Error Creating New Blob Writer. Domain: %d Code: %d", (*err).domain, (*err).code)
	return nil, ErrCBLInternalError
}


 /** Closes a blob-writing stream, if you need to give up without creating a \ref CBLBlob. */
//  void CBLBlobWriter_Close(CBLBlobWriteStream*) CBLAPI;
func (db *Database) CloseBlobWriter(bwrs *BlobWriteStream) {
	C.CBLBlobWriter_Close(bwrs.wrs)
}

 /** Writes data to a new blob.
	 @param writer  The stream to write to.
	 @param data  The address of the data to write.
	 @param length  The length of the data to write.
	 @param outError  On failure, error info will be written here.
	 @return  True on success, false on failure. */
//  bool CBLBlobWriter_Write(CBLBlobWriteStream* writer _cbl_nonnull,
// 						   const void *data _cbl_nonnull,
// 						   size_t length,
// 						   CBLError *outError) CBLAPI;
func (blob *Blob) Write(bwrs *BlobWriteStream, data []byte) bool {
	err := (*C.CBLError)(C.malloc(C.sizeof_CBLError))
	defer C.free(unsafe.Pointer(err))
	length := C.size_t(len(data))
	c_data := C.CBytes(data)
	result := bool(C.CBLBlobWriter_Write(bwrs.wrs, c_data, length, err))
	if result && (*err).code == 0 {
		copy(data, C.GoBytes(c_data, C.int(length)))
		return result
	}
	C.free(c_data)
	return result
}


 /** Creates a new blob after its data has been written to a \ref CBLBlobWriteStream.
	 You should then add the blob to a mutable document as a property -- see
	 \ref FLMutableDict_SetBlob and \ref FLMutableArray_SetBlob.
	 @note  You are responsible for releasing the CBLBlob reference.
	 @note  Do not free the stream; the blob will do that.
	 @param contentType  The MIME type (optional).
	 @param writer  The blob-writing stream the data was written to.
	 @return  A new CBLBlob instance. */
//  CBLBlob* CBLBlob_CreateWithStream(const char *contentType,
// 								   CBLBlobWriteStream* writer _cbl_nonnull) CBLAPI;
func CreateBlobWithStream(contentType string, writer *BlobWriteStream) (*Blob, error) {
	c_ct := C.CString(contentType)
	c_blob := C.CBLBlob_CreateWithStream(c_ct, writer.wrs)
	C.free(unsafe.Pointer(c_ct))
	if props, err := getKeyValuePropMap(getBlobPoperties(c_blob)); err == nil {
		blob := Blob{c_blob, props}
		return &blob, nil
	}
	return nil, ErrProblemCreatingBlobWithData
}

// #pragma mark - FLEECE UTILITIES:

 /** Returns true if a value in a document is a blob reference.
	 If so, you can call \ref FLValue_GetBlob to access it. */
//  static inline bool FLValue_IsBlob(FLValue v) {
// 	 return CBL_IsBlob(FLValue_AsDict(v));
//  }

 /** Instantiates a \ref CBLBlob object corresponding to a blob dictionary in a document.
	 @param value  The value (dictionary) in the document.
	 @return  A \ref CBLBlob instance for this blob, or `NULL` if the value is not a blob.
	 @note You are responsible for releasing the \ref CBLBlob object.  */
//  static inline const CBLBlob* FLValue_GetBlob(FLValue value) {
// 	 return CBLBlob_Get(FLValue_AsDict(value));
//  }

 /** Stores a blob in a mutable array. */
//  void FLMutableArray_SetBlob(FLMutableArray array _cbl_nonnull,
// 							 uint32_t index,
// 							 CBLBlob* blob _cbl_nonnull) CBLAPI;

//  /** Stores a blob in a mutable dictionary. */
//  void FLMutableDict_SetBlob(FLMutableDict dict _cbl_nonnull,
// 							FLString key,
// 							CBLBlob* blob _cbl_nonnull) CBLAPI;

/** @} */