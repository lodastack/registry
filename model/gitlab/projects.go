package gitlab

import (
	"errors"
	"fmt"

	"github.com/lodastack/registry/config"
)

var (
	sep        = "%2f"
	projectUrl = "%s/api/v3/projects/%s"

	InvalidProject = errors.New("invalid project")
)

type ProjectInfo struct {
	ID   int    `json:"id"`
	Name string `json:"name"`

	SshUrlToRepo string `json:"ssh_url_to_repo"`
	PathWithNs   string `json:"path_with_namespace"`
}

func getProcectID(domain, pName string) (int, error) {
	var u gitUrl = gitUrl(fmt.Sprintf(projectUrl, domain, config.C.PluginConf.Group+"%2f"+pName))
	var projectInfo ProjectInfo
	if err := u.ToJSON(&projectInfo); err != nil {
		return 0, err
	}
	if projectInfo.ID == 0 {
		return 0, InvalidProject
	}
	return projectInfo.ID, nil
}
