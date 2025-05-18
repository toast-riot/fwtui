package profiles

type viewState string

func (v viewState) isViewHome() bool {
	return v == viewStateHome
}

func (v viewState) isViewList() bool {
	return v == viewStateInstalledProfilesList
}

func (v viewState) isViewInstall() bool {
	return v == viewStateInstallProfile
}

const viewStateHome = "home"

const viewStateInstalledProfilesList = "installed_profiles_list"
const viewStateInstallProfile = "install_profile"
