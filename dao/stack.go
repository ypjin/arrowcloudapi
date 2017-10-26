package dao

import (
	"arrowcloudapi/models"
	"arrowcloudapi/mongo"
	"time"

	"gopkg.in/mgo.v2/bson"
)

var STACKS_COLLECTION = "arrowcloud:stacks"

func GetStack(stackId string) (*models.Stack, error) {

	query := bson.M{
		"_id": bson.ObjectIdHex(stackId),
	}

	stackM, err := mongo.FindOneDocument(STACKS_COLLECTION, query)
	if err != nil {
		return nil, err
	}

	stack := models.Stack{
		ID:                     stackM["_id"].(bson.ObjectId).Hex(),
		Name:                   stackM["name"].(string),
		UserID:                 stackM["user_id"].(bson.ObjectId).Hex(),
		OrgID:                  stackM["org_id"].(bson.ObjectId).Hex(),
		CreationTime:           stackM["creation_time"].(time.Time),
		UpdateTime:             stackM["update_time"].(time.Time),
		OriginalComposeFile:    stackM["compose_file_original"].(string),
		TransformedComposeFile: stackM["compose_file_transformed"].(string),
		VolumeFolders:          stackM["volume_folders"].(string),
	}

	return &stack, nil
}

func RemoveStack(stackId string) (err error) {

	query := bson.M{
		"_id": bson.ObjectIdHex(stackId),
	}

	return mongo.RemoveDocument(STACKS_COLLECTION, query)
}

/*
 * A case confusing user in the current single-service implementation.
 * * A user belongs to multiple organizations.
 * * There is an app <app_name> in one of the organization <org1>.
 * * The user logins to another organization <org2> with appc cli.
 * * The user publishes <app_name> and expects the app will be created in <org2>.
 * * The result is that the app in <org1> gets published.
 *
 * Entitlement for stacks is similar to single-service apps.
 * * A stack belongs to one organziation.
 * * Stack name should be unique in an organziation.
 * * To avoid the above case we can ask user to provide organization Id when the user belongs
 * to more than one organization.
 *
 *
 */

func GetStacks(user models.User, orgID string, stackName string, userOnly bool) (*[]models.Stack, error) {

	dbQ := bson.M{}

	// If orgId is provided get stacks of the organization only.
	// If orgId is not provided get stacks of all organizations the user belongs to.
	// If userOnly is specified get stacks created by the user only.

	if orgID != "" {
		dbQ["org_id"] = orgID
	} else {
		userOrgIDs := []string{}
		for _, org := range user.Orgs {
			userOrgIDs = append(userOrgIDs, org.ID)
		}
		dbQ["org_id"] = bson.M{
			"$in": userOrgIDs,
		}
	}

	if userOnly {
		dbQ["user_id"] = bson.ObjectIdHex(user.ID)
	}

	if stackName != "" {
		dbQ["name"] = bson.RegEx{stackName, ""}
	}

	result, err := mongo.FindDocuments(STACKS_COLLECTION, dbQ)
	if err != nil {
		return nil, err
	}

	stacks := []models.Stack{}
	for _, stackM := range result {
		stack := models.Stack{
			ID:            stackM["_id"].(bson.ObjectId).Hex(),
			Name:          stackM["name"].(string),
			UserID:        stackM["user_id"].(bson.ObjectId).Hex(),
			OrgID:         stackM["org_id"].(string),
			CreationTime:  stackM["creation_time"].(time.Time),
			UpdateTime:    stackM["update_time"].(time.Time),
			VolumeFolders: stackM["volume_folders"].(string),
		}
		if stackM["compose_file_original"] != nil {
			stack.OriginalComposeFile = stackM["compose_file_original"].(string)
		}
		if stackM["compose_file_transformed"] != nil {
			stack.TransformedComposeFile = stackM["compose_file_transformed"].(string)
		}

		stacks = append(stacks, stack)
	}

	return &stacks, nil
}

func SaveStack(stack models.Stack) (string, error) {

	update := bson.M{
		"name":                     stack.Name,
		"user_id":                  bson.ObjectIdHex(stack.UserID),
		"org_id":                   stack.OrgID,
		"creation_time":            stack.CreationTime,
		"update_time":              stack.UpdateTime,
		"compose_file_original":    stack.OriginalComposeFile,
		"compose_file_transformed": stack.TransformedComposeFile,
		"volume_folders":           stack.VolumeFolders,
	}

	var query bson.M
	if stack.ID == "" {
		query = bson.M{
			"name":    stack.Name,
			"user_id": bson.ObjectIdHex(stack.UserID),
			"org_id":  stack.OrgID,
		}
	} else {
		query = bson.M{
			"_id": bson.ObjectIdHex(stack.ID),
		}
	}

	saved, err := mongo.UpsertDocument(STACKS_COLLECTION, query, update)
	if err != nil {
		return "", err
	}

	return saved["_id"].(bson.ObjectId).Hex(), err
}
