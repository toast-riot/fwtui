package profiles

type viewState string

func (v viewState) isViewHome() bool {
	return v == viewStateHome
}

func (v viewState) isViewList() bool {
	return v == viewStateProfilesList
}

func (v viewState) isViewInstall() bool {
	return v == viewStateInstallProfile
}

func (v viewState) isViewCreate() bool {
	return v == viewStateCreateProfile
}

const viewStateHome = "home"

const viewStateProfilesList = "installed_profiles_list"
const viewStateInstallProfile = "install_profile"
const viewStateCreateProfile = "create_profile"
