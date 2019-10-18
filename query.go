package cblcgo
/*
#cgo LDFLAGS: -L. -lCouchbaseLiteC
#include <stdlib.h>
#include <stdio.h>
#include "include/CouchbaseLite.h"

void queryListenerBride(void *, CBLQuery *);

void gatewayQueryChangeGoCallback(void *context, CBLQuery* query _cbl_nonnull) {
	queryListenerBride(context, query);
}

*/
import "C"
import "unsafe"
import "fmt"
import "context"

/** \defgroup queries   Queries
    @{
    A CBLQuery represents a compiled database query. The query language is a large subset of
    the [N1QL](https://www.couchbase.com/products/n1ql) language from Couchbase Server, which
    you can think of as "SQL for JSON" or "SQL++".

    Queries may be given either in
    [N1QL syntax](https://docs.couchbase.com/server/6.0/n1ql/n1ql-language-reference/index.html),
    or in JSON using a
    [schema](https://github.com/couchbase/couchbase-lite-core/wiki/JSON-Query-Schema)
    that resembles a parse tree of N1QL. The JSON syntax is harder for humans, but much more
    amenable to machine generation, if you need to create queries programmatically or translate
    them from some other form.
 */


/** Query languages */
type QueryLanguage uint32

const (
    JSONLanguage QueryLanguage = iota   ///< [JSON query schema](https://github.com/couchbase/couchbase-lite-core/wiki/JSON-Query-Schema)
    N1QLLanguage    ///< [N1QL syntax](https://docs.couchbase.com/server/6.0/n1ql/n1ql-language-reference/index.html)
)

type Query struct {
	q *C.CBLQuery
}

type ResultSet struct {
	rs *C.CBLResultSet
}


/** \name  Query objects
    @{ */

/** Creates a new query by compiling the input string.
    This is fast, but not instantaneous. If you need to run the same query many times, keep the
    \ref CBLQuery around instead of compiling it each time. If you need to run related queries
    with only some values different, create one query with placeholder parameter(s), and substitute
    the desired value(s) with \ref CBLQuery_SetParameters each time you run the query.
    @note  You must release the \ref CBLQuery when you're finished with it.
    @param db  The database to query.
    @param language  The query language,
            [JSON](https://github.com/couchbase/couchbase-lite-core/wiki/JSON-Query-Schema) or
            [N1QL](https://docs.couchbase.com/server/4.0/n1ql/n1ql-language-reference/index.html).
    @param queryString  The query string.
    @param outErrorPos  If non-NULL, then on a parse error the approximate byte offset in the
                    input expression will be stored here (or -1 if not known/applicable.)
    @param error  On failure, the error will be written here.
    @return  The new query object. */
// _cbl_warn_unused
// CBLQuery* CBLQuery_New(const CBLDatabase* db _cbl_nonnull,
//                        CBLQueryLanguage language,
//                        const char *queryString _cbl_nonnull,
//                        int *outErrorPos,
//                        CBLError* error) CBLAPI;
func (db *Database) NewQuery(language QueryLanguage, queryString string) (*Query, error) {
	err := (*C.CBLError)(C.malloc(C.sizeof_CBLError))
	defer C.free(unsafe.Pointer(err))
  outErrorPos := (*C.int)(C.malloc(C.sizeof_int))
  defer C.free(unsafe.Pointer(outErrorPos))
	c_query_str := C.CString(queryString)
	c_query := C.CBLQuery_New(db.db, C.CBLQueryLanguage(language), c_query_str, outErrorPos, err)
	C.free(unsafe.Pointer(c_query_str))
	if (*err).code == 0 {
		query := Query{c_query}
		return &query, nil
	}
	ErrProblemPreparingQuery = fmt.Errorf("CBL: Problem Preparing Query. Domain: %d Code: %d", (*err).domain, (*err).code)
	return nil, ErrProblemPreparingQuery
}


//CBL_REFCOUNTED(CBLQuery*, Query);

/** Assigns values to the query's parameters.
    These values will be substited for those parameters whenever the query is executed,
    until they are next assigned.

    Parameters are specified in the query source as
    e.g. `$PARAM` (N1QL) or `["$PARAM"]` (JSON). In this example, the `parameters` dictionary
    to this call should have a key `PARAM` that maps to the value of the parameter.
    @param query  The query.
    @param parameters  The parameters in the form of a Fleece \ref FLDict "dictionary" whose
            keys are the parameter names. (It's easiest to construct this by using the mutable
            API, i.e. calling \ref FLMutableDict_New and adding keys/values.) */
//void CBLQuery_SetParameters(CBLQuery* _cbl_nonnull query,
							//FLDict _cbl_nonnull parameters) CBLAPI;
func (q *Query) SetParameters(parameters map[string]interface{}) error {
	mutable_dict := C.FLMutableDict_New()

	for key, val := range parameters {
		c_key := C.CString(key)
		v_slot := C.FLMutableDict_Set(mutable_dict, C.FLStr(c_key))
		storeGoValueInSlot(v_slot, val)
		C.free(unsafe.Pointer(c_key))
	}

	fl_dict := C.FLMutableDict_GetSource(mutable_dict)
	C.CBLQuery_SetParameters(q.q, fl_dict)
	return nil
}

/** Returns the query's current parameter bindings, if any. */
//FLDict CBLQuery_Parameters(CBLQuery* _cbl_nonnull query) CBLAPI;
func (q *Query) GetParameters() (map[string]interface{}, error) {
	fl_dict := C.CBLQuery_Parameters(q.q)
	if params, err := getKeyValuePropMap(fl_dict); err == nil {
		return params, nil
	} else {
		return nil, err
	}
}

/** Assigns values to the query's parameters, from JSON data.
    See \ref CBLQuery_SetParameters for details.
    @param query  The query.
    @param json  The parameters in the form of a JSON-encoded object whose
            keys are the parameter names. (You may use JSON5 syntax.) */
// bool CBLQuery_SetParametersAsJSON(CBLQuery* _cbl_nonnull query,
//                                   const char* _cbl_nonnull json) CBLAPI;
func (q *Query) SetParametersASJSON(json string) bool {
	c_json := C.CString(json)
	result := bool(C.CBLQuery_SetParametersAsJSON(q.q, c_json))
	C.free(unsafe.Pointer(c_json))
	return result
}

/** Runs the query, returning the results.
    To obtain the results you'll typically call \ref CBLResultSet_Next in a `while` loop,
    examining the values in the \ref CBLResultSet each time around.
    @note  You must release the result set when you're finished with it. */
// _cbl_warn_unused
// CBLResultSet* CBLQuery_Execute(CBLQuery* _cbl_nonnull, CBLError*) CBLAPI;
func (q *Query) Execute() (*ResultSet, error) {
	err := (*C.CBLError)(C.malloc(C.sizeof_CBLError))
	defer C.free(unsafe.Pointer(err))
	c_result_set := C.CBLQuery_Execute(q.q, err)
	if (*err).code == 0 {
		results := ResultSet{c_result_set}
		return &results, nil
	}
	ErrProblemExecutingQuery = fmt.Errorf("CBL: Problem Executing Query. Domain: %d Code: %d", (*err).domain, (*err).code)
	return nil, ErrProblemExecutingQuery
}

/** Returns information about the query, including the translated SQLite form, and the search
    strategy. You can use this to help optimize the query: the word `SCAN` in the strategy
    indicates a linear scan of the entire database, which should be avoided by adding an index.
    The strategy will also show which index(es), if any, are used. */
//FLSliceResult CBLQuery_Explain(CBLQuery* _cbl_nonnull) CBLAPI;
func (q *Query) Explain() string {
	fl_slice := C.CBLQuery_Explain(q.q)
	strategy := C.GoStringN((*C.char)(fl_slice.buf), C.int(fl_slice.size))
	C.FLSliceResult_Release(fl_slice)
	return strategy
}

/** Returns the number of columns in each result. */
// unsigned CBLQuery_ColumnCount(CBLQuery* _cbl_nonnull) CBLAPI;
func (q *Query) ColumnCount() uint {
	return uint(C.CBLQuery_ColumnCount(q.q))
}

/** Returns the name of a column in the result.
    The column name is based on its expression in the `SELECT...` or `WHAT:` section of the
    query. A column that returns a property or property path will be named after that property.
    A column that returns an expression will have an automatically-generated name like `$1`.
    To give a column a custom name, use the `AS` syntax in the query.*/
// FLSlice CBLQuery_ColumnName(CBLQuery* _cbl_nonnull,
//                             unsigned columnIndex) CBLAPI;
func (q *Query) ColumnNameAt(columnIndex uint) string {
	c_slice := C.CBLQuery_ColumnName(q.q, C.unsigned(columnIndex))
	col := C.GoStringN((*C.char)(c_slice.buf), C.int(c_slice.size))	
	return col
}

/** @} */



/** \name  Result sets
    @{
    A `CBLResultSet` is an iterator over the results returned by a query. It exposes one
    result at a time -- as a collection of values indexed either by position or by name --
    and can be stepped from one result to the next.

    It's important to note that the initial position of the iterator is _before_ the first
    result, so \ref CBLResultSet_Next must be called _first_.
 */

/** Moves the result-set iterator to the next result.
    Returns false if there are no more results.
    @warning This must be called _before_ examining the first result. */
// bool CBLResultSet_Next(CBLResultSet* _cbl_nonnull) CBLAPI;
func (res *ResultSet) Next() bool {
	result := bool(C.CBLResultSet_Next(res.rs))
	return result
}

/** Returns the value of a column of the current result, given its (zero-based) numeric index.
    This may return a NULL pointer, indicating `MISSING`, if the value doesn't exist, e.g. if
    the column is a property that doesn't exist in the document. */
// FLValue CBLResultSet_ValueAtIndex(CBLResultSet* _cbl_nonnull,
								//   unsigned index) CBLAPI;
func (res *ResultSet) ValueAtIndex(index uint) interface{} {
	fl_val := C.CBLResultSet_ValueAtIndex(res.rs, C.unsigned(index))
	if value, err := getFLValueToGoValue(fl_val); err != nil {
		return value
	}
	return nil
}

/** Returns the value of a column of the current result, given its name.
    This may return a NULL pointer, indicating `MISSING`, if the value doesn't exist, e.g. if
    the column is a property that doesn't exist in the document. (Or, of course, if the key
    is not a column name in this query.)
    @note  See \ref CBLQuery_ColumnName for a discussion of column names. */
// FLValue CBLResultSet_ValueForKey(CBLResultSet* _cbl_nonnull,
								//  const char* key _cbl_nonnull) CBLAPI;
func (res *ResultSet) ValueForKey(key string) interface{} {
	c_key := C.CString(key)
	fl_val := C.CBLResultSet_ValueForKey(res.rs, c_key)
	C.free(unsafe.Pointer(c_key))
	if value, err := getFLValueToGoValue(fl_val); err != nil {
		return value
	}
	return nil
}

// CBL_REFCOUNTED(CBLResultSet*, ResultSet);

/** @} */



/** \name  Change listener
    @{
    Adding a change listener to a query turns it into a "live query". When changes are made to
    documents, the query will periodically re-run and compare its results with the prior
    results; if the new results are different, the listener callback will be called.

    @note  The result set passed to the listener is the _entire new result set_, not just the
            rows that changed.
 */

/** A callback to be invoked after the query's results have changed.
    The actual result set can be obtained by calling \ref CBLQuery_CurrentResults, either during
    the callback or at any time thereafter.
    @warning  By default, this listener may be called on arbitrary threads. If your code isn't
                    prepared for that, you may want to use \ref CBLDatabase_BufferNotifications
                    so that listeners will be called in a safe context.
    @param context  The same `context` value that you passed when adding the listener.
    @param query  The query that triggered the listener. */
// typedef void (*CBLQueryChangeListener)(void *context,
									//    CBLQuery* query _cbl_nonnull);
type QueryChangeListener func(ctx context.Context, query *Query)

/** Registers a change listener callback with a query, turning it into a "live query" until
    the listener is removed (via \ref CBLListener_Remove).

    When the first change listener is added, the query will run (in the background) and notify
    the listener(s) of the results when ready. After that, it will run in the background after
    the database changes, and only notify the listeners when the result set changes.
    @param query  The query to observe.
    @param listener  The callback to be invoked.
    @param context  An opaque value that will be passed to the callback.
    @return  A token to be passed to \ref CBLListener_Remove when it's time to remove the
            listener.*/
// _cbl_warn_unused
// CBLListenerToken* CBLQuery_AddChangeListener(CBLQuery* query _cbl_nonnull,
                                            //  CBLQueryChangeListener listener _cbl_nonnull,
											//  void *context) CBLAPI;
func (q *Query) AddChangeListener(listener QueryChangeListener, ctx context.Context) *ListenerToken {
  ctx = context.WithValue(ctx, callback, listener)
  token := C.CBLQuery_AddChangeListener(q.q, (C.CBLQueryChangeListener)(C.gatewayQueryChangeGoCallback), unsafe.Pointer(&ctx))
  listener_token := ListenerToken{token}
  return &listener_token
}

/** Returns the query's _entire_ current result set, after it's been announced via a call to the
    listener's callback.
    The returned object is valid until the next call to \ref CBLQuery_CurrentResults (with the
    same query and listener) or until you free the listener. If you need to keep it alive longer,
    retain it yourself.
    @param query  The query being listened to.
    @param listener  The query listener that was notified.
    @param error  If the query failed to run, the error will be stored here.
    @return  The query's current result set, or NULL if the query failed to run. */
// CBLResultSet* CBLQuery_CurrentResults(CBLQuery* query _cbl_nonnull,
//                                       CBLListenerToken *listener _cbl_nonnull,
//                                       CBLError *error) CBLAPI;

// func (q *Query) CurrentResults(listener *ListenerToken) (*ResultSet, error) {
// err := (*C.CBLError)(C.malloc(C.sizeof_CBLError))
// defer C.free(unsafe.Pointer(err))
// 	c_result_set := C.CBLQuery_CurrentResults(q.q, listener.token, err)
// 	result_set := ResultSet{c_result_set}
// 	return &result_set, nil
// }

/** @} */



/** \name  Database Indexes
    @{
    Indexes are used to speed up queries by allowing fast -- O(log n) -- lookup of documents
    that have specific values or ranges of values. The values may be properties, or expressions
    based on properties.

    An index will speed up queries that use the expression it indexes, but it takes up space in
    the database file, and it slows down document saves slightly because it needs to be kept up
    to date when documents change.

    Tuning a database with indexes can be a tricky task. Fortunately, a lot has been written about
    it in the relational-database (SQL) realm, and much of that advice holds for Couchbase Lite.
    You may find SQLite's documentation particularly helpful since Couchbase Lite's querying is
    based on SQLite.

    Two types of indexes are currently supported:
        * Value indexes speed up queries by making it possible to look up property (or expression)
          values without scanning every document. They're just like regular indexes in SQL or N1QL.
          Multiple expressions are supported; the first is the primary key, second is secondary.
          Expressions must evaluate to scalar types (boolean, number, string).
        * Full-Text Search (FTS) indexes enable fast search of natural-language words or phrases
          by using the `MATCH` operator in a query. A FTS index is **required** for full-text
          search: a query with a `MATCH` operator will fail to compile unless there is already a
          FTS index for the property/expression being matched. Only a single expression is
          currently allowed, and it must evaluate to a string. */


/** Types of database indexes. */
// typedef CBL_ENUM(uint32_t, CBLIndexType) {
//     kCBLValueIndex,         ///< An index that stores property or expression values
//     kCBLFullTextIndex       ///< An index of strings, that enables searching for words with `MATCH`
// };
type IndexType uint32

const (
	ValueIndex IndexType = iota
	FullTextIndex
)


/** Parameters for creating a database index. */
// typedef struct {
//     /** The type of index to create. */
//     CBLIndexType type;

//     /** A JSON array describing each column of the index. */
//     const char* keyExpressionsJSON;

//     /** In a full-text index, should diacritical marks (accents) be ignored?
//         Defaults to false. Generally this should be left `false` for non-English text. */
//     bool ignoreAccents;

//     /** In a full-text index, the dominant language. Setting this enables word stemming, i.e.
//         matching different cases of the same word ("big" and "bigger", for instance) and ignoring
//         common "stop-words" ("the", "a", "of", etc.)

//         Can be an ISO-639 language code or a lowercase (English) language name; supported
//         languages are: da/danish, nl/dutch, en/english, fi/finnish, fr/french, de/german,
//         hu/hungarian, it/italian, no/norwegian, pt/portuguese, ro/romanian, ru/russian,
//         es/spanish, sv/swedish, tr/turkish.
     
//         If left null,  or set to an unrecognized language, no language-specific behaviors
//         such as stemming and stop-word removal occur. */
//     const char* language;
// } CBLIndexSpec;
type IndexSpec struct {
	Type IndexType
	KeyExpressionsJSON string
	IgnoreAccents bool
	Language string
}

func goIndexSpecToCBLIndexSpec(index IndexSpec) C.CBLIndexSpec {	
	c_key_ex_json := C.CString(index.KeyExpressionsJSON)
	c_lang := C.CString(index.Language)
	c_index := C.CBLIndexSpec{C.CBLIndexType(index.Type), c_key_ex_json, C.bool(index.IgnoreAccents), c_lang}
	return c_index
}

/** Creates a database index.
    Indexes are persistent.
    If an identical index with that name already exists, nothing happens (and no error is returned.)
    If a non-identical index with that name already exists, it is deleted and re-created. */
// bool CBLDatabase_CreateIndex(CBLDatabase *db _cbl_nonnull,
//                              const char* name _cbl_nonnull,
//                              CBLIndexSpec,
//                              CBLError *outError) CBLAPI;
func (db *Database) CreateIndex(name string, indexSpec IndexSpec) bool {
	err := (*C.CBLError)(C.malloc(C.sizeof_CBLError))
	defer C.free(unsafe.Pointer(err))
	c_name := C.CString(name)
	c_index_spec := goIndexSpecToCBLIndexSpec(indexSpec)
	result := bool(C.CBLDatabase_CreateIndex(db.db, c_name, c_index_spec, err))
	return result
}


/** Deletes an index given its name. */
// bool CBLDatabase_DeleteIndex(CBLDatabase *db _cbl_nonnull,
//                              const char *name _cbl_nonnull,
//                              CBLError *outError) CBLAPI;
func (db *Database) DeleteIndex(name string) bool {
	err := (*C.CBLError)(C.malloc(C.sizeof_CBLError))
	defer C.free(unsafe.Pointer(err))
	c_name := C.CString(name)
	result := bool(C.CBLDatabase_DeleteIndex(db.db, c_name, err))
	return result
}

/** Returns the names of the indexes on this database, as an array of strings.
    @note  You are responsible for releasing the returned Fleece array. */
// FLMutableArray CBLDatabase_IndexNames(CBLDatabase *db _cbl_nonnull) CBLAPI;
func (db *Database) IndexNames() []string {
	fl_mutable_arr := C.CBLDatabase_IndexNames(db.db)
	var iter C.FLArrayIterator
	C.FLArrayIterator_Begin(fl_mutable_arr, &iter);

	numTimes := C.FLArrayIterator_GetCount(&iter)
	indexes := make([]string, uint32(numTimes))
	i := 0

	value := C.FLArrayIterator_GetValue(&iter)
	for value != nil {
		fl_string := C.FLValue_AsString(value)
		idx_name := C.GoStringN((*C.char)(fl_string.buf), C.int(fl_string.size))
		indexes[i] = idx_name
		i++

		C.FLArrayIterator_Next(&iter);
		value = C.FLArrayIterator_GetValue(&iter)  
	}
	return indexes
}

/** @} */