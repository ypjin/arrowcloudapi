package dao

import (
	"arrowcloudapi/models"
)

func GetStack(name string) (*models.Stack, error) {

	return &models.Stack{}, nil

}

func SaveStack(stack models.Stack) (string, error) {
	return "", nil
}

func StackExists(stackName string) (bool, error) {

	return true, nil
}
