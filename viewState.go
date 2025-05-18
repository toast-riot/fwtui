package main

type viewState string

func (v viewState) isCreateRule() bool {
	return v == createRule
}

func (v viewState) isHome() bool {
	return v == home
}

func (v viewState) isProfilesHome() bool {
	return v == profilesHome
}

func (v viewState) isInstalledProfilesList() bool {
	return v == installedProfilesList
}

func (v viewState) isInstallProfile() bool {
	return v == installProfile
}

func (v viewState) isDeleteRule() bool {
	return v == deleteRule
}

const home = "home"

const profilesHome = "profiles_home"

const installedProfilesList = "installed_profiles_list"
const installProfile = "install_profile"

const createRule = "create_rule"
const deleteRule = "delete_rule"
