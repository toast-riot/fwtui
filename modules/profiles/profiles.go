package profiles

import (
	"fmt"
	"fwtui/entity"
	"fwtui/modules/create_profile"
	oscmd "fwtui/utils/cmd"
	"fwtui/utils/set"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/samber/lo"
)

// MODEl

var profileHomeActions = []string{menuInstalledProfiles, menuInstallProfile, menuCreateProfile}

const menuInstalledProfiles = "INSTALLED_PROFILES"
const menuInstallProfile = "INSTALL_PROFILE"
const menuCreateProfile = "CREATE_PROFILE"

type ProfilesModule struct {
	view              viewState
	cursor            int
	installedProfiles []entity.UFWProfile
	profilesToInstall []entity.UFWProfile
	selectedItems     set.Set[int]

	createProfileModule create_profile.ProfileForm
}

func Init() (ProfilesModule, tea.Cmd) {
	model := ProfilesModule{
		view:          viewStateHome,
		selectedItems: set.NewSet[int](),
	}
	model = model.reloadInstalledProfiles()
	model = model.reloadProfilesToInstall()
	return model, nil
}

// UPDATE

type ProfilesOutMsg string

const ProfilesOutMsgEsc ProfilesOutMsg = "ProfilesRuleEsc"

func (mod ProfilesModule) UpdateProfilesModule(msg tea.Msg) (ProfilesModule, tea.Cmd, ProfilesOutMsg) {
	m := mod
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()

		switch true {
		case m.view.isViewHome():
			switch key {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(profileHomeActions)-1 {
					m.cursor++
				}
			case "esc":
				return m, nil, ProfilesOutMsgEsc
			case "enter":
				switch profileHomeActions[m.cursor] {
				case menuInstalledProfiles:
					m.view = viewStateInstalledProfilesList
					m.cursor = 0
				case menuInstallProfile:
					m.view = viewStateInstallProfile
					m.cursor = 0
				case menuCreateProfile:
					m.view = viewStateCreateProfile
					m.createProfileModule = create_profile.NewProfileForm()
					m.cursor = 0
				}
			}
		case m.view.isViewList():
			switch key {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.installedProfiles)-1 {
					m.cursor++
				}
			case "esc":
				m.view = viewStateHome
				m.cursor = 0
				m.selectedItems = set.NewSet[int]()
			case " ":
				m.selectedItems = m.selectedItems.Toggle(m.cursor)
			case "enter":
				var output string
				var cmds []string
				if m.selectedItems.IsEmpty() {
					profile := m.installedProfiles[m.cursor]
					output = oscmd.RunCommand(fmt.Sprintf("sudo ufw allow \"%s\"", profile.Name))()
				} else {
					lo.ForEach(m.selectedItems.ToSlice(), func(i int, _ int) {
						profile := m.installedProfiles[i]
						command := fmt.Sprintf("sudo ufw allow \"%s\"", profile.Name)
						output += oscmd.RunCommand(command)()
						cmds = append(cmds, command)

					})
					m.cursor = 0
					m.selectedItems = set.NewSet[int]()
				}
				return m, oscmd.OsCmdExecutedMsg(cmds, output), ""
			}

		case m.view.isViewInstall():
			switch key {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.profilesToInstall)-1 {
					m.cursor++
				}
			case "esc":
				m.view = viewStateHome
				m.cursor = 0
				m.selectedItems = set.NewSet[int]()
			case " ":
				m.selectedItems = m.selectedItems.Toggle(m.cursor)
			case "enter":
				var output string

				if m.selectedItems.IsEmpty() {
					profile := m.profilesToInstall[m.cursor]
					output = entity.CreateProfile(profile)
				} else {
					lo.ForEach(m.selectedItems.ToSlice(), func(i int, _ int) {
						profile := m.profilesToInstall[i]
						output += "\n" + entity.CreateProfile(profile)
					})
					m.cursor = 0
					m.selectedItems = set.NewSet[int]()
				}

				m = m.reloadInstalledProfiles()
				m = m.reloadProfilesToInstall()
				return m, oscmd.OsCmdExecutedMsg([]string{}, output), ""
			}
		case m.view.isViewCreate():
			newForm, cmd, outMsg := m.createProfileModule.UpdateProfileForm(msg)
			m.createProfileModule = newForm
			switch outMsg {
			case create_profile.CreateProfileCreated:
				m = m.reloadInstalledProfiles()
				m.view = viewStateHome
			case create_profile.CreateProfileEsc:
				m.view = viewStateHome
			}
			return m, cmd, ""
		}
	}
	return m, nil, ""
}

func (m ProfilesModule) reloadInstalledProfiles() ProfilesModule {
	profiles, _ := entity.LoadInstalledProfiles()
	m.installedProfiles = profiles
	return m
}

func (m ProfilesModule) reloadProfilesToInstall() ProfilesModule {
	m.profilesToInstall = entity.InstallableProfiles()
	return m
}

// VIEW

func (m ProfilesModule) ViewProfiles() string {
	var output string

	switch true {
	case m.view.isViewHome():
		lines := []string{"Select profile action:"}
		for i, item := range profileHomeActions {
			prefix := " "
			if i == m.cursor {
				prefix = ">"
			}
			var itemName string
			switch item {
			case menuInstalledProfiles:
				itemName = "List installed"
			case menuInstallProfile:
				itemName = "Install"
			case menuCreateProfile:
				itemName = "Create"
			}
			lines = append(lines, fmt.Sprintf("%s %s", prefix, itemName))
		}
		output = strings.Join(lines, "\n")
	case m.view.isViewList():
		lines := []string{"Select profile:"}
		for i, profile := range m.installedProfiles {
			prefix := "  "
			if i == m.cursor && m.selectedItems.Has(i) {
				prefix = ">*"
			} else if i == m.cursor {
				prefix = "> "
			} else if m.selectedItems.Has(i) {
				prefix = " *"
			}
			lines = append(lines, fmt.Sprintf("%s %-20s | %-45s | %-45s", prefix, profile.Name, profile.Title, strings.Join(profile.Ports, ", ")))
		}
		output = strings.Join(lines, "\n")
		output += "\n Press Space to select"
		output += "\n Press Enter to allow"
	case m.view.isViewInstall():
		lines := []string{"Select profile to install:"}
		for i, profile := range m.profilesToInstall {
			prefix := "  "
			if i == m.cursor && m.selectedItems.Has(i) {
				prefix = ">*"
			} else if i == m.cursor {
				prefix = "> "
			} else if m.selectedItems.Has(i) {
				prefix = " *"
			}
			lines = append(lines, fmt.Sprintf("%s %-20s | %-45s | %-45s", prefix, profile.Name, profile.Title, strings.Join(profile.Ports, ", ")))
		}
		output = strings.Join(lines, "\n")
		output += "\n Press Space to select"
		output += "\n Press Enter to delete"
	case m.view.isViewCreate():
		output = m.createProfileModule.ViewCreateProfile()
	}

	return output
}
