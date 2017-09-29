package swarm

import (
	"arrowcloudapi/utils/log"
	"errors"
	"os"
	"os/exec"
)

func execServiceCommand(args ...string) (output string, err error) {

	serviceCmdArgs := append([]string{"service"}, args...)

	return execDockerCommand(serviceCmdArgs...)
}

func execStackCommand(args ...string) (output string, err error) {

	stackCmdArgs := append([]string{"stack"}, args...)

	return execDockerCommand(stackCmdArgs...)
}

func execDockerCommand(args ...string) (output string, err error) {

	dockerCmdName := "docker"

	log.Debugf("execute command: %s %v", dockerCmdName, args)

	cmd := exec.Command(dockerCmdName, args...)

	envVars, err := buildEnv()
	if err != nil {
		return
	}
	cmd.Env = append(os.Environ(), envVars...)

	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		log.Error(string(stdoutStderr))
		log.Error(err)
		return
	}

	output = string(stdoutStderr)
	log.Debugf("stdoutStderr: %s\n", stdoutStderr)

	return
}

func buildEnv() ([]string, error) {

	kvs := []string{}

	hostSpec := os.Getenv("DOCKER_HOST")
	if hostSpec == "" {
		return nil, errors.New("DOCKER_HOST env var not found!")
	}

	certPath := os.Getenv("DOCKER_CERT_PATH")
	if certPath == "" {
		log.Warning("DOCKER_CERT_PATH is not set!")
	}

	kvs = append(kvs, "DOCKER_HOST="+hostSpec)
	if certPath != "" {
		kvs = append(kvs, "DOCKER_CERT_PATH="+certPath)
		kvs = append(kvs, "DOCKER_TLS_VERIFY=1")
	}

	return kvs, nil
}
