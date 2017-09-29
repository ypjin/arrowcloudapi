package dashboard

import (
	"arrowcloudapi/models"
	"arrowcloudapi/mongo"
	"arrowcloudapi/utils/log"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/jeffail/gabs"
)

const (
	host360         = "platform.appcelerator.com"
	authPath        = "/api/v1/auth/login"
	logoutPath      = "/api/v1/auth/logout"
	orgInfoPath     = "/api/v1/user/organizations"
	thisEnvAdminURL = "http://admin.cloudapp-1.appctest.com"
)

// Auth implements Authenticator interface to authenticate user against DB.
type Auth struct{}

/**
 * Authenticate user against appcelerator 360 (dashboard). This is for enterprise user only.
 * @param username
 * @param password
 * @param cb
 */
//function validateThrough360(mid, username, password, callback) {
func (d *Auth) Authenticate(m models.AuthModel) (*models.User, error) {

	//check whether the dashboard session is still valid first
	/*
	   >db.dashboard_sessions.findOne()
	   {
	   "_id" : ObjectId("53d07fcba38d8ba60518c900"),
	   "username" : "rdong@appcelerator.com",
	   "sid_360": "s%3ANpiTvlGoViClfe_peVLfBJFN.r7IEVTSaVKnz2a6nQ8joUn2Uf8o1QMKv40YRnnime3E",
	   "cookie": [
	   "connect.sid=s%3ANpiTvlGoViClfe_peVLfBJFN.r7IEVTSaVKnz2a6nQ8joUn2Uf8o1QMKv40YRnnime3E; Domain=360-preprod.appcelerator.com; Path=/; HttpOnly; Secure"
	   ]
	   }
	*/

	//TODO find and invalidate previous 360 session

	loginUrl := "https://" + host360 + authPath

	username := m.Principal
	creds := url.Values{}
	creds.Set("username", username)
	creds.Add("password", m.Password)
	// v.Encode() == "name=Ava&friend=Jess&friend=Sarah&friend=Zoe"

	//curl -i -b cookies.txt -c cookies.txt -F "username=mgoff@appcelerator.com" -F "password=food" http://360-dev.appcelerator.com/api/v1/auth/login
	/*
	   response for bad username/password
	   HTTP/1.1 400 Bad Request
	   X-Powered-By: Express
	   Access-Control-Allow-Origin: *
	   Access-Control-Allow-Methods: GET, POST, DELETE, PUT
	   Access-Control-Allow-Headers: Content-Type, api_key
	   Content-Type: application/json; charset=utf-8
	   Content-Length: 79
	   Date: Fri, 19 Apr 2013 01:25:24 GMT
	   Connection: keep-alive

	   {"success":false,"description":"Invalid password.","code":400,"internalCode":2}
	*/
	resp, err := http.PostForm(loginUrl, creds)

	if err != nil {
		log.Errorf("Failed to login to dashboard. %v", err)
		return nil, err
	}

	//log.Debugf("resp: %v", resp)

	if resp.StatusCode != 200 {
		log.Debugf("dashboard returns status %s", resp.Status)
		return nil, errors.New("authentication failed")
	}

	bodyBuf, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Errorf("Failed to read response body. %v", err)
		return nil, err
	}

	jsonBody, err := gabs.ParseJSON(bodyBuf)
	if err != nil {
		log.Errorf("Failed to parse response body. %v", err)
		return nil, err
	}

	// 'set-cookie': ['t=UpnUzNztGWO7K8A%2BCYihZz056Bk%3D; Path=/; Expires=Sat, 16 Nov 2013 06:27:19 GMT',
	// 	'un=mgoff%40appcelerator.com; Path=/; Expires=Sat, 16 Nov 2013 06:27:19 GMT',
	// 	'sid=33f33a6b7f8fef7b0fc649654187d467; Path=/; Expires=Sat, 16 Nov 2013 06:27:19 GMT',
	// 	'dvid=2019bea3-9e7b-48e3-890f-00e3e22b39e2; Path=/; Expires=Sat, 17 Oct 2015 06:27:19 GMT',
	// 	'connect.sid=s%3Aj0kX71OMFpIQ11Vf1ruhqJLH.on4RLy9q9tpVqnUeoQJBWlDPiB6bS8rWWhq8sOCDGPc; Domain=360-dev.appcelerator.com; Path=/; Expires=Sat, 16 Nov 2013 06:27:19 GMT; HttpOnly'
	// ]
	// {
	// 	"success": true,
	// 	"result": {
	// 		"success": true,
	// 		"username": "mgoff@appcelerator.com",
	// 		"email": "mgoff@appcelerator.com",
	// 		"guid": "ae6150453b3599b2875b311c40785b40",
	// 		"org_id": 14301,
	// 		"connect.sid": "s:QGW1cqj5h9B3fL6jwJTtjkuT.iuwQ23WOgiK/E+QfkRNVWi7G5S9DA00Li6BQPLGkROM"
	// }

	cookie := resp.Header.Get("set-cookie")
	if cookie == "" {
		log.Error("No cookie found in response")
		return nil, errors.New("authentication failed")
	}

	sid := strings.Split(strings.Split(cookie, ";")[0], "=")[1]
	log.Debugf("sid: %s", sid)

	// for _, cookie := range cookies.([]string) {
	// 	log.Debugf("cookie: %s", cookie)
	// }

	success := jsonBody.Path("success").Data().(bool)
	if !success {
		log.Error("dashboard returns false for success field")
		return nil, errors.New("authentication failed")
	}

	err = handleDashboardSession(username, sid, cookie)
	if err != nil {
		return nil, err
	}

	haveAccess, orgs, err := getAndVerifyOrgInfoFrom360(username, sid)
	if err != nil {
		return nil, err
	}
	if !haveAccess {
		log.Errorf("user's organizations do not have access to this domain")
		return nil, errors.New("No access to this domain")
	}

	user := bson.M{
		"username": jsonBody.Path("result.username").Data().(string),
		"email":    jsonBody.Path("result.email").Data().(string),
		"guid":     jsonBody.Path("result.guid").Data().(string),
	}
	if jsonBody.Path("result.firstname").Data() != nil {
		user["firstname"] = jsonBody.Path("result.firstname").Data().(string)
	} else {
		user["firstname"] = jsonBody.Path("result.username").Data().(string)
	}

	//user's organization info returned from dashboard. It's an array since a user can belong to
	//multiple organizations.
	user["orgs_360"] = orgs
	user["orgs_360_updated_at"] = time.Now()

	savedUser, err := saveUser(user)
	if err != nil {
		log.Errorf("Failed to save user. %v", err)
		return nil, err
	}

	mUser := &models.User{
		UserID:   savedUser["_id"].(bson.ObjectId).Hex(),
		Username: jsonBody.Path("result.username").Data().(string),
		Email:    jsonBody.Path("result.email").Data().(string),
		Orgs:     orgs,
	}
	if jsonBody.Path("result.firstname").Data() != nil {
		mUser.Firstname = jsonBody.Path("result.firstname").Data().(string)
	} else {
		mUser.Firstname = jsonBody.Path("result.username").Data().(string)
	}

	return mUser, nil
}

func handleDashboardSession(username, sid_360, cookie string) error {

	old_db_session, err := findDashboardSession(username)
	if err != nil && !strings.Contains(err.Error(), "not found") {
		return err
	}
	if _, ok := old_db_session["sid_360"]; ok {
		err = logoutFromDashboard(old_db_session["sid_360"].(string))
		if err != nil {
			return err
		}
	}

	// save 360 session to DB
	db_session := bson.M{
		"username": username,
		"sid_360":  sid_360,
		"cookie":   cookie,
	}
	return saveDashboardSession(db_session)
}

func getAndVerifyOrgInfoFrom360(username, sid string) (haveAccess bool, orgs []models.Org, err error) {

	// reqTimeout := 20000; //20s

	//curl -i -b connect.sid=s%3AaJaL7IWQ_cDvmVBeQRY997hf.vVzLV2aFvrYiEKmfdTARTuHessesQ0Xm87JvFESaus http://dashboard.appcelerator.com/api/v1/user/organizations
	/*
	   response for invalid session
	   HTTP/1.1 401 Unauthorized
	   X-Frame-Options: SAMEORIGIN
	   Cache-Control: no-cache, max-age=0, must-revalidate
	   Pragma: no-cache
	   Vary: Accept-Encoding
	   Access-Control-Allow-Origin: *
	   Access-Control-Allow-Methods: GET, POST, PUT, PATCH, DELETE
	   Access-Control-Allow-Headers: Content-Type, api_key
	   Content-Type: application/json; charset=utf-8
	   Content-Length: 59
	   Set-Cookie: connect.sid=s%3AIEpzWmzs4MQJGJMEcLmjlZm_.Cyi4LlO8gP%2B4sPHR0bdEGqjiqjuW3RJlZe6O2bt8QkI; Domain=dashboard.appcelerator.com; Path=/; Expires=Sat, 12 Apr 2014 13:04:07 GMT; HttpOnly; Secure
	   Date: Thu, 13 Mar 2014 13:04:07 GMT
	   Connection: close

	   {"success":false,"description":"Login Required","code":401}
	*/

	log.Debug("Get user organization information from " + host360 + orgInfoPath)

	orgInfoUrl := "https://" + host360 + orgInfoPath

	//https://webcache.googleusercontent.com/search?q=cache:OVK76hrG4T8J:https://medium.com/%40nate510/don-t-use-go-s-default-http-client-4804cb19f779+&cd=4&hl=en&ct=clnk&gl=jp
	client := &http.Client{}

	req, err := http.NewRequest("GET", orgInfoUrl, nil)

	req.Header.Add("Cookie", "connect.sid="+sid)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)

	if err != nil {
		log.Errorf("Failed to get organization info from dashboard. %v", err)
		return
	}

	//log.Debugf("resp: %v", resp)

	if resp.StatusCode == 401 {
		log.Warning("getAndVerifyOrgInfoFrom360 - Failed to get organization information. Session is invalid")
		err = errors.New("Failed to get organization information. Session is invalid.")
		return
	}

	if resp.StatusCode != 200 {
		log.Debugf("dashboard returns status %s", resp.Status)
		err = errors.New("Failed to get organization info")
		return
	}

	bodyBuf, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Errorf("Failed to read response body. %v", err)
		return
	}

	jsonBody, err := gabs.ParseJSON(bodyBuf)
	if err != nil {
		log.Errorf("Failed to parse response body. %v", err)
		return
	}

	success := jsonBody.Path("success").Data().(bool)
	if !success {
		log.Error("dashboard returns false for success field")
		err = errors.New("Failed to get organization info")
		return
	}

	/*
		     {
			     "success": true,
			     "result": [{
				     "_id": "51c40b4497a98c6046000002",
				     "org_id": 14301,
				     "name": "Appcelerator, Inc",
				     "guid": "64310644-794b-c8d0-a8b8-0a373d20dabc",
				     "user_count": 97,
				     "current_users_role": "normal",
				     "is_node_acs_admin": false,
				     "trial_end_date": "",
				     "created": "2012-01-11 10:58:09.0",
				     "reseller": false,
				     "active": true,
				     "envs": [{
				     "_id": "production",
				     "name": "production",
				     "isProduction": true,
				     "acsBaseUrl": "https://preprod-api.cloud.appcelerator.com",
				     "acsAuthBaseUrl": "https://dolphin-secure-identity.cloud.appcelerator.com",
				     "nodeACSEndpoint": "https://admin.cloudapp-enterprise-preprod.appcelerator.com"
			     }, {
				     "_id": "development",
				     "name": "development",
				     "isProduction": false,
				     "acsBaseUrl": "https://preprod-api.cloud.appcelerator.com",
				     "acsAuthBaseUrl": "https://dolphin-secure-identity.cloud.appcelerator.com",
				     "nodeACSEndpoint": "https://admin.cloudapp-enterprise-preprod.appcelerator.com"
			     }],
			     "parent_org_guid": ""
			     }]
		     }
	*/

	organizations := jsonBody.Path("result").Data().([]interface{})
	if !validateOrgs(organizations) {
		log.Errorf("getAndVerifyOrgInfoFrom360 - Bad response from dashboard: invalid organization info. %v", organizations)
		err = errors.New("Bad response from dashboard")
		return
		//TODO send mail
	}

	//check if the user's organizations have access to current deployment (identified by admin host)
	orgs, haveAccess = checkOrgs(organizations)
	return

}

func checkOrgs(orgArray []interface{}) (orgs []models.Org, haveAccess bool) {

	log.Debugf("check if user's organizations have access to this domain")
	re := regexp.MustCompile("^(http|https)://") //https://golang.org/pkg/regexp/#MustCompile
	thisEnvHost := re.ReplaceAllString(thisEnvAdminURL, "")

	orgs = []models.Org{} //organizations which can access this domain (deployment)
	userOrgIds := []string{}

	for _, orgData := range orgArray {

		orgDoc := orgData.(map[string]interface{})
		orgToSave := models.Org{
			Id:             strconv.FormatFloat(orgDoc["org_id"].(float64), 'f', -1, 64),
			Name:           orgDoc["name"].(string),
			Admin:          orgDoc["current_users_role"].(string) == "admin",
			Node_acs_admin: orgDoc["is_node_acs_admin"].(bool),
		}
		userOrgIds = append(userOrgIds, orgToSave.Id)

		//check if the org has access to this domain (deployment)
		//if yes save it in "orgs"
		if envsData, ok := orgDoc["envs"]; ok {
			envs := envsData.([]interface{})
			for _, envData := range envs {
				env := envData.(map[string]interface{})
				adminHost, hok := env["nodeACSEndpoint"].(string)
				if hok {
					re := regexp.MustCompile("^(http|https)://")
					adminHost := re.ReplaceAllString(adminHost, "")
					log.Debugf("org %s(%s) have access to %s", orgToSave.Name, orgToSave.Id, adminHost)
					if adminHost == thisEnvHost {
						orgs = append(orgs, orgToSave)
						break
					}
				}
			}
		}
	}

	//workaround for testing - start
	// userOrgIds.push('14301');
	// orgs.push({id:'14301', name:'appcelerator Inc.', admin: true, node_acs_admin: true});
	//workaround for testing - end

	if len(orgs) < 1 {
		log.Errorf("getAndVerifyOrgInfoFrom360 - User's organization(s) %v doesn't have access to current deployment (%s).", userOrgIds, thisEnvHost)
		haveAccess = false
		return
	}

	haveAccess = true
	return
}

/**
 * Validate the organization info got from 360 for a user is valid.
 * @param orgArray
 * @returns {boolean}
 */
func validateOrgs(orgArray []interface{}) bool {

	if len(orgArray) == 0 {
		return false
	}

	for _, orgData := range orgArray {
		orgDoc := orgData.(map[string]interface{})

		if _, ok := orgDoc["org_id"]; !ok {
			return false
		}
		if _, ok := orgDoc["name"]; !ok {
			return false
		}
		if _, ok := orgDoc["is_node_acs_admin"]; !ok {
			return false
		}
	}
	return true
}

/**
 * Load user's 360 session information based on username.
 */
func findDashboardSession(username string) (bson.M, error) {

	log.Debugf("find dashboard session for user %s", username)
	re, err := mongo.FindOneDocument(mongo.STRATUS_DASHBOARD_SESSIONS_COLL,
		bson.M{"username": username})

	if err != nil {
		log.Errorf("Failed to find dashboard session. %v", err)
		return nil, err
	}

	return re, err
}

/**
 * insert or update user's 360 session information upon login to 360
 */
func saveDashboardSession(session bson.M) error {

	log.Debugf("save dashboard session for user %v", session["username"])
	_, err := mongo.UpsertDocument(mongo.STRATUS_DASHBOARD_SESSIONS_COLL,
		bson.M{"username": session["username"]}, session)

	if err != nil {
		log.Errorf("Failed to save dashboard session. %v", err)
		return err
	}

	log.Debugf("Upserted %v into %s collection.", session, mongo.STRATUS_DASHBOARD_SESSIONS_COLL)
	return nil
}

/**
 * insert or update user information upon login to Appcelerator's sso interface
 */
func saveUser(user bson.M) (bson.M, error) {

	saved, err := mongo.UpsertDocument(mongo.STRATUS_USERS_COLL,
		bson.M{"guid": user["guid"]}, user)

	if err != nil {
		log.Errorf("Failed to save user. %v", err)
		return nil, err
	}

	log.Debugf("Upserted %v into %s collection.", user, mongo.STRATUS_USERS_COLL)
	return saved, nil
}

//TOOD use request module to support proxy
func logoutFromDashboard(sid_360 string) (err error) {

	log.Debugf("Logout session %s from Appcelerator 360.", sid_360)

	logOutUrl := "https://" + host360 + logoutPath

	client := &http.Client{}

	req, err := http.NewRequest("GET", logOutUrl, nil)

	req.Header.Add("Cookie", "connect.sid="+sid_360)

	resp, err := client.Do(req)

	if err != nil {
		log.Errorf("Failed to logout from dashboard. %v", err)
		return
	}

	//log.Debugf("resp: %v", resp)

	if resp.StatusCode == 400 {
		log.Warning("Failed to logout from dashboard. Session is invalid")
		err = errors.New("Failed to logout from dashboard. Session is invalid.")
		return
	}

	if resp.StatusCode != 200 {
		log.Debugf("dashboard returns status %s", resp.Status)
		err = errors.New("Failed to logout from dashboard")
		return
	}

	bodyBuf, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Errorf("Failed to read response body. %v", err)
		return
	}

	jsonBody, err := gabs.ParseJSON(bodyBuf)
	if err != nil {
		log.Errorf("Failed to parse response body. %v", err)
		return
	}

	success := jsonBody.Path("success").Data().(bool)
	if !success {
		log.Error("dashboard returns false for success field")
		err = errors.New("Failed to logout from dashboard")
		return
	}

	return
}
