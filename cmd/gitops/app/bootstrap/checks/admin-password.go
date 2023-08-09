package checks

const ADMIN_SECRET_NAME string = "cluster-user-auth"
const ADMIN_SECRET_NAMESPACE string = "flux-system"

func GetAdminPasswordSecrets() (string, string) {
	AdminUsernamePromptContent := promptContent{
		"Admin username can't be empty",
		"Please enter your admin username: ",
	}
	adminUsername := promptGetStringInput(AdminUsernamePromptContent)

	AdminPasswordPromptContent := promptContent{
		"Admin password can't be empty",
		"Please enter your admin password",
	}
	adminPassword := promptGetPasswordInput(AdminPasswordPromptContent)

	return adminUsername, adminPassword
}

func CreateAdminPasswordSecret() {
	adminUsername, adminPassword := GetAdminPasswordSecrets()
	data := map[string][]byte{
		"username": []byte(adminUsername),
		"password": []byte(adminPassword),
	}
	createSecret(ADMIN_SECRET_NAME, ADMIN_SECRET_NAMESPACE, data)
}
