package profiles

import (
	"fmt"
	"fwtui/domain/entity"
	"fwtui/domain/ufw"
	"fwtui/modules/profiles/createprofile"
	"fwtui/modules/shared/confirmation"
	"fwtui/utils/focusablelist"
	"fwtui/utils/multiselect"
	"fwtui/utils/teacmd"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/samber/lo"
)

// MODEl

const menuListProfiles = "INSTALLED_PROFILES"
const menuCreateFromList = "CREATE_PROFILE_FROM_LIST"
const menuCreateProfile = "CREATE_PROFILE"

type ProfilesModule struct {
	view              viewState
	menu              *focusablelist.SelectableList[string]
	installedProfiles multiselect.MultiSelectableList[entity.UFWProfile]
	profilesToInstall multiselect.MultiSelectableList[entity.UFWProfile]

	deleteDialog        *confirmation.ConfirmDialog
	createProfileModule createprofile.ProfileForm
}

func Init() (ProfilesModule, tea.Cmd) {
	model := ProfilesModule{
		menu: focusablelist.FromList([]string{menuListProfiles, menuCreateFromList, menuCreateProfile}),
		view: viewStateHome,
	}
	model = model.reloadInstalledProfiles()
	model = model.reloadProfilesToInstall()
	return model, nil
}

// UPDATE

type profilesDeletedMsg struct {
	Output string
}

type profilesAppliedMsg struct {
	Output string
}
type profilesCreatedMsg struct {
	Output string
}

type ProfilesEscMsg struct{}

func (mod ProfilesModule) UpdateProfilesModule(msg tea.Msg) (ProfilesModule, tea.Cmd) {
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
				return m, func() tea.Msg {
					return ProfilesEscMsg{}
				}
			case "enter":
				switch m.menu.Focused() {
				case menuListProfiles:
					m.view = viewStateProfilesList
					m.installedProfiles.ClearSelection()
					m.installedProfiles.FocusFirst()
				case menuCreateFromList:
					m.view = viewStateCreateProfileFromList
					m.profilesToInstall.ClearSelection()
					m.profilesToInstall.FocusFirst()
				case menuCreateProfile:
					m.view = viewStateCreateProfile
					m.createProfileModule = createprofile.NewProfileForm()
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
				return m, teacmd.RunOsCmdAndAfter(func() string {
					if m.installedProfiles.NoneSelected() {
						return entity.DeleteProfile(m.installedProfiles.FocusedItem())
					} else {
						var output string
						lo.ForEach(m.installedProfiles.GetSelectedItems(), func(profile entity.UFWProfile, _ int) {
							output += "\n" + entity.DeleteProfile(profile)
						})
						return output
					}

				}, func(output string) tea.Msg {
					return profilesDeletedMsg{Output: output}
				},
				)

			case confirmation.ConfirmationDialogNo:
				m.deleteDialog = nil
			case confirmation.ConfirmationDialogEsc:
				m.deleteDialog = nil
			}

			return m, nil
		}

		switch msg := msg.(type) {
		case profilesAppliedMsg:
			m.installedProfiles.ClearSelection()
			m.installedProfiles.FocusFirst()
			return m, teacmd.OsCmdExecutionFinishedCmd(msg.Output)

		case profilesDeletedMsg:
			m = m.reloadInstalledProfiles()
			m = m.reloadProfilesToInstall()
			m.installedProfiles.ClearSelection()
			return m, teacmd.OsCmdExecutionFinishedCmd(msg.Output)
		case tea.KeyMsg:
			key := msg.String()
			switch key {
			case "up", "k":
				m.installedProfiles.Prev()
			case "down", "j":
				m.installedProfiles.Next()
			case "delete", "d":
				if m.installedProfiles.NoneSelected() {
					m.deleteDialog = confirmation.NewConfirmDialog("Are you sure you want to delete this profile?")
				} else {
					m.deleteDialog = confirmation.NewConfirmDialog("Are you sure you want to delete selected profiles?")
				}
			case "esc":
				m.view = viewStateHome
				m.menu.FocusFirst()
			case " ":
				m.installedProfiles.Toggle()
			case "enter":
				return m, teacmd.RunOsCmdAndAfter(func() string {
					if m.installedProfiles.NoneSelected() {
						profile := m.installedProfiles.FocusedItem()
						return ufw.AllowProfile(profile.Name)
					} else {
						var output string
						lo.ForEach(m.installedProfiles.GetSelectedItems(), func(profile entity.UFWProfile, _ int) {
							output += "\n" + ufw.AllowProfile(profile.Name)

						})
						return output
					}
				}, func(s string) tea.Msg {
					return profilesAppliedMsg{Output: s}
				},
				)
			}
		}
	case m.view.isViewCreateFromList():
		switch msg := msg.(type) {
		case profilesCreatedMsg:
			m = m.reloadInstalledProfiles()
			m = m.reloadProfilesToInstall()
			return m, teacmd.OsCmdExecutionFinishedCmd(msg.Output)

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
				return m, teacmd.RunOsCmdAndAfter(func() string {
					if m.profilesToInstall.NoneSelected() {
						res := entity.CreateProfile(m.profilesToInstall.FocusedItem())
						if res.IsErr() {
							return res.Err().Error()
						}
						return res.Value()

					} else {
						var output string
						lo.ForEach(m.profilesToInstall.GetSelectedItems(), func(profile entity.UFWProfile, _ int) {
							res := entity.CreateProfile(profile)
							if res.IsErr() {
								output += "\n" + res.Err().Error()
							}
							output += "\n" + res.Value()
						})
						return output
					}
				}, func(s string) tea.Msg {
					return profilesCreatedMsg{Output: s}
				},
				)
			}
		}
	case m.view.isViewCreate():
		switch msg.(type) {
		case createprofile.CreateProfileCreatedMsg:
			m = m.reloadInstalledProfiles()
			m.view = viewStateHome
			return m, nil
		case createprofile.CreateProfileEscMsg:
			m.view = viewStateHome
			return m, nil
		}

		newForm, cmd := m.createProfileModule.UpdateProfileForm(msg)
		m.createProfileModule = newForm
		return m, cmd
	}

	return m, nil
}

func (m ProfilesModule) reloadInstalledProfiles() ProfilesModule {
	profiles, _ := entity.LoadInstalledProfiles()
	m.installedProfiles = multiselect.FromList(profiles)
	return m
}

func (m ProfilesModule) reloadProfilesToInstall() ProfilesModule {
	m.profilesToInstall = multiselect.FromList(entity.InstallableProfiles())
	return m
}

// VIEW

func (m ProfilesModule) ViewProfiles() string {
	var output string

	switch true {
	case m.view.isViewHome():
		lines := []string{"Focus profile action:"}
		m.menu.ForEach(func(item string, _ int, isSelected bool) {
			prefix := lo.Ternary(isSelected, ">", " ")
			var itemName string

			switch item {
			case menuListProfiles:
				itemName = "List"
			case menuCreateFromList:
				itemName = "Create (from list)"
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
		lines := []string{"Focus profile:"}
		m.installedProfiles.ForEach(func(profile entity.UFWProfile, index int, isFocused, isSelected bool) {
			focusedPrefix := lo.Ternary(isFocused, ">", " ")
			selectedPrefix := lo.Ternary(isSelected, "*", " ")
			prefix := focusedPrefix + selectedPrefix
			lines = append(lines, fmt.Sprintf("%s %-20s | %-45s | %-45s", prefix, profile.Name, profile.Title, strings.Join(profile.Ports, ", ")))
		})

		output = strings.Join(lines, "\n")
		output += "\n\n↑↓ to navigate, d to delete, Space to select, Enter to enable profile, Esc to cancel"
	case m.view.isViewCreateFromList():
		lines := []string{"Focus profile to install:"}
		m.profilesToInstall.ForEach(func(profile entity.UFWProfile, index int, isFocused, isSelected bool) {
			focusedPrefix := lo.Ternary(isFocused, ">", " ")
			selectedPrefix := lo.Ternary(isSelected, "*", " ")
			prefix := focusedPrefix + selectedPrefix
			lines = append(lines, fmt.Sprintf("%s %-20s | %-45s | %-45s", prefix, profile.Name, profile.Title, strings.Join(profile.Ports, ", ")))
		})

		output = strings.Join(lines, "\n")
		output += "\n\n↑↓ to navigate, Space to select, Enter to create profile, Esc to cancel"
	case m.view.isViewCreate():
		output = m.createProfileModule.ViewCreateProfile()
	}

	return output
}
