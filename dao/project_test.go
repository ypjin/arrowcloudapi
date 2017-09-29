package dao

import (
	"fmt"
	"testing"

	"arrowcloudapi/models"
)

func TestDeleteProject(t *testing.T) {
	name := "project_for_test"
	project := models.Project{
		OwnerID: currentUser.UserID,
		Name:    name,
	}

	id, err := AddProject(project)
	if err != nil {
		t.Fatalf("failed to add project: %v", err)
	}
	defer func() {
		if err := delProjPermanent(id); err != nil {
			t.Errorf("failed to clear up project %d: %v", id, err)
		}
	}()

	if err = DeleteProject(id); err != nil {
		t.Fatalf("failed to delete project: %v", err)
	}

	p := &models.Project{}
	if err = GetOrmer().Raw(`select * from project where project_id = ?`, id).
		QueryRow(p); err != nil {
		t.Fatalf("failed to get project: %v", err)
	}

	if p.Deleted != 1 {
		t.Errorf("unexpeced deleted column: %d != %d", p.Deleted, 1)
	}

	deletedName := fmt.Sprintf("%s#%d", name, id)
	if p.Name != deletedName {
		t.Errorf("unexpected name: %s != %s", p.Name, deletedName)
	}

}

func delProjPermanent(id int64) error {
	_, err := GetOrmer().QueryTable("access_log").
		Filter("ProjectID", id).
		Delete()
	if err != nil {
		return err
	}

	_, err = GetOrmer().Raw(`delete from project_member 
		where project_id = ?`, id).Exec()
	if err != nil {
		return err
	}

	_, err = GetOrmer().QueryTable("project").
		Filter("ProjectID", id).
		Delete()
	return err
}
