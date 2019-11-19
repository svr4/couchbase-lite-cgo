package cblcgo
/*
#cgo LDFLAGS: -L. -lCouchbaseLiteC
#include <stdlib.h>
#include <stdio.h>
#include "include/CouchbaseLite.h"

bool pushFilterBridge(void *, CBLDocument*, bool);
bool pullFilterBridge(void *, CBLDocument*, bool);
void replicatorChangeBridge(void *, CBLReplicator*, CBLReplicatorStatus*);
void replicatedDocumentBridge(void *, CBLReplicator*, bool, unsigned, CBLReplicatedDocument*);
const CBLDocument * conflictResolverBridge(void *, const char *, const CBLDocument *, const CBLDocument *);
void Set_Null(void *);

bool gatewayPushFilterCallback(void *context, CBLDocument* doc, bool isDeleted) {
	return pushFilterBridge(context, doc, isDeleted);
}

bool gatewayPullFilterCallback(void *context, CBLDocument* doc, bool isDeleted) {
	return pullFilterBridge(context, doc, isDeleted);
}

void gatewayReplicatorChangeCallback(void *context, CBLReplicator *replicator _cbl_nonnull, 
									const CBLReplicatorStatus *status _cbl_nonnull) {
	replicatorChangeBridge(context, replicator, (CBLReplicatorStatus*)status);
}

void gatewayReplicatedDocumentCallback(void *context, CBLReplicator *replicator _cbl_nonnull, bool isPush,
									unsigned numDocuments, const CBLReplicatedDocument* documents) {
	replicatedDocumentBridge(context, replicator, isPush, numDocuments, (CBLReplicatedDocument*)documents);
}

const CBLDocument* gatewayConflictResolverCallback(void *context, const char *documentID, const CBLDocument *localDocument, const CBLDocument *remoteDocument) {
	return conflictResolverBridge(context, documentID, localDocument, remoteDocument);
}

void SetProxyType(CBLProxySettings * proxy, CBLProxyType type) {
	proxy->type = type;
}

*/
import "C"
import "unsafe"
import "context"
import "fmt"

/** \defgroup replication   Replication
    A replicator is a background task that synchronizes changes between a local database and
    another database on a remote server (or on a peer device, or even another local database.)
    @{ */

/** \name  Configuration
    @{ */

/** The name of the HTTP cookie used by Sync Gateway to store session keys. */
//CBL_CORE_API extern const char* kCBLAuthDefaultCookieName;
var AuthDefaultCookieName string

/** An opaque object representing the location of a database to replicate with. */
//typedef struct CBLEndpoint CBLEndpoint;
type Endpoint struct {
	endpoint *C.CBLEndpoint
}

/** Creates a new endpoint representing a server-based database at the given URL.
    The URL's scheme must be `ws` or `wss`, it must of course have a valid hostname,
    and its path must be the name of the database on that server.
    The port can be omitted; it defaults to 80 for `ws` and 443 for `wss`.
    For example: `wss://example.org/dbname` */
//CBLEndpoint* CBLEndpoint_NewWithURL(const char *url _cbl_nonnull) CBLAPI;
func NewEndpointWithURL(url string) *Endpoint {
	c_url := C.CString(url)
	c_endpoint := C.CBLEndpoint_NewWithURL(c_url)
	endpoint := Endpoint{}
	endpoint.endpoint = c_endpoint
	return &endpoint
}


// #ifdef COUCHBASE_ENTERPRISE
// /** Creates a new endpoint representing another local database. (Enterprise Edition only.) */
// CBLEndpoint* CBLEndpoint_NewWithLocalDB(CBLDatabase* _cbl_nonnull) CBLAPI;
// #endif

/** Frees a CBLEndpoint object. */
//void CBLEndpoint_Free(CBLEndpoint*) CBLAPI;


/** An opaque object representing authentication credentials for a remote server. */
//typedef struct CBLAuthenticator CBLAuthenticator;
type Authenticator struct {
	auth *C.CBLAuthenticator
}

/** Creates an authenticator for HTTP Basic (username/password) auth. */
// CBLAuthenticator* CBLAuth_NewBasic(const char *username _cbl_nonnull,
//                                    const char *password _cbl_nonnull) CBLAPI;
func NewBasicAuthentication(username, password string) *Authenticator {
	c_usr := C.CString(username)
	c_pass := C.CString(password)
	c_auth := C.CBLAuth_NewBasic(c_usr, c_pass)
	auth := Authenticator{c_auth}
	return &auth
}

/** Creates an authenticator using a Couchbase Sync Gateway login session identifier,
    and optionally a cookie name (pass NULL for the default.) */
// CBLAuthenticator* CBLAuth_NewSession(const char *sessionID _cbl_nonnull,
//                                      const char *cookieName) CBLAPI;
func NewAuthSession(sessionId, cookieName string) (*Authenticator, error) {
	c_sess := C.CString(sessionId)
	defer C.free(unsafe.Pointer(c_sess))
	c_cookie := C.CString(cookieName)
	defer C.free(unsafe.Pointer(c_cookie))
	c_auth := C.CBLAuth_NewSession(c_sess, c_cookie)
	auth := Authenticator{c_auth}
	return &auth, nil
}

func NewProxySettings(proxyType ProxyType, hostname string, port uint16, username, password string) *ProxySettings {
	p := ProxySettings{proxyType, hostname, port, username, password}
	return &p
}

/** Frees a CBLAuthenticator object. */
//void CBLAuth_Free(CBLAuthenticator*) CBLAPI;


/** Direction of replication: push, pull, or both. */
// typedef CBL_ENUM(uint8_t, CBLReplicatorType) {
//     kCBLReplicatorTypePushAndPull = 0,    ///< Bidirectional; both push and pull
//     kCBLReplicatorTypePush,               ///< Pushing changes to the target
//     kCBLReplicatorTypePull                ///< Pulling changes from the target
// };
type ReplicatorType uint8

const (
	PushAndPull ReplicatorType = iota ///< Bidirectional; both push and pull
	Push ///< Pushing changes to the target
	Pull ///< Pulling changes from the target
)

/** A callback that can decide whether a particular document should be pushed or pulled.
    @warning  This callback will be called on a background thread managed by the replicator.
                It must pay attention to thread-safety. It should not take a long time to return,
				or it will slow down the replicator.
				@param context  The `context` field of the \ref CBLReplicatorConfiguration.
			    @param document  The document in question.
			    @param isDeleted True if the document has been deleted.
			    @return  True if the document should be replicated, false to skip it. */
//typedef bool (*CBLReplicationFilter)(void *context, CBLDocument* document, bool isDeleted);
type ReplicationFilter func (ctx context.Context, doc *Document, isDeleted bool) bool

/** Conflict-resolution callback for use in replications. This callback will be invoked
			    when the replicator finds a newer server-side revision of a document that also has local
			    changes. The local and remote changes must be resolved before the document can be pushed
			    to the server.
			    @warning  This callback will be called on a background thread managed by the replicator.
			                It must pay attention to thread-safety. However, unlike a filter callback,
			                it does not need to return quickly. If it needs to prompt for user input,
			                that's OK.
			    @param context  The `context` field of the \ref CBLReplicatorConfiguration.
			    @param documentID  The ID of the conflicted document.
			    @param localDocument  The current revision of the document in the local database,
			                or NULL if the local document has been deleted.
			    @param remoteDocument  The revision of the document found on the server,
			                or NULL if the document has been deleted on the server.
			    @return  The resolved document to save locally (and push, if the replicator is pushing.)
			        This can be the same as \p localDocument or \p remoteDocument, or you can create
			        a mutable copy of either one and modify it appropriately.
					Or return NULL if the resolution is to delete the document. */
					
// typedef const CBLDocument* (*CBLConflictResolver)(void *context,
// 							const char *documentID,
// 							const CBLDocument *localDocument,
// 							const CBLDocument *remoteDocument);

type ConflictResolver func (ctx context.Context, documentId string,
							localDocument *Document, remoteDocument *Document) *Document


/** Default conflict resolver. This always returns `localDocument`. */
// extern const CBLConflictResolver CBLDefaultConflictResolver;
var DefaultConflictResolver ConflictResolver

/** Types of proxy servers, for CBLProxySettings. */
// typedef CBL_ENUM(uint8_t, CBLProxyType) {
// 	kCBLProxyHTTP,                      ///< HTTP proxy; must support 'CONNECT' method
// 	kCBLProxyHTTPS,                     ///< HTTPS proxy; must support 'CONNECT' method
// };
type ProxyType uint8

const (
	ProxyHTTP ProxyType = iota ///< HTTP proxy; must support 'CONNECT' method
	ProxyHTTPS ///< HTTPS proxy; must support 'CONNECT' method
)

// /** Proxy settings for the replicator. */
// typedef struct {
// 	CBLProxyType type;                  ///< Type of proxy
// 	const char *hostname;               ///< Proxy server hostname or IP address
// 	uint16_t port;                      ///< Proxy server port
// 	const char *username;               ///< Username for proxy auth (optional)
// 	const char *password;               ///< Password for proxy auth
// } CBLProxySettings;

type ProxySettings struct {
	Type ProxyType		///< Type of proxy
	Hostname string		///< Proxy server hostname or IP address
	Port uint16			///< Proxy server port
	Username string		///< Username for proxy auth (optional)
	Password string 	///< Password for proxy auth
}


/** The configuration of a replicator. */
// typedef struct {
//     CBLDatabase* database;              ///< The database to replicate
//     CBLEndpoint* endpoint;              ///< The address of the other database to replicate with
//     CBLReplicatorType replicatorType;   ///< Push, pull or both
//     bool continuous;                    ///< Continuous replication?
//-- HTTP settings:
//     CBLAuthenticator* authenticator;    ///< Authentication credentials, if needed
// New:
//     const CBLProxySettings* proxy;      ///< HTTP client proxy settings
//     FLDict headers;                     ///< Extra HTTP headers to add to the WebSocket request
//-- TLS settings:
//     FLSlice pinnedServerCertificate;    ///< An X.509 cert to "pin" TLS connections to (PEM or DER)
// New:
//     FLSlice trustedRootCertificates;    ///< Set of anchor certs (PEM format)
//-- Filtering:
//     FLArray channels;                   ///< Optional set of channels to pull from
//     FLArray documentIDs;                ///< Optional set of document IDs to replicate
//     CBLReplicationFilter pushFilter;    ///< Optional callback to filter which docs are pushed
//     CBLReplicationFilter pullFilter;    ///< Optional callback to validate incoming docs
// New
//     CBLConflictResolver conflictResolver;///< Optional conflict-resolver callback
//     void* context;                      ///< Arbitrary value that will be passed to callbacks
// } CBLReplicatorConfiguration;

type ReplicatorConfiguration struct {
	Db *Database
	Endpt *Endpoint
	Replicator ReplicatorType
	Continious bool
	Auth *Authenticator
	Proxy *ProxySettings
	PinnedServerCertificate []byte
	TrustedRootCertificates []byte
	Headers map[string]interface{}
	Channels []string
	DocumentIds []string
	PushFilter ReplicationFilter
	PullFilter ReplicationFilter
	Resolver ConflictResolver
	FilterContext context.Context
	FilterKeys []string
}

/** @} */

type Replicator struct {
	rep *C.CBLReplicator
}

/** \name  Lifecycle
    @{ */

// CBL_REFCOUNTED(CBLReplicator*, Replicator);

/** Creates a replicator with the given configuration. */
// CBLReplicator* CBLReplicator_New(const CBLReplicatorConfiguration* _cbl_nonnull,
//                                  CBLError*) CBLAPI;
func NewReplicator(config ReplicatorConfiguration) (*Replicator, error) {
	err := (*C.CBLError)(C.malloc(C.sizeof_CBLError))
	defer C.free(unsafe.Pointer(err))
	c_config := (*C.CBLReplicatorConfiguration)(C.malloc(C.sizeof_CBLReplicatorConfiguration))

	c_config.database = config.Db.db
	c_config.endpoint = config.Endpt.endpoint
	c_config.replicatorType = C.CBLReplicatorType(config.Replicator)
	c_config.continuous = C.bool(config.Continious)
	c_config.authenticator = config.Auth.auth

	// Proxy Settings
	if config.Proxy != nil {
		proxy := (*C.CBLProxySettings)(C.malloc(C.sizeof_CBLProxySettings))
		// I use this function because Go thinks proxy.type is a type assertion.
		C.SetProxyType(proxy, C.CBLProxyType(config.Proxy.Type))
		proxy.hostname = C.CString(config.Proxy.Hostname)
		proxy.port = C.uint16_t(config.Proxy.Port)
		proxy.username = C.CString(config.Proxy.Username)
		proxy.password = C.CString(config.Proxy.Password)

		c_config.proxy = proxy
	} else {
		C.Set_Null(unsafe.Pointer(c_config.proxy))
	}

	if len(config.PinnedServerCertificate) > 0 {
		certSize := len(config.PinnedServerCertificate)
		certBytes := C.CBytes(config.PinnedServerCertificate)
		c_config.pinnedServerCertificate = C.FLSlice{unsafe.Pointer(certBytes), C.size_t(certSize)}
	} else {
		c_config.pinnedServerCertificate = C.kFLSliceNull
	}


	if len(config.TrustedRootCertificates) > 0 {
		// Trusted Certificates
		trustedCertSize := len(config.TrustedRootCertificates)
		trustedCertBytes := C.CBytes(config.TrustedRootCertificates)
		c_config.trustedRootCertificates = C.FLSlice{unsafe.Pointer(trustedCertBytes), C.size_t(trustedCertSize)}
	} else {
		c_config.trustedRootCertificates = C.kFLSliceNull
	}

	// Process Headers
	if len(config.Headers) > 0 {
		mutableDict := C.FLMutableDict_New()

		for k, v := range config.Headers {
			c_key := C.CString(k)
			fl_slot := C.FLMutableDict_Set(mutableDict, C.FLStr(c_key))
			storeGoValueInSlot(fl_slot, v)
			C.free(unsafe.Pointer(c_key))
		}
		fl_dict := C.FLDict(mutableDict)
		c_config.headers = fl_dict
	} else {
		c_config.headers = C.FLDict(C.FLMutableDict_New())
	}

	// Process channels
	if len(config.Channels) > 0 {
		chan_array := C.FLMutableArray_New()
		for i:=0; i < len(config.Channels); i++ {
			chan_slot := C.FLMutableArray_Append(chan_array)
			storeGoValueInSlot(chan_slot, config.Channels[i]);
		}
		c_config.channels = C.FLArray(chan_array)
	} else {
		// c_config.channels = C.kFLEmptyArray
		c_config.documentIDs = C.FLArray(C.FLMutableArray_New())
	}

	// Process documentIds
	if len(config.DocumentIds) > 0 {
		docIds_array := C.FLMutableArray_New()
		for ii:=0; ii < len(config.DocumentIds); ii++ {
			doc_slot := C.FLMutableArray_Append(docIds_array)
			storeGoValueInSlot(doc_slot, config.DocumentIds[ii]);
		}
		c_config.documentIDs = C.FLArray(docIds_array)
	} else {
		c_config.documentIDs = C.FLArray(C.FLMutableArray_New())
	}

	// The pullCallback and pushCallback keys should already be in the context.
	if config.PushFilter != nil {
		// Put the C callbacks in
		c_config.pushFilter = (C.CBLReplicationFilter)(C.gatewayPushFilterCallback)
		pushKey := config.FilterContext.Value(pushCallback).(string)
		pushFilterCallbacks[pushKey] = config.PushFilter
	} else {
		C.Set_Null(unsafe.Pointer(c_config.pushFilter))
	}

	if config.PullFilter != nil {
		// Put the C callbacks in
		c_config.pullFilter = (C.CBLReplicationFilter)(C.gatewayPullFilterCallback)
		pullKey := config.FilterContext.Value(pullCallback).(string)
		pullFilterCallbacks[pullKey] = config.PullFilter
	} else {
		C.Set_Null(unsafe.Pointer(c_config.pullFilter))
	}

	// Place the context into a mutable dict.
	if config.FilterContext != nil && len(config.FilterKeys) > 0 {
		dict := storeContextInMutableDict(config.FilterContext, config.FilterKeys)
		c_config.context = unsafe.Pointer(dict)
	} else {
		C.Set_Null(unsafe.Pointer(c_config.context))
	}

	// Conflict Resolver callback
	if config.Resolver != nil {
		c_config.conflictResolver = (C.CBLConflictResolver)(C.gatewayConflictResolverCallback)
		conflictKey := config.FilterContext.Value(conflictResolver).(string)
		conflictResolverCallbacks[conflictKey] = config.Resolver
	} else {
		C.Set_Null(unsafe.Pointer(c_config.conflictResolver))
	}

	c_replicator := C.CBLReplicator_New(c_config, err)
	if (*err).code == 0 {
		replicator := Replicator{c_replicator}
		return &replicator, nil
	}
	c_err_msg := C.CBLError_Message(err)
	ErrCBLInternalError = fmt.Errorf("CBL: %s. Domain: %d Code: %d", C.GoString(c_err_msg), (*err).domain, (*err).code)
	C.free(unsafe.Pointer(c_err_msg))
	return nil, ErrCBLInternalError
}





/** Returns the configuration of an existing replicator. */
// const CBLReplicatorConfiguration* CBLReplicator_Config(CBLReplicator* _cbl_nonnull) CBLAPI;

/** Instructs the replicator to ignore existing checkpoints the next time it runs.
    This will cause it to scan through all the documents on the remote database, which takes
    a lot longer, but it can resolve problems with missing documents if the client and
    server have gotten out of sync somehow. */
// void CBLReplicator_ResetCheckpoint(CBLReplicator* _cbl_nonnull) CBLAPI;
func (rep *Replicator) ResetCheckpoint(){
	C.CBLReplicator_ResetCheckpoint(rep.rep)
}

/** Starts a replicator, asynchronously. Does nothing if it's already started. */
// void CBLReplicator_Start(CBLReplicator* _cbl_nonnull) CBLAPI;
func (rep *Replicator) Start() {
	C.CBLReplicator_Start(rep.rep)
}

/** Stops a running replicator, asynchronously. Does nothing if it's not already started.
    The replicator will call your \ref CBLReplicatorChangeListener with an activity level of
    \ref kCBLReplicatorStopped after it stops. Until then, consider it still active. */
// void CBLReplicator_Stop(CBLReplicator* _cbl_nonnull) CBLAPI;
func (rep *Replicator) Stop() {
	C.CBLReplicator_Stop(rep.rep)
}

func RemovePushFilterListener(key string) {
	delete(pushFilterCallbacks, key)
}

func RemovePullFilterListener(key string) {
	delete(pullFilterCallbacks, key)
}

func (rep *Replicator) Release() {
	C.CBLReplicator_Release(rep.rep)
}
/** @} */



/** \name  Status and Progress
    @{
 */

/** The possible states a replicator can be in during its lifecycle. */
// typedef CBL_ENUM(uint8_t, CBLReplicatorActivityLevel) {
//     kCBLReplicatorStopped,    ///< The replicator is unstarted, finished, or hit a fatal error.
//     kCBLReplicatorOffline,    ///< The replicator is offline, as the remote host is unreachable.
//     kCBLReplicatorConnecting, ///< The replicator is connecting to the remote host.
//     kCBLReplicatorIdle,       ///< The replicator is inactive, waiting for changes to sync.
//     kCBLReplicatorBusy        ///< The replicator is actively transferring data.
// };
type ReplicatorActivityLevel uint8

const (
	Stopped ReplicatorActivityLevel = iota ///< The replicator is unstarted, finished, or hit a fatal error.
    Offline		///< The replicator is offline, as the remote host is unreachable.
    Connecting	///< The replicator is connecting to the remote host.
    Idle			///< The replicator is inactive, waiting for changes to sync.
    Busy			///< The replicator is actively transferring data.
)

/** A fractional progress value. The units are undefined; the only meaningful number is the
    (fractional) result of `completed` ÷ `total`, which will range from 0.0 to 1.0.
    Before anything happens, both `completed` and `total` will be 0. */
// typedef struct {
//     float fractionComplete;     /// Very-approximate completion, from 0.0 to 1.0
//     uint64_t documentCount;     ///< Number of documents transferred so far
// } CBLReplicatorProgress;
type ReplicatorProgress struct {
	FractionComplete float32
	DocumentCount uint64
}

type Error struct {
	InternalInfo uint32
	Code uint32
	Domain uint32
} 

/** A replicator's current status. */
// typedef struct {
//     CBLReplicatorActivityLevel activity;    ///< Current state
//     CBLReplicatorProgress progress;         ///< Approximate fraction complete
//     CBLError error;                         ///< Error, if any
// } CBLReplicatorStatus;
type ReplicatorStatus struct {
	Activity ReplicatorActivityLevel
	Progress ReplicatorProgress
	Err Error
}

/** Returns the replicator's current status. */
// CBLReplicatorStatus CBLReplicator_Status(CBLReplicator* _cbl_nonnull) CBLAPI;
func (rep *Replicator) Status() ReplicatorStatus {
	c_replicator := C.CBLReplicator_Status(rep.rep)
	e := Error{uint32(c_replicator.error.internal_info), uint32(c_replicator.error.code), uint32(c_replicator.error.domain)}
	activity := ReplicatorActivityLevel(c_replicator.activity)
	progress := ReplicatorProgress{float32(c_replicator.progress.fractionComplete), uint64(c_replicator.progress.documentCount)}
	repStatus := ReplicatorStatus{activity, progress, e}
	return repStatus
}


/** A callback that notifies you when the replicator's status changes.
    @warning  This callback will be called on a background thread managed by the replicator.
                It must pay attention to thread-safety. It should not take a long time to return,
                or it will slow down the replicator.
    @param context  The value given when the listener was added.
    @param replicator  The replicator.
    @param status  The replicator's status. */
// typedef void (*CBLReplicatorChangeListener)(void *context, 
                                            // CBLReplicator *replicator _cbl_nonnull,
											// const CBLReplicatorStatus *status _cbl_nonnull);
type ReplicatorChangeListener func(ctx context.Context, replicator *Replicator, status *ReplicatorStatus)

/** Adds a listener that will be called when the replicator's status changes.
    @warning UNIMPLEMENTED! */
// CBLListenerToken* CBLReplicator_AddChangeListener(CBLReplicator* _cbl_nonnull,
//                                                   CBLReplicatorChangeListener _cbl_nonnull, 
//                                                   void *context) CBLAPI;

func (rep *Replicator) AddChangeListener(listener ReplicatorChangeListener, ctx context.Context, ctxKeys []string) (*ListenerToken, error) {
	if v := ctx.Value(uuid); v != nil {
		key, ok := v.(string)
		if ok {
			replicatorCallbacks[key] = listener
			mutableDictContext := storeContextInMutableDict(ctx, ctxKeys)
			token := C.CBLReplicator_AddChangeListener(rep.rep,
				(C.CBLReplicatorChangeListener)(C.gatewayReplicatorChangeCallback), unsafe.Pointer(mutableDictContext))			
			listener_token := ListenerToken{key,token,"ReplicatorChangeListener"}
			return &listener_token, nil
		}
	}
	ErrCBLInternalError = fmt.Errorf("CBL: No UUID present in context.")
	return nil, ErrCBLInternalError
}

/** Flags describing a replicated document. */
// typedef CBL_ENUM(unsigned, CBLDocumentFlags) {
//     kCBLDocumentFlagsDeleted        = 1 << 0,   ///< The document has been deleted.
//     kCBLDocumentFlagsAccessRemoved  = 1 << 1    ///< Lost access to the document on the server.
// };
type DocumentFlags uint

const (
	DocumentFlagsDeleted DocumentFlags = 1 << iota   ///< The document has been deleted.
	DocumentFlagsAccessRemoved    ///< Lost access to the document on the server.
)


/** Information about a document that's been pushed or pulled. */
// typedef struct {
//     const char *ID;             ///< The document ID
//     CBLDocumentFlags flags;     ///< Indicates whether the document was deleted or removed
//     CBLError error;             ///< If the code is nonzero, the document failed to replicate.
// } CBLReplicatedDocument;
type ReplicatedDocument struct {
	ID string
	Flags DocumentFlags
	Err Error
}


/** A callback that notifies you when documents are replicated.
    @warning  This callback will be called on a background thread managed by the replicator.
                It must pay attention to thread-safety. It should not take a long time to return,
                or it will slow down the replicator.
    @param context  The value given when the listener was added.
    @param replicator  The replicator.
    @param isPush  True if the document(s) were pushed, false if pulled.
    @param numDocuments  The number of documents reported by this callback.
    @param documents  An array with information about each document. */
// typedef void (*CBLReplicatedDocumentListener)(void *context,
//                                               CBLReplicator *replicator _cbl_nonnull,
//                                               bool isPush,
//                                               unsigned numDocuments,
//                                               const CBLReplicatedDocument* documents);
type ReplicatedDocumentListener func(ctx context.Context, replicator *Replicator,
									isPush bool, numDocuments uint, documents *ReplicatedDocument)

/** Adds a listener that will be called when documents are replicated.
    @warning UNIMPLEMENTED! */
// CBLListenerToken* CBLReplicator_AddDocumentListener(CBLReplicator* _cbl_nonnull,
//                                                     CBLReplicatedDocumentListener _cbl_nonnull,
//                                                     void *context) CBLAPI;
func (rep *Replicator) AddDocumentListener(listener ReplicatedDocumentListener, ctx context.Context, ctxKeys []string) (*ListenerToken, error) {
	if v := ctx.Value(uuid); v != nil {
		key, ok := v.(string)
		if ok {
			replicatedDocCallbacks[key] = listener
			mutableDictContext := storeContextInMutableDict(ctx, ctxKeys)
			token := C.CBLReplicator_AddDocumentListener(rep.rep,
				(C.CBLReplicatedDocumentListener)(C.gatewayReplicatedDocumentCallback), unsafe.Pointer(mutableDictContext))			
			listener_token := ListenerToken{key,token,"ReplicatedDocumentListener"}
			return &listener_token, nil
		}
	}
	ErrCBLInternalError = fmt.Errorf("CBL: No UUID present in context.")
	return nil, ErrCBLInternalError
}