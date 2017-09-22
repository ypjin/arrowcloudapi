package api

import (
	"net/http"
	swarmService "service/swarm"
	"utils/log"

	"github.com/docker/docker/api/types/swarm"
)

// ServiceAPI handles request to /api/services/{} /api/services/{}/logs
type ServiceAPI struct {
	BaseAPI
	userID      int
	serviceID   int64
	serviceName string
}

type serviceReq struct {
	ServiceName string `json:"service_name"`
	Public      int    `json:"public"`
}

const serviceNameMaxLen int = 30
const serviceNameMinLen int = 4
const dupServicePattern = `Duplicate entry '\w+' for key 'name'`

// Prepare validates the URL and the user
func (p *ServiceAPI) Prepare() {
	// idStr := p.Ctx.Input.Param(":id")
	// if len(idStr) > 0 {
	// 	var err error
	// 	p.serviceID, err = strconv.ParseInt(idStr, 10, 64)
	// 	if err != nil {
	// 		log.Errorf("Error parsing service id: %s, error: %v", idStr, err)
	// 		p.CustomAbort(http.StatusBadRequest, "invalid service id")
	// 	}

	// 	service, err := dao.GetServiceByID(p.serviceID)
	// 	if err != nil {
	// 		log.Errorf("failed to get service %d: %v", p.serviceID, err)
	// 		p.CustomAbort(http.StatusInternalServerError, "Internal error.")
	// 	}
	// 	if service == nil {
	// 		p.CustomAbort(http.StatusNotFound, fmt.Sprintf("service does not exist, id: %v", p.serviceID))
	// 	}
	// 	p.serviceName = service.Name
	// }
}

// Post ...
// func (p *ServiceAPI) Post() {
// 	p.userID = p.ValidateUser()

// 	var req serviceReq
// 	p.DecodeJSONReq(&req)
// 	public := req.Public
// 	err := validateServiceReq(req)
// 	if err != nil {
// 		log.Errorf("Invalid service request, error: %v", err)
// 		p.RenderError(http.StatusBadRequest, fmt.Sprintf("invalid request: %v", err))
// 		return
// 	}
// 	serviceName := req.ServiceName
// 	exist, err := dao.ServiceExists(serviceName)
// 	if err != nil {
// 		log.Errorf("Error happened checking service existence in db, error: %v, service name: %s", err, serviceName)
// 	}
// 	if exist {
// 		p.RenderError(http.StatusConflict, "")
// 		return
// 	}
// 	service := models.Service{OwnerID: p.userID, Name: serviceName, CreationTime: time.Now(), Public: public}
// 	serviceID, err := dao.AddService(service)
// 	if err != nil {
// 		log.Errorf("Failed to add service, error: %v", err)
// 		dup, _ := regexp.MatchString(dupServicePattern, err.Error())
// 		if dup {
// 			p.RenderError(http.StatusConflict, "")
// 		} else {
// 			p.RenderError(http.StatusInternalServerError, "Failed to add service")
// 		}
// 		return
// 	}
// 	p.Redirect(http.StatusCreated, strconv.FormatInt(serviceID, 10))
// }

// Head ...
// func (p *ServiceAPI) Head() {
// 	serviceName := p.GetString("service_name")
// 	if len(serviceName) == 0 {
// 		p.CustomAbort(http.StatusBadRequest, "service_name is needed")
// 	}

// 	service, err := dao.GetServiceByName(serviceName)
// 	if err != nil {
// 		log.Errorf("error occurred in GetServiceByName: %v", err)
// 		p.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
// 	}

// 	// only public service can be Headed by user without login
// 	if service != nil && service.Public == 1 {
// 		return
// 	}

// 	userID := p.ValidateUser()
// 	if service == nil {
// 		p.CustomAbort(http.StatusNotFound, http.StatusText(http.StatusNotFound))
// 	}

// 	if !checkServicePermission(userID, service.ServiceID) {
// 		p.CustomAbort(http.StatusForbidden, http.StatusText(http.StatusForbidden))
// 	}
// }

// Get ...
// func (p *ServiceAPI) Get() {
// 	service, err := dao.GetServiceByID(p.serviceID)
// 	if err != nil {
// 		log.Errorf("failed to get service %d: %v", p.serviceID, err)
// 		p.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
// 	}

// 	if service.Public == 0 {
// 		userID := p.ValidateUser()
// 		if !checkServicePermission(userID, p.serviceID) {
// 			p.CustomAbort(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
// 		}
// 	}

// 	p.Data["json"] = service
// 	p.ServeJSON()
// }

// Delete ...
// func (p *ServiceAPI) Delete() {
// 	if p.serviceID == 0 {
// 		p.CustomAbort(http.StatusBadRequest, "service ID is required")
// 	}

// 	userID := p.ValidateUser()

// 	if !hasServiceAdminRole(userID, p.serviceID) {
// 		p.CustomAbort(http.StatusForbidden, "")
// 	}

// 	contains, err := serviceContainsRepo(p.serviceName)
// 	if err != nil {
// 		log.Errorf("failed to check whether service %s contains any repository: %v", p.serviceName, err)
// 		p.CustomAbort(http.StatusInternalServerError, "")
// 	}
// 	if contains {
// 		p.CustomAbort(http.StatusPreconditionFailed, "service contains repositores, can not be deleted")
// 	}

// 	contains, err = serviceContainsPolicy(p.serviceID)
// 	if err != nil {
// 		log.Errorf("failed to check whether service %s contains any policy: %v", p.serviceName, err)
// 		p.CustomAbort(http.StatusInternalServerError, "")
// 	}
// 	if contains {
// 		p.CustomAbort(http.StatusPreconditionFailed, "service contains policies, can not be deleted")
// 	}

// 	if err = dao.DeleteService(p.serviceID); err != nil {
// 		log.Errorf("failed to delete service %d: %v", p.serviceID, err)
// 		p.CustomAbort(http.StatusInternalServerError, "")
// 	}

// 	go func() {
// 		if err := dao.AddAccessLog(models.AccessLog{
// 			UserID:    userID,
// 			ServiceID: p.serviceID,
// 			RepoName:  p.serviceName + "/",
// 			RepoTag:   "N/A",
// 			Operation: "delete",
// 		}); err != nil {
// 			log.Errorf("failed to add access log: %v", err)
// 		}
// 	}()
// }

// func serviceContainsRepo(name string) (bool, error) {
// 	repositories, err := getReposByService(name)
// 	if err != nil {
// 		return false, err
// 	}

// 	return len(repositories) > 0, nil
// }

// func serviceContainsPolicy(id int64) (bool, error) {
// 	policies, err := dao.GetRepPolicyByService(id)
// 	if err != nil {
// 		return false, err
// 	}

// 	return len(policies) > 0, nil
// }

// List ...
func (p *ServiceAPI) List() {
	var total int64
	var err error

	page, pageSize := p.GetPaginationParams()

	var serviceList []swarm.Service

	serviceList, err = swarmService.ListServices()
	if err != nil {
		log.Errorf("Error retrieving service info, error: %v", err)
		p.RenderError(http.StatusBadRequest, "failed to retrieve service info")
		return
	}

	total = int64(len(serviceList))

	p.SetPaginationHeader(total, page, pageSize)
	p.Data["json"] = serviceList
	p.ServeJSON()
}

// ToggleServicePublic ...
// func (p *ServiceAPI) ToggleServicePublic() {
// 	p.userID = p.ValidateUser()
// 	var req serviceReq

// 	serviceID, err := strconv.ParseInt(p.Ctx.Input.Param(":id"), 10, 64)
// 	if err != nil {
// 		log.Errorf("Error parsing service id: %d, error: %v", serviceID, err)
// 		p.RenderError(http.StatusBadRequest, "invalid service id")
// 		return
// 	}

// 	p.DecodeJSONReq(&req)
// 	public := req.Public
// 	if !isServiceAdmin(p.userID, serviceID) {
// 		log.Warningf("Current user, id: %d does not have service admin role for service, id: %d", p.userID, serviceID)
// 		p.RenderError(http.StatusForbidden, "")
// 		return
// 	}
// 	err = dao.ToggleServicePublicity(p.serviceID, public)
// 	if err != nil {
// 		log.Errorf("Error while updating service, service id: %d, error: %v", serviceID, err)
// 		p.RenderError(http.StatusInternalServerError, "Failed to update service")
// 	}
// }

// FilterAccessLog handles GET to /api/services/{}/logs
// func (p *ServiceAPI) FilterAccessLog() {
// 	p.userID = p.ValidateUser()

// 	var query models.AccessLog
// 	p.DecodeJSONReq(&query)

// 	if !checkServicePermission(p.userID, p.serviceID) {
// 		log.Warningf("Current user, user id: %d does not have permission to read accesslog of service, id: %d", p.userID, p.serviceID)
// 		p.RenderError(http.StatusForbidden, "")
// 		return
// 	}
// 	query.ServiceID = p.serviceID
// 	query.BeginTime = time.Unix(query.BeginTimestamp, 0)
// 	query.EndTime = time.Unix(query.EndTimestamp, 0)

// 	page, pageSize := p.GetPaginationParams()

// 	total, err := dao.GetTotalOfAccessLogs(query)
// 	if err != nil {
// 		log.Errorf("failed to get total of access log: %v", err)
// 		p.CustomAbort(http.StatusInternalServerError, "")
// 	}

// 	logs, err := dao.GetAccessLogs(query, pageSize, pageSize*(page-1))
// 	if err != nil {
// 		log.Errorf("failed to get access log: %v", err)
// 		p.CustomAbort(http.StatusInternalServerError, "")
// 	}

// 	p.SetPaginationHeader(total, page, pageSize)

// 	p.Data["json"] = logs

// 	p.ServeJSON()
// }

// func isServiceAdmin(userID int, pid int64) bool {
// 	isSysAdmin, err := dao.IsAdminRole(userID)
// 	if err != nil {
// 		log.Errorf("Error occurred in IsAdminRole, returning false, error: %v", err)
// 		return false
// 	}

// 	if isSysAdmin {
// 		return true
// 	}

// 	rolelist, err := dao.GetUserServiceRoles(userID, pid)
// 	if err != nil {
// 		log.Errorf("Error occurred in GetUserServiceRoles, returning false, error: %v", err)
// 		return false
// 	}

// 	hasServiceAdminRole := false
// 	for _, role := range rolelist {
// 		if role.RoleID == models.PROJECTADMIN {
// 			hasServiceAdminRole = true
// 			break
// 		}
// 	}

// 	return hasServiceAdminRole
// }

// func validateServiceReq(req serviceReq) error {
// 	pn := req.ServiceName
// 	if isIllegalLength(req.ServiceName, serviceNameMinLen, serviceNameMaxLen) {
// 		return fmt.Errorf("Service name is illegal in length. (greater than 4 or less than 30)")
// 	}
// 	validServiceName := regexp.MustCompile(`^[a-z0-9](?:-*[a-z0-9])*(?:[._][a-z0-9](?:-*[a-z0-9])*)*$`)
// 	legal := validServiceName.MatchString(pn)
// 	if !legal {
// 		return fmt.Errorf("Service name is not in lower case or contains illegal characters!")
// 	}
// 	return nil
// }
