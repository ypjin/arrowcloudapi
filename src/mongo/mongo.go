package mongo

import (
	"utils/log"
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
		log.Errorf("query err: %v", err)
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
		log.Errorf("findAndModify err: %v", err)
	}

	return

}
