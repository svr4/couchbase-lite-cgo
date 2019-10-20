package cblcgo

import "testing"
// import "fmt"

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

func TestDocuments(t *testing.T) {
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
					break;
				case "purge":
					if e := db.Purge(_doc); e != nil {
						t.Error(e)
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