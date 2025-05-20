package profiles

import (
	"fmt"
	"fwtui/entity"
	"fwtui/modules/create_profile"
	"fwtui/modules/shared/confirmation"
	oscmd "fwtui/utils/cmd"
	"fwtui/utils/multiselect_list"
	"fwtui/utils/selectable_list"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/samber/lo"
)

// MODEl

const menuListProfiles = "INSTALLED_PROFILES"
const menuInstallProfile = "INSTALL_PROFILE"
const menuCreateProfile = "CREATE_PROFILE"

type ProfilesModule struct {
	view              viewState
	menu              *selectable_list.SelectableList[string]
	installedProfiles multiselect_list.MultiSelectableList[entity.UFWProfile]
	profilesToInstall multiselect_list.MultiSelectableList[entity.UFWProfile]

	deleteDialog        *confirmation.ConfirmDialog
	createProfileModule create_profile.ProfileForm
}

func Init() (ProfilesModule, tea.Cmd) {
	model := ProfilesModule{
		menu: selectable_list.NewSelectableList([]string{menuListProfiles, menuInstallProfile, menuCreateProfile}),
		view: viewStateHome,
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

	switch true {
	case m.view.isViewHome():
		switch msg := msg.(type) {
		case tea.KeyMsg:
			key := msg.String()
			switch key {
			case "up", "k":
				m.menu.Prev()
			case "down", "j":
				m.menu.Next()
			case "esc":
				return m, nil, ProfilesOutMsgEsc
			case "enter":
				switch m.menu.Selected() {
				case menuListProfiles:
					m.view = viewStateProfilesList
					m.installedProfiles.ClearSelection()
					m.installedProfiles.FocusFirst()
				case menuInstallProfile:
					m.view = viewStateInstallProfile
					m.profilesToInstall.ClearSelection()
					m.profilesToInstall.FocusFirst()
				case menuCreateProfile:
					m.view = viewStateCreateProfile
					m.createProfileModule = create_profile.NewProfileForm()
				}
			}
		}
	case m.view.isViewList():
		if m.deleteDialog != nil {
			newDeleteDialog, _, outMsg := m.deleteDialog.UpdateDialog(msg)
			m.deleteDialog = newDeleteDialog
			switch outMsg {
			case confirmation.ConfirmationDialogYes:
				m.deleteDialog = nil
				entity.DeleteProfile(m.installedProfiles.FocusedItem())
				m = m.reloadInstalledProfiles()
				m = m.reloadProfilesToInstall()
			case confirmation.ConfirmationDialogNo:
				m.deleteDialog = nil
			case confirmation.ConfirmationDialogEsc:
				m.deleteDialog = nil
			}

			return m, nil, ""
		}

		switch msg := msg.(type) {
		case tea.KeyMsg:
			key := msg.String()
			switch key {
			case "up", "k":
				m.installedProfiles.Prev()
			case "down", "j":
				m.installedProfiles.Next()
			case "delete", "d":
				m.deleteDialog = confirmation.NewConfirmDialog("Are you sure you want to delete this profile?")

			case "esc":
				m.view = viewStateHome
				m.menu.FocusFirst()
			case " ":
				m.installedProfiles.Toggle()
			case "enter":
				var output string
				var cmds []string
				if m.installedProfiles.NoneSelected() {
					profile := m.installedProfiles.FocusedItem()
					output = oscmd.RunCommand(fmt.Sprintf("sudo ufw allow \"%s\"", profile.Name))
				} else {
					lo.ForEach(m.installedProfiles.GetSelectedItems(), func(profile entity.UFWProfile, _ int) {
						command := fmt.Sprintf("sudo ufw allow \"%s\"", profile.Name)
						output += oscmd.RunCommand(command)
						cmds = append(cmds, command)

					})
				}
				return m, oscmd.OsCmdExecutedMsg(cmds, output), ""
			}
		}
	case m.view.isViewInstall():
		switch msg := msg.(type) {
		case tea.KeyMsg:
			key := msg.String()
			switch key {
			case "up", "k":
				m.profilesToInstall.Prev()
			case "down", "j":
				m.profilesToInstall.Next()
			case "esc":
				m.view = viewStateHome
			case " ":
				m.profilesToInstall.Toggle()
			case "enter":
				var output string

				if m.profilesToInstall.NoneSelected() {
					output = entity.CreateProfile(m.profilesToInstall.FocusedItem())
				} else {
					lo.ForEach(m.profilesToInstall.GetSelectedItems(), func(profile entity.UFWProfile, _ int) {
						output += "\n" + entity.CreateProfile(profile)
					})
				}
				m = m.reloadInstalledProfiles()
				m = m.reloadProfilesToInstall()
				return m, oscmd.OsCmdExecutedMsg([]string{}, output), ""
			}
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

	return m, nil, ""
}

func (m ProfilesModule) reloadInstalledProfiles() ProfilesModule {
	profiles, _ := entity.LoadInstalledProfiles()
	m.installedProfiles = multiselect_list.NewMultiSelectableList(profiles)
	return m
}

func (m ProfilesModule) reloadProfilesToInstall() ProfilesModule {
	m.profilesToInstall = multiselect_list.NewMultiSelectableList(entity.InstallableProfiles())
	return m
}

// VIEW

func (m ProfilesModule) ViewProfiles() string {
	var output string

	switch true {
	case m.view.isViewHome():
		lines := []string{"Select profile action:"}
		m.menu.ForEach(func(item string, _ int, isSelected bool) {
			prefix := lo.Ternary(isSelected, ">", " ")
			var itemName string

			switch item {
			case menuListProfiles:
				itemName = "List"
			case menuInstallProfile:
				itemName = "Install"
			case menuCreateProfile:
				itemName = "Create"
			}
			lines = append(lines, fmt.Sprintf("%s %s", prefix, itemName))
		})
		output = strings.Join(lines, "\n")
	case m.view.isViewList():
		if m.deleteDialog != nil {
			return m.deleteDialog.ViewDialog()
		}
		lines := []string{"Select profile:"}
		m.installedProfiles.ForEach(func(profile entity.UFWProfile, index int, isFocused, isSelected bool) {
			focusedPrefix := lo.Ternary(isFocused, ">", " ")
			selectedPrefix := lo.Ternary(isSelected, "*", " ")
			prefix := focusedPrefix + selectedPrefix
			lines = append(lines, fmt.Sprintf("%s %-20s | %-45s | %-45s", prefix, profile.Name, profile.Title, strings.Join(profile.Ports, ", ")))
		})

		output = strings.Join(lines, "\n")
		output += "\n Press Space to select"
		output += "\n Press Enter to allow"
	case m.view.isViewInstall():
		lines := []string{"Select profile to install:"}
		m.profilesToInstall.ForEach(func(profile entity.UFWProfile, index int, isFocused, isSelected bool) {
			focusedPrefix := lo.Ternary(isFocused, ">", " ")
			selectedPrefix := lo.Ternary(isSelected, "*", " ")
			prefix := focusedPrefix + selectedPrefix
			lines = append(lines, fmt.Sprintf("%s %-20s | %-45s | %-45s", prefix, profile.Name, profile.Title, strings.Join(profile.Ports, ", ")))
		})

		output = strings.Join(lines, "\n")
		output += "\n Press Space to select"
		output += "\n Press Enter to delete"
	case m.view.isViewCreate():
		output = m.createProfileModule.ViewCreateProfile()
	}

	return output
}
