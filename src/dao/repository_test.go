package dao

import (
	"testing"

	"models"
)

var (
	project    = "library"
	name       = "library/repository-test"
	repository = &models.RepoRecord{
		Name:        name,
		OwnerName:   "admin",
		ProjectName: project,
	}
)

func TestGetRepositoryByProjectName(t *testing.T) {
	if err := addRepository(repository); err != nil {
		t.Fatalf("failed to add repository %s: %v", name, err)
	}
	defer func() {
		if err := deleteRepository(name); err != nil {
			t.Fatalf("failed to delete repository %s: %v", name, err)
		}
	}()

	repositories, err := GetRepositoryByProjectName(project)
	if err != nil {
		t.Fatalf("failed to get repositories of project %s: %v",
			project, err)
	}

	if len(repositories) == 0 {
		t.Fatal("unexpected length of repositories: 0, at least 1")
	}

	exist := false
	for _, repo := range repositories {
		if repo.Name == name {
			exist = true
			break
		}
	}
	if !exist {
		t.Errorf("there is no repository whose name is %s", name)
	}
}

func TestGetTotalOfRepositories(t *testing.T) {
	total, err := GetTotalOfRepositories("")
	if err != nil {
		t.Fatalf("failed to get total of repositoreis: %v", err)
	}

	if err := addRepository(repository); err != nil {
		t.Fatalf("failed to add repository %s: %v", name, err)
	}
	defer func() {
		if err := deleteRepository(name); err != nil {
			t.Fatalf("failed to delete repository %s: %v", name, err)
		}
	}()

	n, err := GetTotalOfRepositories("")
	if err != nil {
		t.Fatalf("failed to get total of repositoreis: %v", err)
	}

	if n != total+1 {
		t.Errorf("unexpected total: %d != %d", n, total+1)
	}
}

func TestGetTotalOfPublicRepositories(t *testing.T) {
	total, err := GetTotalOfPublicRepositories("")
	if err != nil {
		t.Fatalf("failed to get total of public repositoreis: %v", err)
	}

	if err := addRepository(repository); err != nil {
		t.Fatalf("failed to add repository %s: %v", name, err)
	}
	defer func() {
		if err := deleteRepository(name); err != nil {
			t.Fatalf("failed to delete repository %s: %v", name, err)
		}
	}()

	n, err := GetTotalOfPublicRepositories("")
	if err != nil {
		t.Fatalf("failed to get total of public repositoreis: %v", err)
	}

	if n != total+1 {
		t.Errorf("unexpected total: %d != %d", n, total+1)
	}
}

func TestGetTotalOfUserRelevantRepositories(t *testing.T) {
	total, err := GetTotalOfUserRelevantRepositories(1, "")
	if err != nil {
		t.Fatalf("failed to get total of repositoreis for user %d: %v", 1, err)
	}

	if err := addRepository(repository); err != nil {
		t.Fatalf("failed to add repository %s: %v", name, err)
	}
	defer func() {
		if err := deleteRepository(name); err != nil {
			t.Fatalf("failed to delete repository %s: %v", name, err)
		}
	}()

	users, err := GetUserByProject(1, models.User{})
	if err != nil {
		t.Fatalf("failed to list members of project %d: %v", 1, err)
	}
	exist := false
	for _, user := range users {
		if user.UserID == 1 {
			exist = true
			break
		}
	}
	if !exist {
		if err = AddProjectMember(1, 1, models.DEVELOPER); err != nil {
			t.Fatalf("failed to add user %d to be member of project %d: %v", 1, 1, err)
		}
		defer func() {
			if err = DeleteProjectMember(1, 1); err != nil {
				t.Fatalf("failed to delete user %d from member of project %d: %v", 1, 1, err)
			}
		}()
	}

	n, err := GetTotalOfUserRelevantRepositories(1, "")
	if err != nil {
		t.Fatalf("failed to get total of public repositoreis for user %d: %v", 1, err)
	}

	if n != total+1 {
		t.Errorf("unexpected total: %d != %d", n, total+1)
	}
}

func addRepository(repository *models.RepoRecord) error {
	return AddRepository(*repository)
}

func deleteRepository(name string) error {
	return DeleteRepository(name)
}
