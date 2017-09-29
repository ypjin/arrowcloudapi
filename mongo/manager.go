package mongo

import (
	"arrowcloudapi/utils/log"
	"strconv"
	"strings"

	"gopkg.in/mgo.v2"
)

var session *mgo.Session

//mongodb://username:password@host1:port,host2:port,host3:port/dbname?replicaSet=rsname
//mongodb://username:password@host:port/dbname
func Initialize(options map[string]string) (err error) {

	hostname := options["hostname"]
	port := options["port"]
	rsname := options["rsname"]
	dbname := options["dbname"]
	numConn := options["poolsize"]
	username := options["username"]
	password := options["password"]

	url := "mongodb://"
	if username != "" {
		url += (username + ":" + password + "@")
	}

	var url_servers = ""
	if strings.Contains(hostname, ",") {
		// This is a mongodb replica
		hosts := strings.Split(hostname, ",")
		for i, host := range hosts {
			if i > 0 {
				url_servers += ","
			}
			url_servers += (host + ":" + port)
		}
		url += url_servers
		url += ("/" + dbname)
		if rsname != "" {
			url += ("?replicaSet=" + rsname)
		}

	} else {
		// This is normal single mongodb
		url_servers = hostname + ":" + port
		url += url_servers
		url += ("/" + dbname)
	}

	log.Debugf("Mongo URL: %s", url)

	session, err = mgo.Dial(url)

	if err != nil {
		log.Errorf("Failed to create mongo session. %v", err)
		return err
	}
	// defer session.Close()

	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)

	if numConn == "" {
		numConn = "5"
	}
	intNumConn, err := strconv.ParseInt(numConn, 10, 0)
	if err != nil {
		log.Warningf("Failed to parse numConn %s as int. Will use 5. %v", numConn, err)
		intNumConn = 5
	}
	session.SetPoolLimit(int(intNumConn))

	return nil
}

func GetSession() *mgo.Session {
	return session.Copy()
}
