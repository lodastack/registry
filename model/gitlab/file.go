package gitlab

import (
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/lodastack/registry/config"
)

var (
	fileUrl = "%s/api/v3/projects/%d/repository/files?file_path=%s&ref=%s"

	InvalidFile = errors.New("invalid file")
)

type FileDetail struct {
	FileName string `json:"file_name"`
	FilePath string `json:"file_path"`
	Size     int    `json:"size"`
	Encoding string `json:"encoding"`
	Content  string `json:"content"`
}

func getFileContent(domain string, pid int, filePath, ref string) (string, error) {
	var u gitUrl = gitUrl(fmt.Sprintf(fileUrl, domain, pid, filePath, ref))

	var fileDetail FileDetail
	if err := u.ToJSON(&fileDetail); err != nil {
		return "", err
	}
	if len(fileDetail.Content) == 0 {
		return "", InvalidFile
	}

	content, err := base64.StdEncoding.DecodeString(fileDetail.Content)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func GetFileContent(projectName string) (string, error) {
	pID, err := getProcectID(config.C.PluginConf.GitlabDomain, projectName)
	if err != nil {
		return "", err
	}
	return getFileContent(config.C.PluginConf.GitlabDomain, pID, config.C.PluginConf.AlarmFile, config.C.PluginConf.Branch)
}
