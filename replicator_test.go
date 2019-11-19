// +build replication

package cblcgo

import "testing"
import "fmt"
import "context"
import "os"

/*
Notes on testing replication:

* I separeted replicator tests from the rest because it's easier to perform them this way.
* To run the test make sure couchbase server and sync gateway are up and running.
* Make sure the sync gateway is properly configured.
* Continous push/pull doesn't support document Id's.
* It's best to work with separate replicators for push and pull for now.
* PushPull isn't working properly for me, I need to investigate further.
* It's only pulling when no DocId's are defined and pull with no DocId's defined.
*/

func TestReplicationHTTP(t *testing.T) {
	var config DatabaseConfiguration

	// var encryption_key EncryptionKey
	// encryption_key.algorithm = EncryptionNone
	// encryption_key.bytes = make([]byte, 0)

	config.directory = "./db"
	// config.encryptionKey = encryption_key
	
	config.flags = Database_Create

	if db, db_err := Open("my_db11", &config); db_err == nil {

		doc := NewDocument()
		doc.SetPropertiesAsJSON("{\"name\": \"Marcel\", \"lastname\": \"Rivera\", \"age\": 30, \"email\": \"marcel.rivera@gmail.com\"}")
		
		// Save the doc, the next time you run the replicator it should sync the doc back to the bucket.
		if _, err := db.Save(doc, LastWriteWins); err != nil {
			t.Error(err)
		}


		var replicator_config ReplicatorConfiguration
		replicator_config.Db = db
		replicator_config.Endpt = NewEndpointWithURL("ws://localhost:4985/my_db11")
		replicator_config.Replicator =  PushAndPull
		replicator_config.Continious = false
		replicator_config.Auth = NewBasicAuthentication("test", "testtest")
		// No proxy settings
		// No certificates this is HTTP

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

		var replicatorChangeToken *ListenerToken
		completedSync := false

		if replicator, rerr := NewReplicator(replicator_config); rerr == nil {
			rep_listener := func(ctx context.Context, replicator *Replicator, status *ReplicatorStatus) {
				fmt.Println("Replicator Change Listener")
				switch status.Activity {
				case Idle:
					fmt.Println("Activity: Idle")
					break;
				case Offline:
					fmt.Println("Activity: Offline")
					break;
				case Busy:
					fmt.Println("Activity: Busy")
					break;
				case Stopped:
					fmt.Println("Activity: Stopped")
					break;
				case Connecting:
					fmt.Println("Activity: Connecting")
					break;
				}
				// Workaround for BUG :https://github.com/couchbase/couchbase-lite-ios/issues/1816.
				if status.Progress.FractionComplete ==  1.0 {
					fmt.Println("All documents synced")
					completedSync = true
				} else {
					 fmt.Println("Documents to synced so far: " + string(status.Progress.DocumentCount))
				}
			}
			rep_ctx := context.WithValue(context.Background(), uuid, "myReplicatorCallback")
			if change_token, r_err := replicator.AddChangeListener(rep_listener, rep_ctx, []string{uuid}); r_err == nil {
				replicatorChangeToken = change_token
				
			} else {
				t.Error(r_err)
			}

			replicator.Start()

			for !completedSync {}
			replicator.Stop()
			db.RemoveListener(replicatorChangeToken)
			replicator.Release()

		} else {
			t.Error(rerr)
		}

		if !db.Close() {
			t.Error("Couldn't close the database.")
		}

	} else {
		t.Error(db_err)
	}
}

func TestReplicationHTTPS(t *testing.T) {
	var config DatabaseConfiguration
	config.directory = "./db"
	
	config.flags = Database_Create

	if db, db_err := Open("my_db12", &config); db_err == nil {

		// doc := NewDocument()
		// doc.SetPropertiesAsJSON("{\"name\": \"Marcel\", \"lastname\": \"Rivera\", \"age\": 30, \"email\": \"marcel.rivera@gmail.com\"}")
		
		// // Save the doc, the next time you run the replicator it should sync the doc back to the bucket.
		// if _, err := db.Save(doc, LastWriteWins); err != nil {
		// 	t.Error(err)
		// }

		var replicator_config ReplicatorConfiguration
		replicator_config.Db = db
		replicator_config.Endpt = NewEndpointWithURL("wss://localhost:4985/my_db12")
		replicator_config.Replicator =  PushAndPull
		replicator_config.Continious = false
		replicator_config.Auth = NewBasicAuthentication("test", "testtest")
		// No proxy settings
		// Setup pinned certificate
		if file, ferr := os.Open("./domain.der"); ferr == nil {
			file_info, _ := file.Stat()
			cert_buff := make([]byte, file_info.Size())
			file.Read(cert_buff)
			replicator_config.PinnedServerCertificate = cert_buff
		} else {
			t.Error(ferr)
		}

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
		replicator_config.DocumentIds = []string{"~SkrJzahKTd9aapC25wRtju", "~az9TJV_eCdBEAdNu-GALkR"}
		replicator_config.Channels = []string{"foo2"}
		replicator_config.FilterContext = ctx
		replicator_config.FilterKeys = []string{"package", pushCallback, pullCallback, conflictResolver}
		replicator_config.FilterKeys = []string{"package", pushCallback, conflictResolver}

		var replicatorChangeToken *ListenerToken
		completedSync := false

		if replicator, rerr := NewReplicator(replicator_config); rerr == nil {
			rep_listener := func(ctx context.Context, replicator *Replicator, status *ReplicatorStatus) {
				fmt.Println("Replicator Change Listener")
				switch status.Activity {
				case Idle:
					fmt.Println("Activity: Idle")
					break;
				case Offline:
					fmt.Println("Activity: Offline")
					break;
				case Busy:
					fmt.Println("Activity: Busy")
					break;
				case Stopped:
					fmt.Println("Activity: Stopped")
					break;
				case Connecting:
					fmt.Println("Activity: Connecting")
					break;
				}
				// Workaround for BUG :https://github.com/couchbase/couchbase-lite-ios/issues/1816.
				if status.Progress.FractionComplete ==  1.0 {
					fmt.Println("All documents synced")
					completedSync = true
				} else {
					 fmt.Println("Documents to synced so far: " + string(status.Progress.DocumentCount))
				}
			}
			rep_ctx := context.WithValue(context.Background(), uuid, "myReplicatorCallback")
			if change_token, r_err := replicator.AddChangeListener(rep_listener, rep_ctx, []string{uuid}); r_err == nil {
				replicatorChangeToken = change_token
				
			} else {
				t.Error(r_err)
			}

			replicator.Start()

			for !completedSync {}
			replicator.Stop()
			db.RemoveListener(replicatorChangeToken)
			replicator.Release()

		} else {
			t.Error(rerr)
		}

		if !db.Close() {
			t.Error("Couldn't close the database.")
		}

	} else {
		t.Error(db_err)
	}
}