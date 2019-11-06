// +build replication

package cblcgo

import "testing"
import "fmt"
import "context"

func TestReplicationHTTP(t *testing.T) {
	var config DatabaseConfiguration

	var encryption_key EncryptionKey
	encryption_key.algorithm = EncryptionNone
	encryption_key.bytes = make([]byte, 0)

	config.directory = "./db"
	config.encryptionKey = encryption_key
	
	config.flags = Database_Create

	if db, db_err := Open("my_db10", &config); db_err == nil {

		doc := NewDocumentWithId("replicatedDoc1")
		doc.SetPropertiesAsJSON("{\"name\": \"Marcel\", \"lastname\": \"Rivera\", \"age\": 30, \"email\": \"marcel.rivera@gmail.com\"}")
		// Save the doc, returns the same doc so only release one reference at the end.
		if _, err := db.Save(doc, LastWriteWins); err == nil {
			var replicator_config ReplicatorConfiguration
			replicator_config.Db = db
			replicator_config.Endpt = NewEndpointWithURL("ws://localhost:4985/my_db")
			replicator_config.Replicator = PushAndPull
			replicator_config.Continious = true
			replicator_config.Auth = NewBasicAuthentication("test", "testtest")
			// No proxy settings
			// No certificates this is HTTP
			replicator_config.DocumentIds = []string{"replicatedDoc1"}
			replicator_config.PullFilter = func (ctx context.Context, doc *Document, isDeleted bool) bool {
				fmt.Println("Pull Filter callback")
				return true
			}

			replicator_config.PushFilter = func (ctx context.Context, doc *Document, isDeleted bool) bool {
				fmt.Println("Push Filter callback")
				return true
			}

			replicator_config.Resolver = func (ctx context.Context, documentId string,
				localDocument *Document, remoteDocument *Document) *Document {
					fmt.Println("Conflict Resolver")
					return localDocument
				}
			
			ctx := context.WithValue(context.Background(), "package", "cblcgo")
			ctx = context.WithValue(ctx, pushCallback, "myPushCallback")
			ctx = context.WithValue(ctx, pullCallback, "myPullCallback")
			ctx = context.WithValue(ctx, conflictResolver, "myResolverCallback")
			replicator_config.Channels = []string{"foo"}
			replicator_config.FilterContext = ctx
			replicator_config.FilterKeys = []string{"package", pushCallback, pullCallback, conflictResolver}


			if replicator, rerr := NewReplicator(replicator_config); rerr == nil {
				rep_listener := func(ctx context.Context, replicator *Replicator, status *ReplicatorStatus) {
					fmt.Println("Replicator Change Listener")
				}
				rep_ctx := context.WithValue(context.Background(), uuid, "myReplicatorCallback")
				if change_token, r_err := replicator.AddChangeListener(rep_listener, rep_ctx, []string{uuid}); r_err == nil {
					rep_doc_ctx := context.WithValue(context.Background(), uuid, "myReplicatorDocCallback")
					rep_doc_listener := func(ctx context.Context, replicator *Replicator,
						isPush bool, numDocuments uint, documents *ReplicatedDocument) {
							fmt.Println("Replicator Doc Listener")
						}
					if doc_token, d_err := replicator.AddDocumentListener(rep_doc_listener, rep_doc_ctx, []string{uuid}); d_err == nil {
						replicator.Start()
						// Lets edit and see if it makes it back to the server.
						doc.Props["name"] = "Lecram"
						if _, derr := db.Save(doc, LastWriteWins); derr == nil {
							doc.Release()
						}
						// Let the db sync
						//time.Sleep(8 * time.Second)
						for {}
						replicator.Stop()
						db.RemoveListener(doc_token)
					} else {
						t.Error(d_err)
					}
					db.RemoveListener(change_token)
				} else {
					t.Error(r_err)
				}
			} else {
				t.Error(rerr)
			}

		} else {
			t.Error(err)
		}

		if !db.Close() {
			t.Error("Couldn't close the database.")
		}

	} else {
		t.Error(db_err)
	}
}