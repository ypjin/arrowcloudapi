package mongo

import (
	"arrowcloudapi/utils/log"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//http://goinbigdata.com/how-to-build-microservice-with-mongodb-in-golang/
//https://godoc.org/gopkg.in/mgo.v2/bson#M
func FindOneDocument(collectionPrefixedWithDB string, query bson.M) (result bson.M, err error) {

	dBName, collectionName := getDBAndCollectionName(collectionPrefixedWithDB)

	session := GetSession()
	defer session.Close()

	result = bson.M{}

	c := session.DB(dBName).C(collectionName)

	//https://godoc.org/gopkg.in/mgo.v2#Collection.Find
	err = c.Find(query).One(&result)

	if err != nil {
		log.Errorf("Query err: %v", err)
	}

	return
}

func FindDocuments(collectionPrefixedWithDB string, query bson.M) (result []bson.M, err error) {

	dBName, collectionName := getDBAndCollectionName(collectionPrefixedWithDB)

	session := GetSession()
	defer session.Close()

	result = []bson.M{}

	c := session.DB(dBName).C(collectionName)

	//https://godoc.org/gopkg.in/mgo.v2#Collection.Find
	err = c.Find(query).All(&result)

	if err != nil {
		log.Errorf("Query err: %v", err)
	}

	return
}

/*
  {
      "lastErrorObject": {
          "updatedExisting": true,
          "n": 1
      },
      "value": {
          "_id": "57de1cfde15f77045fdaa39a",
          "guid": "132203bc-e6dc-460f-b46d-5c9f34464730",
          "email": "yjin@appcelerator.com",
          "username": "yjin@appcelerator.com",
          "firstname": "Yuping",
          "orgs_360": [
              {
                  "id": "14301",
                  "name": "appcelerator Inc.",
                  "admin": true,
                  "node_acs_admin": true
              }
          ],
          "orgs_360_updated_at": "2016-10-08T08:27:05.520Z"
      },
      "ok": 1
  }
*/
func UpsertDocument(collectionPrefixedWithDB string, query, update bson.M) (result bson.M, err error) {

	dBName, collectionName := getDBAndCollectionName(collectionPrefixedWithDB)

	session := GetSession()
	defer session.Close()

	change := mgo.Change{
		Update:    update,
		Upsert:    true,
		ReturnNew: true,
	}

	result = bson.M{}

	c := session.DB(dBName).C(collectionName)

	//https://godoc.org/gopkg.in/mgo.v2#Query.Apply
	_, err = c.Find(query).Apply(change, &result)

	if err != nil {
		log.Errorf("FindAndModify err: %v", err)
	}

	//log.Debugf("findAndModify result: %v", result)
	return
}

func InsertDocument(collectionPrefixedWithDB string, doc bson.M) (err error) {

	dBName, collectionName := getDBAndCollectionName(collectionPrefixedWithDB)

	session := GetSession()
	defer session.Close()

	c := session.DB(dBName).C(collectionName)

	err = c.Insert(doc)

	if err != nil {
		log.Errorf("Insert err: %v", err)
	}

	return err
}

func RemoveDocument(collectionPrefixedWithDB string, query bson.M) (err error) {

	dBName, collectionName := getDBAndCollectionName(collectionPrefixedWithDB)

	session := GetSession()
	defer session.Close()

	c := session.DB(dBName).C(collectionName)

	//https://godoc.org/gopkg.in/mgo.v2#Collection.Remove
	err = c.Remove(query)

	if err != nil {
		log.Errorf("Remove err: %v", err)
	}

	return
}

func RemoveDocuments(collectionPrefixedWithDB string, query bson.M) (err error) {

	dBName, collectionName := getDBAndCollectionName(collectionPrefixedWithDB)

	session := GetSession()
	defer session.Close()

	c := session.DB(dBName).C(collectionName)

	_, err = c.RemoveAll(query)

	if err != nil {
		log.Errorf("RemoveAll err: %v", err)
	}

	return
}
