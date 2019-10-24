package cblcgo

import "testing"
import "fmt"
import "context"
// import "time"

func TestConnection(t *testing.T) {
	var config DatabaseConfiguration

	var encryption_key EncryptionKey
	encryption_key.algorithm = EncryptionNone
	encryption_key.bytes = make([]byte, 0)

	config.directory = "./db"
	config.encryptionKey = encryption_key
	// sun pass & internet for credit card
	config.flags = Database_Create

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
	encryption_key.algorithm = EncryptionNone
	encryption_key.bytes = make([]byte, 0)

	config.directory = "./db"
	config.encryptionKey = encryption_key
	
	config.flags = Database_NoUpgrade

	if db, db_err := Open("my_db", &config); db_err == nil {

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
		// if doc, err := db.Save(doc_to_delete, LastWriteWins); err == nil {
		// 	fmt.Println("cblcgo_test")
		// 	fmt.Println(doc)
		// 	if e := db.DeleteDocument(doc_to_delete, LastWriteWins); e != nil {
		// 		t.Error(e)
		// 	}
		// } else {
		// 	t.Error(err)
		// }

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
	encryption_key.algorithm = EncryptionNone
	encryption_key.bytes = make([]byte, 0)

	config.directory = "./db"
	config.encryptionKey = encryption_key
	
	config.flags = Database_NoUpgrade

	if db, db_err := Open("my_db", &config); db_err == nil {

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
	encryption_key.algorithm = EncryptionNone
	encryption_key.bytes = make([]byte, 0)

	config.directory = "./db"
	config.encryptionKey = encryption_key
	
	config.flags = Database_NoUpgrade

	if db, db_err := Open("my_db", &config); db_err == nil {

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
	encryption_key.algorithm = EncryptionNone
	encryption_key.bytes = make([]byte, 0)

	config.directory = "./db"
	config.encryptionKey = encryption_key
	
	config.flags = Database_NoUpgrade

	if db, db_err := Open("my_db", &config); db_err == nil {

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
				//time.Sleep(3 * time.Second)
				doc.Props["name"] = "test"
				if _, e := db.Save(doc, LastWriteWins); e != nil {
					t.Error(e)
				}
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
	encryption_key.algorithm = EncryptionNone
	encryption_key.bytes = make([]byte, 0)

	config.directory = "./db"
	config.encryptionKey = encryption_key
	
	config.flags = Database_NoUpgrade

	if db, db_err := Open("my_db", &config); db_err == nil {

		// Create an index
		var spec IndexSpec
		spec.IgnoreAccents = true
		spec.KeyExpressionsJSON = "[\"name\"]"
		spec.Language = "en"
		spec.Type = ValueIndex

		if db.CreateIndex("myFirstIndex", spec) {

			indexes := db.IndexNames()
			//fmt.Println(indexes)
			if indexes[1] != "myFirstIndex" {
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
					if resultSet, eerr := query.Execute(); eerr == nil {
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
				query.Release()
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

}