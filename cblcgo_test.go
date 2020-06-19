// +build !replication

package cblcgo

import "testing"
import "fmt"
import "context"
import "time"

func TestConnection(t *testing.T) {
	var config DatabaseConfiguration

	var encryption_key EncryptionKey
	encryption_key.Algorithm = EncryptionNone
	encryption_key.Bytes = make([]byte, 0)

	config.Directory = "./db"
	config.EncryptionKey = encryption_key
	config.Flags = Database_Create

	if db, err := Open("my_db", &config); err == nil {
		if !DatabaseExists("my_db", "./db") {
			t.Error("Database doesn't exist.")
		}
	
		if !CopyDatabase("./db/my_db.cblite2", "my_db2", &config) {
			t.Error("Can't copy database.")
		}
	
		if DatabaseExists("my_db2", "./db") && !DeleteDatabase("my_db2", "./db") {
			t.Error("Can't delete database.")
		}
	
		if !db.Close() {
			t.Error("Couldn't close the database.")
		}
	} else {
		t.Error(err)
	}
}

func TestSaveAndDeleteDocuments(t *testing.T) {
	var config DatabaseConfiguration

	var encryption_key EncryptionKey
	encryption_key.Algorithm = EncryptionNone
	encryption_key.Bytes = make([]byte, 0)

	config.Directory = "./db"
	config.EncryptionKey = encryption_key
	
	config.Flags = Database_Create

	if db, db_err := Open("my_db3", &config); db_err == nil {

		doc_to_delete := NewDocument()
		doc_to_delete.Props["name"] = "Luke"
		doc_to_delete.Props["lastname"] = "Skywalker"
		doc_to_delete.Props["age"] = 30
		doc_to_delete.Props["email"] = "son.of.skywalker@gmail.com"
		doc_to_delete.Props["action"] = "delete"

		doc_to_purge := NewDocument()
		doc_to_purge.Props["name"] = "Anakin"
		doc_to_purge.Props["lastname"] = "Skywalker"
		doc_to_purge.Props["age"] = 30
		doc_to_purge.Props["email"] = "d.chosen.1@gmail.com"
		doc_to_delete.Props["action"] = "purge"

		doc_to_purge_by_id := NewDocumentWithId("myId")
		doc_to_purge_by_id.Props["name"] = "Obi-Wan"
		doc_to_purge_by_id.Props["lastname"] = "Kenobi"
		doc_to_purge_by_id.Props["age"] = 30
		doc_to_purge_by_id.Props["email"] = "obi.1.kenobi@gmail.com"
		doc_to_delete.Props["action"] = "purge_id"

		doc_to_expire := NewDocument()
		doc_to_expire.Props["name"] = "Master"
		doc_to_expire.Props["lastname"] = "Yoda"
		doc_to_expire.Props["age"] = 900
		doc_to_expire.Props["email"] = "yoda@aol.com"
		doc_to_delete.Props["action"] = "expire"

		var docs []*Document = []*Document{doc_to_delete, doc_to_purge, doc_to_purge_by_id, doc_to_expire}

		for i:=0; i < len(docs); i++ {
			if _doc, err := db.Save(docs[i], LastWriteWins); err == nil {
			
				switch _doc.Props["action"] {
				case "delete":
					if e := db.DeleteDocument(_doc, FailOnConflict); e != nil {
						t.Error(e)
					}
					// Release the underlying CBLDocument ptr
					if !_doc.Release() || !docs[i].Release() {
						t.Error("Error releasing a document.")
					}
					break;
				case "purge":
					if e := db.Purge(_doc); e != nil {
						t.Error(e)
					}
					// Release the underlying CBLDocument ptr
					if !_doc.Release() || !docs[i].Release() {
						t.Error("Error releasing a document.")
					}
					break;
				case "purge_id":
					if e := db.PurgeById("myId"); e != nil {
						t.Error(e)
					}
					break;
				case "expire":
					if db.SetDocumentExpiration(_doc.Id(), 1571517982) {
						if db.PurgeExpiredDocuments() <= 0 {
							t.Error("Couldn't purge expired documents.")
						}
					} else {
						t.Error("Couldn't set document expiration.")
					}
					break;
				}

			} else {
				t.Error(err)
			}
		}

		if !db.Close() {
			t.Error("Couldn't close the database.")
		}

	} else {
		t.Error(db_err)
	}

}

func TestSaveAndRetrieveDocuments(t *testing.T) {

	var config DatabaseConfiguration

	var encryption_key EncryptionKey
	encryption_key.Algorithm = EncryptionNone
	encryption_key.Bytes = make([]byte, 0)

	config.Directory = "./db"
	config.EncryptionKey = encryption_key
	
	config.Flags = Database_Create

	if db, db_err := Open("my_db4", &config); db_err == nil {

		doc := NewDocumentWithId("test")
		doc.SetPropertiesAsJSON("{\"name\": \"Marcel\", \"lastname\": \"Rivera\", \"age\": 30, \"email\": \"marcel.rivera@gmail.com\"}")
		// Save the doc, returns the same doc so only release one reference at the end.
		if _doc, err := db.Save(doc, LastWriteWins); err == nil {
			// Retrieve the saved doc
			if savedDoc, e := db.GetMutableDocument("test"); e == nil {
				for k, v := range savedDoc.Props {
					if _doc.Props[k] != v {
						t.Error("Saved document and retrieved document are different.")
					}
				}
				savedDoc.Release()
			} else {
				t.Error(e)
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

func TestProperties(t *testing.T) {
	var config DatabaseConfiguration

	var encryption_key EncryptionKey
	encryption_key.Algorithm = EncryptionNone
	encryption_key.Bytes = make([]byte, 0)

	config.Directory = "./db"
	config.EncryptionKey = encryption_key
	
	config.Flags = Database_Create

	if db, db_err := Open("my_db5", &config); db_err == nil {

		original := NewDocumentWithId("test2")
		original.Props["name"] = "Kylo"
		original.Props["lastname"] = "Ren"
		original.Props["age"] = 30
		original.Props["email"] = "son.of.solo@gmail.com"

		newProps := make(map[string]interface{})
		newProps["name"] = "Rey"
		newProps["lastname"] = ""
		newProps["age"] = 25
		newProps["email"] = "son.of.none@gmail.com"

		cpy := DocumentMutableCopy(original)
		if cpy.SetProperties(newProps) {
			for k, v := range cpy.Props {
				if original.Props[k] == v {
					t.Error("Original document and copied document have a similar property.")
				}
			}
			original.Release()
			cpy.Release()
		} else {
			t.Error("Couldn't set the properties with new map.")
		}

		if !db.Close() {
			t.Error("Couldn't close the database.")
		}
		
	} else {
		t.Error(db_err)
	}
}

func TestDocumentListener(t *testing.T) {
	var config DatabaseConfiguration

	var encryption_key EncryptionKey
	encryption_key.Algorithm = EncryptionNone
	encryption_key.Bytes = make([]byte, 0)

	config.Directory = "./db"
	config.EncryptionKey = encryption_key
	
	config.Flags = Database_Create

	if db, db_err := Open("my_db6", &config); db_err == nil {

		doc := NewDocumentWithId("documentToListenTo")
		doc.SetPropertiesAsJSON("{\"name\": \"Marcel\", \"lastname\": \"Rivera\", \"age\": 30, \"email\": \"marcel.rivera@gmail.com\"}")
		// Save the doc, returns the same doc so only release one reference at the end.
		if _, err := db.Save(doc, LastWriteWins); err == nil {
			ctx := context.WithValue(context.Background(), "package", "cblcgo")
			ctx = context.WithValue(ctx, uuid, "myUniqueIDGlobaly")
			// Create and set the listener
			var documentChangeCallback = func (ctx context.Context, db *Database, docId string) {
				fmt.Println("I'm in callback")
				fmt.Println(ctx.Value("package").(string))
			}
			if token, ee := db.AddDocumentChangeListener(documentChangeCallback, "documentToListenTo", ctx, []string{"package", uuid}); ee == nil {
				// Change and save the document
				time.Sleep(2 * time.Second)
				doc.Props["name"] = "test"
				if _, e := db.Save(doc, LastWriteWins); e != nil {
					t.Error(e)
				}
				time.Sleep(2 * time.Second)
				db.RemoveListener(token)
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

func TestQuery(t *testing.T) {

	var config DatabaseConfiguration

	var encryption_key EncryptionKey
	encryption_key.Algorithm = EncryptionNone
	encryption_key.Bytes = make([]byte, 0)

	config.Directory = "./db"
	config.EncryptionKey = encryption_key
	
	config.Flags = Database_Create

	if db, db_err := Open("my_db7", &config); db_err == nil {

		// Create an index
		var spec IndexSpec
		spec.IgnoreAccents = true
		spec.KeyExpressionsJSON = "[\"name\"]"
		spec.Language = "en"
		spec.Type = ValueIndex

		if db.CreateIndex("myFirstIndex", spec) {

			indexes := db.IndexNames()
			//fmt.Println(indexes)
			if indexes[0] != "myFirstIndex" {
				t.Error("Created index not in db.")
			}

			doc := NewDocumentWithId("docToQuery")
			doc.SetPropertiesAsJSON("{\"name\": \"Marcel\", \"lastname\": \"Rivera\", \"age\": 30, \"email\": \"marcel.rivera@gmail.com\"}")
			// Save the doc, returns the same doc so only release one reference at the end.
			if _, err := db.Save(doc, LastWriteWins); err != nil {
				t.Error(err)
			}
			doc.Release()

			if query, qerr := db.NewQuery(N1QLLanguage, "SELECT COUNT(1) where name = $name"); qerr == nil {

				// Add parameter
				queryParam := make(map[string]interface{})
				queryParam["name"] = "Marcel"
				if perr := query.SetParameters(queryParam); perr == nil {
					fmt.Println(query.Explain())
					if resultSet, eerr := query.Execute(); eerr == nil {
						query.Release()
						for resultSet.Next() {
							if val, ok := resultSet.ValueAtIndex(0).(int64); !ok || val != 1 {
								t.Error("Queried name doesn't equal expected result.")
							}
						}
					} else {
						t.Error(eerr)
					}
				} else {
					t.Error(perr)
				}
			} else {
				t.Error(qerr)
			}

			if !db.DeleteIndex("myFirstIndex") {
				t.Error("Couldn't delete index.")
			}

		} else {
			t.Error("Couldn't Create Index.")
		}

		if !db.Close() {
			t.Error("Couldn't close the database.")
		}

	} else {
		t.Error(db_err)
	}
}

func TestBlob(t *testing.T) {
	var config DatabaseConfiguration

	var encryption_key EncryptionKey
	encryption_key.Algorithm = EncryptionNone
	encryption_key.Bytes = make([]byte, 0)

	config.Directory = "./db"
	config.EncryptionKey = encryption_key
	
	config.Flags = Database_Create

	if db, db_err := Open("my_db8", &config); db_err == nil {

		doc := NewDocumentWithId("docBlob")
		doc.SetPropertiesAsJSON("{\"name\": \"Marcel\", \"lastname\": \"Rivera\", \"age\": 30, \"email\": \"marcel.rivera@gmail.com\"}")

		if blob, berr := NewBlobWithData("", []byte("This is a test")); berr == nil {
			fmt.Println("New Blob")
			fmt.Println(blob)
			doc.Props["blob"] = blob		
			// Save the doc, returns the same doc so only release one reference at the end.
			if _, err := db.Save(doc, LastWriteWins); err == nil {
				blob.Release()
				doc.Release()
				// Read the document back
				if docBlob, dblobErr := db.GetMutableDocument("docBlob"); dblobErr == nil {
					if iblob, ok := docBlob.Props["blob"]; ok {
						var b *Blob

						if b, ok = iblob.(*Blob); ok {
							// Make a stream
							brs := b.NewReadStream()
							// make destination buffer
							dst := make([]byte, b.Length())
							var totalBytes int = 0
							for uint64(totalBytes) < b.Length() {
								if bytesRead, brerr := blob.Read(brs, dst); brerr == nil {
									totalBytes += bytesRead
								} else {
									t.Error(brerr)
								}
							}

							if  "This is a test" != string(dst) {
								t.Error("Blob content doesn't match.")
							}

							b.CloseReader(brs)
						} else {
							t.Error("Couln't convert var iblob to type *Blob")
						}
					} else {
						t.Error("No entry for key \"blob\"")
					}
					// If you get a doc with a blob and try to release the doc after releasing the blob it will cause and error
					// or it could be because of the if's
					docBlob.Release()
				} else {
					t.Error(dblobErr)
				}
			} else {
				t.Error(err)
			}
		} else {
			t.Error(berr)
		}

		if !db.Close() {
			t.Error("Couldn't close the database.")
		}

	} else {
		t.Error(db_err)
	}
}

func TestListeners(t *testing.T) {
	var config DatabaseConfiguration

	var encryption_key EncryptionKey
	encryption_key.Algorithm = EncryptionNone
	encryption_key.Bytes = make([]byte, 0)

	config.Directory = "./db"
	config.EncryptionKey = encryption_key
	
	config.Flags = Database_Create

	if db, db_err := Open("my_db9", &config); db_err == nil {

		// Save the doc, returns the same doc so only release one reference at the end.
		ctx := context.WithValue(context.Background(), "package", "cblcgo")
		ctx = context.WithValue(ctx, uuid, "myUniqueIDGlobaly")
		// set a whole bunch of listeners
		// Database listener
		db_callback := func(ctx context.Context, db *Database, docIDs []string) {
			fmt.Println(len(docIDs))
			fmt.Println("db callback")
		}
		// This will be a live query that will be called when the result set changes.
		query_callback := func(ctx context.Context, query *Query) {
			fmt.Println("query callback")
		}

		if db_token, dberr := db.AddDatabaseChangeListener(db_callback, ctx, []string{"package", uuid}); dberr == nil {
			if query, qerr := db.NewQuery(N1QLLanguage, "SELECT COUNT(1) where name = $name"); qerr == nil {

				if query_token, qerr2 := query.AddChangeListener(query_callback, ctx, []string{"package", uuid}); qerr2 == nil {
					
					doc := NewDocumentWithId("docListener")
					doc.SetPropertiesAsJSON("{\"name\": \"Marcel\", \"lastname\": \"Rivera\", \"age\": 30, \"email\": \"marcel.rivera@gmail.com\"}")
					if _, err := db.Save(doc, LastWriteWins); err == nil {
						time.Sleep(3 * time.Second)
						doc.Release()
						db.RemoveListener(db_token)
						db.RemoveListener(query_token)
					}

				} else {
					t.Error(qerr2)
				}
			} else {
				t.Error(qerr)
			}	
		} else {
			t.Error(dberr)
		}

		if !db.Close() {
			t.Error("Couldn't close the database.")
		}

	} else {
		t.Error(db_err)
	}
}

func TestNotificationCallback(t *testing.T) {
	var config DatabaseConfiguration

	var encryption_key EncryptionKey
	encryption_key.Algorithm = EncryptionNone
	encryption_key.Bytes = make([]byte, 0)

	config.Directory = "./db"
	config.EncryptionKey = encryption_key
	
	config.Flags = Database_Create

	if db, db_err := Open("my_db10", &config); db_err == nil {

		doc := NewDocumentWithId("notifCallback")
		doc.SetPropertiesAsJSON("{\"name\": \"Marcel\", \"lastname\": \"Rivera\", \"age\": 30, \"email\": \"marcel.rivera@gmail.com\"}")
		// Save the doc, returns the same doc so only release one reference at the end.
		if _, err := db.Save(doc, LastWriteWins); err == nil {
			ctx := context.WithValue(context.Background(), "package", "cblcgo")
			ctx = context.WithValue(ctx, uuid, "myUniqueIDGlobaly")
			// Create and set the listener
			var documentChangeCallback = func (ctx context.Context, db *Database, docId string) {
				fmt.Println("I'm in callback")
				fmt.Println(ctx.Value("package").(string))
			}
			if token, ee := db.AddDocumentChangeListener(documentChangeCallback, "notifCallback", ctx, []string{"package", uuid}); ee == nil {
				
				// Notification change listener
				notif_callback := func (ctx context.Context, db *Database) {
					fmt.Println("notification callback")
					db.SendNotifications()
				}
				ctx2 := context.WithValue(context.Background(), "package", "cblcgo")
				db.DatabaseBufferNotifications(notif_callback, ctx2, []string{"package"})

				// Change and save the document
				time.Sleep(2 * time.Second)
				doc.Props["name"] = "test"
				if _, e := db.Save(doc, LastWriteWins); e != nil {
					t.Error(e)
				}
				time.Sleep(4 * time.Second)
				db.RemoveListener(token)
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
