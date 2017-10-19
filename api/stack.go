package api

import (
	"arrowcloudapi/dao"
	"arrowcloudapi/models"
	"arrowcloudapi/service/entitlement"
	"arrowcloudapi/service/swarm"
	"arrowcloudapi/utils/log"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

// StackAPI handles requests to the following APIs
/* https://wiki.appcelerator.org/display/cls/Service+Stack+Support
PUT /stack/<stack-name>		acs stack deploy
GET /stack/services 		acs stack services
GET /stacks?<query> 		acs stack ls
DELETE /stack/<stack-name>	acs stack rm
GET /stack/log?<query>		acs stack log
*/
type StackAPI struct {
	BaseAPI
	userID string
	user   *models.User
	// stackID   string
	// stackName string
}

type stackReq struct {
	StackName string `json:"stack_name"`
	Public    int    `json:"public"`
}

var tmpFileDir string

const stackNameMaxLen int = 30
const stackNameMinLen int = 4
const dupStackPattern = `Duplicate entry '\w+' for key 'name'`

func (p *StackAPI) Prepare() {

	p.userID, p.user = p.ValidateUser()

	// nameStr := p.Ctx.Input.Param(":name") //stack name
	// if len(nameStr) > 0 {
	// 	p.stackName = nameStr

	// 	stack, err := dao.GetStack(p.stackID)
	// 	if err != nil {
	// 		log.Errorf("failed to get stack %d: %v", p.stackName, err)
	// 		p.CustomAbort(http.StatusInternalServerError, "Internal error.")
	// 	}
	// 	if stack == nil {
	// 		p.CustomAbort(http.StatusNotFound, fmt.Sprintf("stack %s does not exist.", p.stackName))
	// 	}
	// 	p.stackID = stack.StackID
	// }
}

// Deploy a stack with provided compose file

// http://sweetohm.net/article/go-yaml-parsers.en.html
// https://stackoverflow.com/questions/32147325/how-to-parse-yaml-with-dyanmic-key-in-golang
// http://ghodss.com/2014/the-right-way-to-handle-yaml-in-golang/

// https://stackoverflow.com/questions/32310838/upload-file-with-same-format-using-beego
// https://stackoverflow.com/questions/26750457/multiple-file-upload-with-beego
func (p *StackAPI) Deploy() {

	stackName := p.Ctx.Input.Param(":name") //stack name
	orgId := p.GetString("orgid")

	log.Debugf("stack name: %s, org id: %s", stackName, orgId)

	// user must provide org id for the stack if he belongs to multiple orgs
	if len(p.user.Orgs) > 1 && orgId == "" {
		p.CustomAbort(http.StatusBadRequest, "Organization ID is required since you belong to more than one organization.")
	}

	// permission check
	stack := models.Stack{}

	if orgId != "" {
		stack.OrgID = orgId
		hasViewPermission, err := entitlement.CanView(*p.user, stack)
		if err != nil {
			p.CustomAbort(http.StatusInternalServerError, "internal error")
		}
		if !hasViewPermission {
			p.CustomAbort(http.StatusUnauthorized, "no permission")
		}
	} else {
		stack.OrgID = p.user.Orgs[0].ID
	}

	stacks, err := dao.GetStacks(*p.user, stack.OrgID, stackName, false)
	if err != nil {
		log.Errorf("failed to query stacks. %v", err)
		p.CustomAbort(http.StatusInternalServerError, "internal error")
	}

	if len(*stacks) == 0 {
		stack.UserID = p.user.ID
		stack.Name = stackName
		stack.CreationTime = time.Now()
		stack.UpdateTime = time.Now()
		stackID, err := dao.SaveStack(stack)
		if err != nil {
			log.Errorf("failed to save stack. %v", err)
			p.CustomAbort(http.StatusInternalServerError, "internal error")
		}
		stack.ID = stackID
	} else {
		stack = (*stacks)[0]
	}

	hasUpdatePermission, err := entitlement.CanUpdate(*p.user, stack)
	if !hasUpdatePermission {
		p.CustomAbort(http.StatusUnauthorized, "no permission")
	}

	// do stack deployment
	myField := p.GetString("my_field")
	log.Debugf("myField: %s", myField)
	//myBuffer := p.Get

	file, header, err := p.GetFile(stackName) // where <<this>> is the controller and <<file>> the id of your form field
	if err != nil {
		log.Errorf("GetFile error: %v", err)
		p.CustomAbort(http.StatusInternalServerError, "internal error")
	}

	composeFile := ""
	if file != nil {
		// get the filename
		fileName := header.Filename

		nowTag := time.Now().Format("20060102T150405Z")
		composeFile = filepath.Join(tmpFileDir, stackName+"_"+nowTag+".yaml")

		log.Debugf("uploaded file name: %s, composeFile: %v", fileName, composeFile)

		// save to server
		err := p.SaveToFile(stackName, composeFile)
		if err != nil {
			log.Errorf("SaveToFile error: %v", err)
			p.CustomAbort(http.StatusInternalServerError, "internal error")
		}
	}

	// var rr interface{}

	// err = json.Unmarshal([]byte(`{"success": true}`), &rr)
	// if err != nil {
	// 	log.Errorf("Unmarshal error: %v", err)
	// 	p.CustomAbort(http.StatusInternalServerError, "internal error")
	// }

	// p.Data["json"] = rr

	output, err := swarm.DeployStack(stack, composeFile)
	if err != nil {
		log.Errorf("Deploy error: %v", err)
		p.CustomAbort(http.StatusInternalServerError, err.Error())
	}

	result := map[string]interface{}{
		"success": true,
		"data":    output,
	}

	p.Data["json"] = result

	p.ServeJSON()

	// var req stackReq
	// p.DecodeJSONReq(&req)
	// err := validateStackReq(req)
	// if err != nil {
	// 	log.Errorf("Invalid stack request, error: %v", err)
	// 	p.RenderError(http.StatusBadRequest, fmt.Sprintf("invalid request: %v", err))
	// 	return
	// }
	// stackName := req.StackName
	// exist, err := dao.StackExists(stackName)
	// if err != nil {
	// 	log.Errorf("Error happened checking stack existence in db, error: %v, stack name: %s", err, stackName)
	// }
	// if exist {
	// 	p.RenderError(http.StatusConflict, "")
	// 	return
	// }
	// stack := models.Stack{UserID: p.userID, Name: stackName, CreationTime: time.Now()}
	// stackID, err := dao.SaveStack(stack)
	// if err != nil {
	// 	log.Errorf("Failed to add stack, error: %v", err)
	// 	dup, _ := regexp.MatchString(dupStackPattern, err.Error())
	// 	if dup {
	// 		p.RenderError(http.StatusConflict, "")
	// 	} else {
	// 		p.RenderError(http.StatusInternalServerError, "Failed to add stack")
	// 	}
	// 	return
	// }
	// p.Redirect(http.StatusCreated, stackID)
}

// Delete a stack
func (p *StackAPI) Delete() {

	stackName := p.Ctx.Input.Param(":name") //stack name
	orgId := p.GetString("orgid")
	log.Debugf("Stack name is %s, orgid is %s", stackName, orgId)

	if len(p.user.Orgs) > 1 && orgId == "" {
		p.CustomAbort(http.StatusBadRequest, "Organization ID is required since you belong to more than one organization.")
	}

	if orgId != "" {
		hasPermission, err := entitlement.CanView(*p.user, models.Stack{OrgID: orgId})
		if err != nil {
			log.Errorf("check permission error: %v", err)
			p.CustomAbort(http.StatusInternalServerError, "internal error")
		}
		if !hasPermission {
			p.CustomAbort(http.StatusUnauthorized, "no permission")
		}
	}

	// 1. get stacks from db
	stacks, err := dao.GetStacks(*p.user, orgId, stackName, false)
	if err != nil {
		p.CustomAbort(http.StatusInternalServerError, "internal error")
	}

	if len(*stacks) == 0 {
		log.Debug("no matched stack found from db.")
		p.CustomAbort(http.StatusOK, "no stack found")
	} else {
		log.Debugf("number of matched stack found from db: %v", len(*stacks))
	}

	stack := (*stacks)[0]
	hasUpdatePermission, err := entitlement.CanUpdate(*p.user, stack)
	if !hasUpdatePermission {
		p.CustomAbort(http.StatusUnauthorized, "no permission")
	}

	// 2. remove the stack from swarm
	output, err := swarm.RemoveStack(stackName)
	if err != nil {
		log.Errorf("swarm.RemoveStack error: %v", err)
		p.CustomAbort(http.StatusInternalServerError, "internal error")
	}

	log.Debugf("data for response: %s", output)

	// 3. remove the stack from db
	err = dao.RemoveStack(stack.ID)
	if err != nil {
		log.Errorf("dao.RemoveStack error: %v", err)
		p.CustomAbort(http.StatusInternalServerError, "internal error")
	}

	result := map[string]interface{}{
		"success": true,
		"data":    output,
	}

	p.Data["json"] = result

	p.ServeJSON()

	// if p.stackID == 0 {
	// 	p.CustomAbort(http.StatusBadRequest, "stack ID is required")
	// }

	// userID := p.ValidateUser()

	// if !hasStackAdminRole(userID, p.stackID) {
	// 	p.CustomAbort(http.StatusForbidden, "")
	// }

	// contains, err := stackContainsRepo(p.stackName)
	// if err != nil {
	// 	log.Errorf("failed to check whether stack %s contains any repository: %v", p.stackName, err)
	// 	p.CustomAbort(http.StatusInternalServerError, "")
	// }
	// if contains {
	// 	p.CustomAbort(http.StatusPreconditionFailed, "stack contains repositores, can not be deleted")
	// }

	// contains, err = stackContainsPolicy(p.stackID)
	// if err != nil {
	// 	log.Errorf("failed to check whether stack %s contains any policy: %v", p.stackName, err)
	// 	p.CustomAbort(http.StatusInternalServerError, "")
	// }
	// if contains {
	// 	p.CustomAbort(http.StatusPreconditionFailed, "stack contains policies, can not be deleted")
	// }

	// if err = dao.DeleteStack(p.stackID); err != nil {
	// 	log.Errorf("failed to delete stack %d: %v", p.stackID, err)
	// 	p.CustomAbort(http.StatusInternalServerError, "")
	// }

	// go func() {
	// 	if err := dao.AddAccessLog(models.AccessLog{
	// 		UserID:    userID,
	// 		StackID:   p.stackID,
	// 		RepoName:  p.stackName + "/",
	// 		RepoTag:   "N/A",
	// 		Operation: "delete",
	// 	}); err != nil {
	// 		log.Errorf("failed to add access log: %v", err)
	// 	}
	// }()
}

// List stacks available for the logged in user
// Get result by calling "docker stack ls" directly
func (p *StackAPI) List() {
	// var total int64
	// var err error

	// page, pageSize := p.GetPaginationParams()

	// var stackList []swarm.Service

	// stackList, err = swarmService.ListServices()
	// if err != nil {
	// 	log.Errorf("Error retrieving stack info, error: %v", err)
	// 	p.RenderError(http.StatusBadRequest, "failed to retrieve stack info")
	// 	return
	// }

	// total = int64(len(stackList))

	// p.SetPaginationHeader(total, page, pageSize)
	// p.Data["json"] = stackList
	// p.ServeJSON()

	stackName := p.GetString("stackname")
	orgId := p.GetString("orgid")
	userOnly, err := p.GetBool("useronly")
	if err != nil {
		p.CustomAbort(http.StatusBadRequest, "useronly is invalid")
	}

	if orgId != "" {
		hasPermission, err := entitlement.CanView(*p.user, models.Stack{OrgID: orgId})
		if err != nil {
			log.Errorf("check permission error: %v", err)
			p.CustomAbort(http.StatusInternalServerError, "internal error")
		}
		if !hasPermission {
			p.CustomAbort(http.StatusUnauthorized, "no permission")
		}
	}

	// 1. get stacks from db
	stacks, err := dao.GetStacks(*p.user, orgId, stackName, userOnly)
	if err != nil {
		p.CustomAbort(http.StatusInternalServerError, "internal error")
	}

	if len(*stacks) == 0 {
		log.Debug("no matched stack found from db.")
		p.CustomAbort(http.StatusOK, "no stack found")
	} else {
		log.Debugf("number of matched stack found from db: %v", len(*stacks))
	}

	stackIds := []string{}
	for _, stack := range *stacks {
		stackIds = append(stackIds, stack.ID)
	}

	// 2. get stack detail from swarm
	stacksFromSwarm, err := swarm.ListStacksFromAPI(stackIds)
	if err != nil {
		log.Errorf("ListStacks error: %v", err)
		p.CustomAbort(http.StatusInternalServerError, "internal error")
	}

	log.Debugf("data for response: %v", stacksFromSwarm)

	result := map[string]interface{}{
		"success": true,
		"data":    stacksFromSwarm,
	}

	p.Data["json"] = result

	p.ServeJSON()

}

// List stacks by calling docker daemon API
func (p *StackAPI) ListByCommand() {

	output, err := swarm.ListStacks()
	if err != nil {
		log.Errorf("ListStacks error: %v", err)
		p.CustomAbort(http.StatusInternalServerError, "internal error")
	}

	log.Debugf("data for response: %s", output)

	result := map[string]interface{}{
		"success": true,
		"data":    output,
	}

	p.Data["json"] = result

	p.ServeJSON()
}

func (p *StackAPI) GetServiceLog() {

	stackName := p.GetString("stackname")     //p.Ctx.Input.Param(":stackname")     //stack name
	serviceName := p.GetString("servicename") //p.Ctx.Input.Param(":servicename") //service name

	log.Debugf("Stack name is %s", stackName)
	log.Debugf("Service name is %s", serviceName)

	output, err := swarm.GetServiceLog(stackName, serviceName)
	if err != nil {
		log.Errorf("getServiceLog error: %v", err)
		p.CustomAbort(http.StatusInternalServerError, "internal error")
	}

	log.Debugf("data for response: %s", output)

	result := map[string]interface{}{
		"success": true,
		"data":    output,
	}

	p.Data["json"] = result

	p.ServeJSON()

}

func (p *StackAPI) CheckServices() {

	stackName := p.GetString("stackname") //p.Ctx.Input.Param(":stackname") //stack name
	orgId := p.GetString("orgid")
	log.Debugf("Stack name is %s, orgid is %s", stackName, orgId)

	if len(p.user.Orgs) > 1 && orgId == "" {
		p.CustomAbort(http.StatusBadRequest, "Organization ID is required since you belong to more than one organization.")
	}

	if orgId != "" {
		hasPermission, err := entitlement.CanView(*p.user, models.Stack{OrgID: orgId})
		if err != nil {
			log.Errorf("check permission error: %v", err)
			p.CustomAbort(http.StatusInternalServerError, "internal error")
		}
		if !hasPermission {
			p.CustomAbort(http.StatusUnauthorized, "no permission")
		}
	}

	// 1. get stacks from db
	stacks, err := dao.GetStacks(*p.user, orgId, stackName, false)
	if err != nil {
		p.CustomAbort(http.StatusInternalServerError, "internal error")
	}

	if len(*stacks) == 0 {
		log.Debug("no matched stack found from db.")
		p.CustomAbort(http.StatusOK, "no stack found")
	} else {
		log.Debugf("number of matched stack found from db: %v", len(*stacks))
	}

	stack := (*stacks)[0]

	output, err := swarm.CheckServices(stack.Name)
	if err != nil {
		log.Errorf("CheckServices error: %v", err)
		p.CustomAbort(http.StatusInternalServerError, "internal error")
	}

	log.Debugf("data for response: %s", output)

	result := map[string]interface{}{
		"success": true,
		"data":    output,
	}

	p.Data["json"] = result

	p.ServeJSON()

}

func validateStackReq(req stackReq) error {
	pn := req.StackName
	if isIllegalLength(req.StackName, stackNameMinLen, stackNameMaxLen) {
		return fmt.Errorf("Stack name is illegal in length. (greater than 4 or less than 30)")
	}
	validStackName := regexp.MustCompile(`^[a-z0-9](?:-*[a-z0-9])*(?:[._][a-z0-9](?:-*[a-z0-9])*)*$`)
	legal := validStackName.MatchString(pn)
	if !legal {
		return fmt.Errorf("stack name is not in lower case or contains illegal characters")
	}
	return nil
}

func isIllegalLength(s string, min int, max int) bool {
	if min == -1 {
		return (len(s) > max)
	}
	if max == -1 {
		return (len(s) <= min)
	}
	return (len(s) < min || len(s) > max)
}

func init() {

	tmpFileDir = filepath.Join(os.TempDir(), "composefiles")

	if _, err := os.Stat(tmpFileDir); err != nil {
		err := os.MkdirAll(tmpFileDir, 0777)
		if err != nil {
			panic(err)
		}
	}
}
