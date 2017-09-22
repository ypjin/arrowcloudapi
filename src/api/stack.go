package api

import (
	"dao"
	"fmt"
	"models"
	"net/http"
	"regexp"
	"time"
	"utils/log"
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
	userID    string
	stackID   string
	stackName string
}

type stackReq struct {
	StackName string `json:"stack_name"`
	Public    int    `json:"public"`
}

const stackNameMaxLen int = 30
const stackNameMinLen int = 4
const dupStackPattern = `Duplicate entry '\w+' for key 'name'`

func (p *StackAPI) Prepare() {

	p.userID = p.ValidateUser()

	nameStr := p.Ctx.Input.Param(":name") //stack name
	if len(nameStr) > 0 {
		p.stackName = nameStr

		stack, err := dao.GetStack(p.stackID)
		if err != nil {
			log.Errorf("failed to get stack %d: %v", p.stackName, err)
			p.CustomAbort(http.StatusInternalServerError, "Internal error.")
		}
		if stack == nil {
			p.CustomAbort(http.StatusNotFound, fmt.Sprintf("stack %s does not exist.", p.stackName))
		}
		p.stackID = stack.StackID
	}
}

// Deploy a stack with provided compose file
func (p *StackAPI) Deploy() {

	var req stackReq
	p.DecodeJSONReq(&req)
	err := validateStackReq(req)
	if err != nil {
		log.Errorf("Invalid stack request, error: %v", err)
		p.RenderError(http.StatusBadRequest, fmt.Sprintf("invalid request: %v", err))
		return
	}
	stackName := req.StackName
	exist, err := dao.StackExists(stackName)
	if err != nil {
		log.Errorf("Error happened checking stack existence in db, error: %v, stack name: %s", err, stackName)
	}
	if exist {
		p.RenderError(http.StatusConflict, "")
		return
	}
	stack := models.Stack{UserID: p.userID, Name: stackName, CreationTime: time.Now()}
	stackID, err := dao.SaveStack(stack)
	if err != nil {
		log.Errorf("Failed to add stack, error: %v", err)
		dup, _ := regexp.MatchString(dupStackPattern, err.Error())
		if dup {
			p.RenderError(http.StatusConflict, "")
		} else {
			p.RenderError(http.StatusInternalServerError, "Failed to add stack")
		}
		return
	}
	p.Redirect(http.StatusCreated, stackID)
}

// Delete a stack
func (p *StackAPI) Delete() {
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
}

func (p *StackAPI) Log() {

}

func (p *StackAPI) CheckServices() {

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
