package entitlement

import (
	"arrowcloudapi/models"
	"arrowcloudapi/utils/log"
)

/**
 * A user can update a stack when
 * * the stack is created by the user
 * * the stack belongs to an org which the is an admin of
 * @param user   models.User [description]
 * @param stack  models.Stack [description]
 * @return (hasPermission bool, err error) [description]
 */
func CanUpdate(user models.User, stack models.Stack) (hasPermission bool, err error) {

	log.Debugf("check permission for deploying stack. User: %s, stackName: %s, stackOrgId: %s", user.Email, stack.Name, stack.OrgID)

	var stackOrg *models.Org
	for _, org := range user.Orgs {
		if org.ID == stack.OrgID {
			stackOrg = &org
			break
		}
	}

	if stackOrg == nil {
		return //hasPermission defaults to false
	}

	if stackOrg.Node_acs_admin {
		hasPermission = true
	} else if stack.UserID == user.ID {
		hasPermission = true
	}

	return
}

/**
 * A user can view a stack when
 * * the stack is created by the user
 * * the stack belongs to one of the user's orgs
 * @param user   models.User [description]
 * @param stack  models.Stack [description]
 * @return (hasPermission bool, err error) [description]
 */
func CanView(user models.User, stack models.Stack) (hasPermission bool, err error) {
	log.Debugf("check permission for viewing stack. User: %s, stackName: %s, stackOrgId: %s", user.Email, stack.Name, stack.OrgID)

	for _, org := range user.Orgs {
		if org.ID == stack.OrgID {
			hasPermission = true
			break
		}
	}

	return
}

/**
 * A user can delete a stack when
 * * the stack is created by the user
 * * the stack belongs to an org which the is an admin of
 * @param user   models.User [description]
 * @param stack  models.Stack [description]
 * @return (hasPermission bool, err error) [description]
 */
func CanDelete(user models.User, stack models.Stack) (hasPermission bool, err error) {
	log.Debugf("check permission for deleting stack. User: %s, stackName: %s, stackOrgId: %s", user.Email, stack.Name, stack.OrgID)

	var stackOrg *models.Org
	for _, org := range user.Orgs {
		if org.ID == stack.OrgID {
			stackOrg = &org
			break
		}
	}

	if stackOrg == nil {
		return //hasPermission defaults to false
	}

	if stackOrg.Node_acs_admin {
		hasPermission = true
	} else if stack.UserID == user.ID {
		hasPermission = true
	}

	return
}
