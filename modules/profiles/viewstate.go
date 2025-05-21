package profiles

type viewState string

func (v viewState) isViewHome() bool {
	return v == viewStateHome
}

func (v viewState) isViewList() bool {
	return v == viewStateProfilesList
}

func (v viewState) isViewCreateFromList() bool {
	return v == viewStateCreateProfileFromList
}

func (v viewState) isViewCreate() bool {
	return v == viewStateCreateProfile
}

const viewStateHome = "home"

const viewStateProfilesList = "profiles_list"
const viewStateCreateProfileFromList = "create_profile_from_list"
const viewStateCreateProfile = "create_profile"
