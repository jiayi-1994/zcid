package deployment

import "strings"

// canDeployToEnv enforces FR49: deploy role by environment.
// - member: only "dev" or "development"
// - project_admin or admin: any environment including staging/prod
// - project_token: any environment when upstream scope middleware has already allowed deployments:write
func canDeployToEnv(role, envName string) bool {
	role = strings.TrimSpace(strings.ToLower(role))
	envName = strings.TrimSpace(strings.ToLower(envName))

	if role == "admin" || role == "project_admin" || role == "project_token" {
		return true
	}
	if role == "member" {
		return envName == "dev" || envName == "development"
	}
	return false
}
