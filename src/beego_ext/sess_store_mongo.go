// The goal of this session store is that it can share sessions in database with Stratus.
// more docs: http://beego.me/docs/module/session.md
package beego_ext

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"mongo"
	"utils/log"

	"github.com/astaxie/beego/session"
	"gopkg.in/mgo.v2/bson"
)

var mongoProvider = &Provider{}
var SESSION_COLLECTION = "arrowcloud:sessions"

// MaxPoolSize redis max pool size
var MaxPoolSize = 100

// MongoSessionStore mongo session store
type MongoSessionStore struct {
	sid         string
	lock        sync.RWMutex
	values      map[interface{}]interface{}
	maxlifetime int64
}

// Set value in mongo session
func (rs *MongoSessionStore) Set(key, value interface{}) error {
	rs.lock.Lock()
	defer rs.lock.Unlock()
	rs.values[key] = value
	return nil
}

// Get value in mongo session
func (rs *MongoSessionStore) Get(key interface{}) interface{} {
	rs.lock.RLock()
	defer rs.lock.RUnlock()
	if v, ok := rs.values[key]; ok {
		return v
	}
	return nil
}

// Delete value in mongo session
func (rs *MongoSessionStore) Delete(key interface{}) error {
	rs.lock.Lock()
	defer rs.lock.Unlock()
	delete(rs.values, key)
	return nil
}

// Flush clear all values in mongo session
func (rs *MongoSessionStore) Flush() error {
	rs.lock.Lock()
	defer rs.lock.Unlock()
	rs.values = make(map[interface{}]interface{})
	return nil
}

// SessionID get mongo session id
func (rs *MongoSessionStore) SessionID() string {
	return rs.sid
}

// SessionRelease save session values to mongo
func (rs *MongoSessionStore) SessionRelease(w http.ResponseWriter) {

	log.Debugf("about to save session %v", rs.sid)

	// Stratus session saves userId in auth map
	// To make the session usable by both sides we need to keep it consistent
	if userId, ok := rs.values["userId"]; ok {
		auth := map[string]interface{}{
			"userId":   userId,
			"loggedIn": true,
		}
		rs.values["auth"] = auth
	}

	strValues, err := convertValues(rs.values)
	if err != nil {
		return
	}

	expires := time.Now().Add(time.Duration(rs.maxlifetime) * time.Second)
	//int64((time.Now().Unix() + rs.maxlifetime) * 1000) //milliseconds

	_, err = mongo.UpsertDocument(SESSION_COLLECTION,
		bson.M{
			"_id": rs.sid,
		}, bson.M{
			"_id":     rs.sid,
			"session": strValues,
			"expires": expires,
		})

	if err != nil {
		log.Errorf("Failed to save session. %v", err)
	}

	return
}

// Provider mongo session provider
type Provider struct {
	maxlifetime int64
	savePath    string
}

// SessionInit init mongo session
// savepath like mongo server addr, username, password, dbname, pool size, etc.
// mongodb://username:password@host1:port,host2:port,host3:port/dbname?replicaSet=rsname
// mongodb://username:password@host:port/dbname
func (rp *Provider) SessionInit(maxlifetime int64, savePath string) error {

	rp.maxlifetime = maxlifetime
	rp.savePath = savePath

	// It should initialize mongo session here.
	// Since we'll use the same mongo database as the system db which should be initialized
	// at system startup we'll skip it here.

	return nil
}

// SessionRead read mongo session by sid
func (rp *Provider) SessionRead(sid string) (session.Store, error) {

	log.Debugf("about to retrieve session %s", sid)
	/*
		stratus session:
		{
			"_id" : "593TVxv8WtuuqglkcsXLnCKq",
			"session" : "{\"cookie\":{\"originalMaxAge\":1209599988,\"expires\":\"2017-09-25T02:50:48.281Z\",\"httpOnly\":true,\"path\":\"/\"},\"mid\":\"be19c3bc52dc83ec5a11559028503582cd5cd07a\",\"lastCommand\":\"sudo\",\"username\":\"yjin@appcelerator.com\",\"auth\":{\"userId\":\"519b3ad4a00c829eff37b1fe\",\"loggedIn\":true},\"lastAccess\":1505098248289,\"sid_360\":\"s:bXCa2nBKjzncb-HLYfq_5xbmuW27xuXx.EeMvq7BgIcQFJzxzUuTRSuJSlV/BOSdHPDCFqHNbCsI\"}",
			"expires" : 1506307848000
		}

		session object:
		{
		  "cookie": {
		    "originalMaxAge": 1209599988,
		    "expires": "2017-09-25T02:50:48.281Z",
		    "httpOnly": true,
		    "path": "/"
		  },

		  "auth": {
		    "userId": "519b3ad4a00c829eff37b1fe",
		    "loggedIn": true
		  },

		  "mid": "be19c3bc52dc83ec5a11559028503582cd5cd07a",
		  "lastCommand": "sudo",
		  "username": "yjin@appcelerator.com",
		  "lastAccess": 1505098248289,
		  "sid_360": "s:bXCa2nBKjzncb-HLYfq_5xbmuW27xuXx.EeMvq7BgIcQFJzxzUuTRSuJSlV/BOSdHPDCFqHNbCsI"
		}
	*/

	kvs := make(map[interface{}]interface{})
	expires := time.Now().Add(time.Duration(rp.maxlifetime) * time.Second)
	//int64((time.Now().Unix() + rp.maxlifetime) * 1000) //for new session

	query := bson.M{
		"_id": sid,
		"expires": bson.M{
			"$gte": time.Now(),
		},
	}

	sess, err := mongo.FindOneDocument(SESSION_COLLECTION, query)
	if err != nil {
		log.Errorf("Failed to find session by sid %v. %v", sid, err)
		rs := &MongoSessionStore{sid: sid, values: kvs, maxlifetime: rp.maxlifetime}
		er := saveSession(sid, kvs, expires)
		return rs, er
	}

	if sess["_id"] == nil {
		log.Warningf("Session not found by sid %v", sid)
		rs := &MongoSessionStore{sid: sid, values: kvs, maxlifetime: rp.maxlifetime}
		er := saveSession(sid, kvs, expires)
		return rs, er
	}

	valuesMap := map[string]interface{}{}
	err = json.Unmarshal([]byte(sess["session"].(string)), &valuesMap)

	log.Debugf("session values map: %v", valuesMap)

	for k, v := range valuesMap {
		kvs[k] = v
	}

	// Stratus session saves userId in auth map
	// If we got a Stratus session copy session.auth.userId to session.userId
	if auth, ok := kvs["auth"]; ok {
		authMap := auth.(map[string]interface{})
		kvs["userId"] = authMap["userId"].(string)
	}

	rs := &MongoSessionStore{sid: sid, values: kvs, maxlifetime: rp.maxlifetime}

	return rs, nil
}

// SessionExist check mongo session exist by sid
func (rp *Provider) SessionExist(sid string) bool {

	sess, err := mongo.FindOneDocument(SESSION_COLLECTION, bson.M{"_id": sid})
	if err != nil {
		log.Errorf("Failed to find session by sid %v. %v", sid, err)
		return false
	}

	if sess["_id"] != nil {
		log.Debugf("Session %v exist!", sid)
		return true
	}

	log.Warningf("Session %v does not exist!", sid)
	return false
}

// SessionRegenerate generate new sid for mongo session
func (rp *Provider) SessionRegenerate(oldsid, sid string) (session.Store, error) {

	// not sure the use case for this, return nil for now

	return nil, nil
}

// SessionDestroy delete mongo session by id
func (rp *Provider) SessionDestroy(sid string) error {

	log.Warningf("about to destroy session %s", sid)
	return mongo.RemoveDocument(SESSION_COLLECTION, bson.M{"_id": sid})
}

// SessionGC Impelment method, no used.
func (rp *Provider) SessionGC() {

	query := bson.M{
		"expires": bson.M{
			"$lt": time.Now(), //int64(time.Now().Unix() * 1000),
		},
	}

	err := mongo.RemoveDocuments(SESSION_COLLECTION, query)
	if err != nil {
		log.Errorf("Failed to gc session. %v", err)
	}

	return
}

// SessionAll return all activeSession
func (rp *Provider) SessionAll() int {
	return 0
}

func saveSession(sid string, values map[interface{}]interface{}, expires time.Time) error {

	strValues, err := convertValues(values)
	if err != nil {
		return err
	}

	return mongo.InsertDocument(SESSION_COLLECTION, bson.M{"_id": sid, "session": strValues, "expires": expires})
}

func convertValues(values map[interface{}]interface{}) (result string, err error) {

	kvMap := map[string]interface{}{}
	for k, v := range values {
		kvMap[k.(string)] = v
	}

	strBytes, err := json.Marshal(kvMap)
	if err != nil {
		log.Errorf("Failed to marshal session values. %v", err)
		return
	}

	result = string(strBytes)

	return
}

func init() {
	session.Register("mongo", mongoProvider)
}
