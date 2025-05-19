package util

import (
	"strings"
)

func GetTemplateFileName(templateFile string, service string) string {
	templateSplit := strings.Split(templateFile, service+"/")
	templateFileName := strings.Split(templateSplit[len(templateSplit)-1], ".")[0]

	return templateFileName
}

// GetProjectService - returns project, service, and path to template on filesystem.
// driverConfig - driver configuration
// templateFile - full path to template file
// returns project, service, templatePath
func GetProjectService(deploymentProjectService string, templateBasis string, templateFile string) (string, string, int, string) {
	templateFile = strings.ReplaceAll(strings.ReplaceAll(templateFile, "\\\\", "/"), "\\", "/")
	splitDir := strings.Split(templateFile, "/")
	var project, service string
	offsetBase := 0

	trcTemplateParam := templateBasis

	for i, component := range splitDir {
		if component == trcTemplateParam {
			offsetBase = i
			break
		}
	}

	project = splitDir[offsetBase+1]
	var serviceIndex int
	if len(project) == 0 && len(deploymentProjectService) > 0 {
		projectServiceParts := strings.Split(deploymentProjectService, "/")
		project = projectServiceParts[0]
		service = projectServiceParts[1]
		serviceIndex = 0
	} else {
		serviceIndex = offsetBase + 2
		service = splitDir[serviceIndex]

		// Clean up service naming (Everything after '.' removed)
		if strings.Contains(templateFile, "Common") &&
			strings.Contains(service, ".mf.tmpl") {
			service = strings.Split(service, ".mf.tmpl")[0]
		}

		dotIndex := strings.Index(service, ".")
		if dotIndex > 0 && dotIndex <= len(service) {
			service = service[0:dotIndex]
		}
	}

	return project, service, serviceIndex, templateFile
}
