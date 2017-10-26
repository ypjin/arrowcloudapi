package dao

import (
	"arrowcloudapi/models"
	"arrowcloudapi/mongo"

	"gopkg.in/mgo.v2/bson"
)

var USERS_COLLECTION = "arrowcloud:users"

// GetUser ...
func GetUser(query models.User) (*models.User, error) {

	dbQ := bson.M{}

	if query.ID != "" {
		dbQ["_id"] = bson.ObjectIdHex(query.ID)
	}
	if query.Username != "" {
		dbQ["username"] = query.Username
	}

	result, err := mongo.FindOneDocument(USERS_COLLECTION, dbQ)

	if err != nil {
		return nil, err
	}

	user := &models.User{
		ID:        result["_id"].(bson.ObjectId).Hex(),
		Username:  result["username"].(string),
		Email:     result["email"].(string),
		Firstname: result["firstname"].(string),
	}

	if result["lastname"] != nil {
		user.Lastname = result["lastname"].(string)
	}

	/*
		"orgs_360" : [
			{
				"id" : "100001450",
				"name" : "jgo@appcelerator.com",
				"admin" : true,
				"node_acs_admin" : true
			}
		],
	*/
	orgs := []models.Org{}
	mapOrgs := result["orgs_360"].([]interface{})
	for _, mapOrg := range mapOrgs {
		bsonMOrg := mapOrg.(bson.M)
		org := models.Org{
			ID:             bsonMOrg["id"].(string),
			Name:           bsonMOrg["name"].(string),
			Admin:          bsonMOrg["admin"].(bool),
			Node_acs_admin: bsonMOrg["node_acs_admin"].(bool),
		}
		orgs = append(orgs, org)
	}

	user.Orgs = orgs

	return user, nil
}
