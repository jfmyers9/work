package tui

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jfmyers9/work/internal/editor"
	"github.com/jfmyers9/work/internal/model"
	"github.com/jfmyers9/work/internal/tracker"
)

type screen int

const (
	screenList screen = iota
	screenDetail
	screenStatus
	screenComment
	screenConfirm
	screenCreate
	screenLink
	screenHistory
)

type editorDoneMsg struct {
	issueID string
	err     error
}

type issueLinkMsg struct {
	childID string
	err     error
}

type issueUnlinkMsg struct {
	childID string
	err     error
}

type linkModel struct {
	childID string
	input   textinput.Model
}

func newLinkModel(childID string) linkModel {
	ti := textinput.New()
	ti.Placeholder = "parent issue ID"
	ti.CharLimit = 64
	ti.Width = 40
	ti.Focus()
	return linkModel{childID: childID, input: ti}
}

type rootModel struct {
	tracker      *tracker.Tracker
	screen       screen
	prevScreen   screen
	list         listModel
	detail       detailModel
	statusPicker statusPicker
	commentInput commentModel
	confirm      confirmModel
	createForm   createModel
	linkInput    linkModel
	history      historyModel
	help         helpModel
	issues       []model.Issue
	user         string
	editor       string
	statusMsg    string
	width        int
	height       int
}

func newModel(t *tracker.Tracker, issues []model.Issue, user, editorCmd string) rootModel {
	return rootModel{
		tracker: t,
		screen:  screenList,
		list:    newListModel(issues, 80),
		issues:  issues,
		user:    user,
		editor:  editorCmd,
		width:   80,
		height:  24,
	}
}

func (m rootModel) Init() tea.Cmd {
	return nil
}

func (m rootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case statusChangedMsg:
		m.statusMsg = fmt.Sprintf("Status → %s", msg.status)
		m.reloadIssues()
		m.screen = m.prevScreen
		if m.screen == screenDetail {
			issue, err := m.tracker.LoadIssue(msg.issueID)
			if err == nil {
				children := tracker.FilterIssues(m.issues, tracker.FilterOptions{ParentID: issue.ID})
				m.detail = newDetailModel(issue, children, m.width, m.height)
			}
		}
		return m, nil

	case commentAddedMsg:
		m.statusMsg = "Comment added"
		m.reloadIssues()
		m.screen = m.prevScreen
		if m.screen == screenDetail {
			issue, err := m.tracker.LoadIssue(msg.issueID)
			if err == nil {
				children := tracker.FilterIssues(m.issues, tracker.FilterOptions{ParentID: issue.ID})
				m.detail = newDetailModel(issue, children, m.width, m.height)
			}
		}
		return m, nil

	case issueCreatedMsg:
		m.statusMsg = fmt.Sprintf("Created %s: %s", msg.id, msg.title)
		m.reloadIssues()
		m.screen = screenList
		return m, nil

	case editorDoneMsg:
		m.reloadIssues()
		if msg.err != nil {
			m.statusMsg = "Editor: " + msg.err.Error()
		} else {
			m.statusMsg = "Issue updated"
			if m.prevScreen == screenDetail {
				issue, err := m.tracker.LoadIssue(msg.issueID)
				if err == nil {
					children := tracker.FilterIssues(m.issues, tracker.FilterOptions{ParentID: issue.ID})
					m.detail = newDetailModel(issue, children, m.width, m.height)
				}
			}
		}
		m.screen = m.prevScreen
		return m, nil

	case issueLinkMsg:
		if msg.err != nil {
			m.statusMsg = "Link: " + msg.err.Error()
		} else {
			m.statusMsg = "Parent linked"
			m.reloadIssues()
		}
		m.screen = m.prevScreen
		if m.screen == screenDetail {
			issue, err := m.tracker.LoadIssue(msg.childID)
			if err == nil {
				children := tracker.FilterIssues(m.issues, tracker.FilterOptions{ParentID: issue.ID})
				m.detail = newDetailModel(issue, children, m.width, m.height)
			}
		}
		return m, nil

	case issueUnlinkMsg:
		if msg.err != nil {
			m.statusMsg = "Unlink: " + msg.err.Error()
		} else {
			m.statusMsg = "Parent unlinked"
			m.reloadIssues()
		}
		if m.screen == screenDetail {
			issue, err := m.tracker.LoadIssue(msg.childID)
			if err == nil {
				children := tracker.FilterIssues(m.issues, tracker.FilterOptions{ParentID: issue.ID})
				m.detail = newDetailModel(issue, children, m.width, m.height)
			}
		}
		return m, nil

	case tea.KeyMsg:
		if m.statusMsg != "" {
			m.statusMsg = ""
		}

		if msg.String() == "?" && !m.list.searching {
			m.help.visible = !m.help.visible
			return m, nil
		}
		if m.help.visible {
			return m, nil
		}

		if m.screen == screenHistory {
			switch msg.String() {
			case "esc":
				m.screen = screenDetail
				return m, nil
			case "q", "ctrl+c":
				return m, tea.Quit
			}
			var cmd tea.Cmd
			m.history, cmd = m.history.Update(msg)
			return m, cmd
		}

		if m.screen == screenStatus {
			switch msg.String() {
			case "esc":
				m.screen = m.prevScreen
				return m, nil
			case "enter":
				target := m.statusPicker.selected()
				if target == "" {
					return m, nil
				}
				return m, m.executeStatusChange(m.statusPicker.issueID, target)
			case "q", "ctrl+c":
				return m, tea.Quit
			}
			var cmd tea.Cmd
			m.statusPicker, cmd = m.statusPicker.Update(msg)
			return m, cmd
		}

		if m.screen == screenComment {
			switch msg.String() {
			case "ctrl+d":
				text := m.commentInput.textarea.Value()
				if text == "" {
					m.screen = m.prevScreen
					return m, nil
				}
				issueID := m.commentInput.issueID
				return m, m.executeAddComment(issueID, text)
			case "esc":
				m.screen = m.prevScreen
				return m, nil
			}
			var cmd tea.Cmd
			m.commentInput, cmd = m.commentInput.Update(msg)
			return m, cmd
		}

		if m.screen == screenCreate {
			switch msg.String() {
			case "ctrl+d":
				title := m.createForm.title()
				if title == "" {
					m.statusMsg = "Title is required"
					m.screen = screenList
					return m, nil
				}
				return m, m.executeCreateIssue()
			case "esc":
				m.screen = screenList
				return m, nil
			case "q", "ctrl+c":
				return m, tea.Quit
			}
			var cmd tea.Cmd
			m.createForm, cmd = m.createForm.Update(msg)
			return m, cmd
		}

		if m.screen == screenLink {
			switch msg.String() {
			case "enter":
				parentID := m.linkInput.input.Value()
				if parentID == "" {
					m.screen = m.prevScreen
					return m, nil
				}
				childID := m.linkInput.childID
				return m, m.executeLinkIssue(childID, parentID)
			case "esc":
				m.screen = m.prevScreen
				return m, nil
			case "q", "ctrl+c":
				return m, tea.Quit
			}
			var cmd tea.Cmd
			m.linkInput.input, cmd = m.linkInput.input.Update(msg)
			return m, cmd
		}

		if m.screen == screenConfirm {
			var confirmed bool
			m.confirm, confirmed, _ = m.confirm.Update(msg)
			if msg.String() == "y" || msg.String() == "Y" {
				if confirmed {
					return m, m.executeStatusChange(m.confirm.issueID, "cancelled")
				}
			} else if msg.String() == "n" || msg.String() == "N" || msg.String() == "esc" {
				m.screen = m.prevScreen
				return m, nil
			}
			return m, nil
		}

		if m.screen == screenList && m.list.searching {
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "s":
			return m.openStatusPicker()
		case "c":
			return m.openComment()
		case "a":
			return m.quickStatus("active")
		case "d":
			return m.quickStatus("done")
		case "r":
			return m.quickStatus("review")
		case "n":
			if m.screen == screenList {
				m.screen = screenCreate
				m.createForm = newCreateModel()
				return m, m.createForm.inputs[0].Focus()
			}
		case "e":
			return m.openEditor()
		case "p":
			return m.openLinkInput()
		case "h":
			if m.screen == screenDetail {
				issueID := m.detail.issue.ID
				events, err := m.tracker.LoadEvents(issueID)
				if err != nil {
					m.statusMsg = "History: " + err.Error()
					return m, nil
				}
				m.history = newHistoryModel(
					issueID[:min(6, len(issueID))],
					events, m.width, m.height,
				)
				m.screen = screenHistory
				return m, nil
			}
		case "P":
			return m.executeUnlink()
		case "x":
			return m.openConfirm()
		case "enter":
			if m.screen == screenList {
				row := m.list.table.SelectedRow()
				if row != nil {
					issue, err := m.tracker.LoadIssue(row[0])
					if err == nil {
						children := tracker.FilterIssues(m.issues, tracker.FilterOptions{ParentID: issue.ID})
						m.detail = newDetailModel(issue, children, m.width, m.height)
						m.screen = screenDetail
					}
				}
				return m, nil
			}
		case "esc":
			if m.screen == screenDetail {
				m.screen = screenList
				return m, nil
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.resize(msg.Width, msg.Height)
	}

	switch m.screen {
	case screenList:
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	case screenDetail:
		var cmd tea.Cmd
		m.detail, cmd = m.detail.Update(msg)
		return m, cmd
	case screenHistory:
		var cmd tea.Cmd
		m.history, cmd = m.history.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *rootModel) reloadIssues() {
	issues, err := m.tracker.ListIssues()
	if err == nil {
		tracker.SortIssues(issues, "priority")
		m.issues = issues
		m.list = newListModel(issues, m.width)
	}
}

func (m rootModel) selectedIssue() (id, title string, ok bool) {
	switch m.screen {
	case screenList:
		row := m.list.table.SelectedRow()
		if row == nil {
			return "", "", false
		}
		return row[0], row[4], true
	case screenDetail:
		return m.detail.issue.ID[:min(6, len(m.detail.issue.ID))], m.detail.issue.Title, true
	default:
		return "", "", false
	}
}

func (m rootModel) openComment() (tea.Model, tea.Cmd) {
	id, title, ok := m.selectedIssue()
	if !ok {
		return m, nil
	}
	m.prevScreen = m.screen
	m.screen = screenComment
	m.commentInput = newCommentModel(id, title)
	return m, m.commentInput.textarea.Focus()
}

func (m rootModel) quickStatus(status string) (tea.Model, tea.Cmd) {
	id, _, ok := m.selectedIssue()
	if !ok {
		return m, nil
	}
	m.prevScreen = m.screen
	return m, m.executeStatusChange(id, status)
}

func (m rootModel) openConfirm() (tea.Model, tea.Cmd) {
	id, _, ok := m.selectedIssue()
	if !ok {
		return m, nil
	}
	m.prevScreen = m.screen
	m.screen = screenConfirm
	m.confirm = newConfirmModel(id, "Cancel")
	return m, nil
}

func (m rootModel) executeAddComment(issueID, text string) tea.Cmd {
	return func() tea.Msg {
		_, err := m.tracker.AddComment(issueID, text, m.user)
		if err != nil {
			return commentAddedMsg{issueID: issueID}
		}
		return commentAddedMsg{issueID: issueID}
	}
}

func (m rootModel) openStatusPicker() (tea.Model, tea.Cmd) {
	var issueID, currentStatus string

	switch m.screen {
	case screenList:
		row := m.list.table.SelectedRow()
		if row == nil {
			return m, nil
		}
		issueID = row[0]
		currentStatus = row[1]
	case screenDetail:
		issueID = m.detail.issue.ID[:min(6, len(m.detail.issue.ID))]
		currentStatus = m.detail.issue.Status
	default:
		return m, nil
	}

	transitions := m.tracker.Config.Transitions[currentStatus]
	if len(transitions) == 0 {
		m.statusMsg = "No transitions available"
		return m, nil
	}

	m.prevScreen = m.screen
	m.screen = screenStatus
	m.statusPicker = newStatusPicker(issueID, currentStatus, transitions)
	return m, nil
}

func (m rootModel) executeCreateIssue() tea.Cmd {
	f := m.createForm
	return func() tea.Msg {
		issue, err := m.tracker.CreateIssue(
			f.title(), f.description(), m.user,
			f.priority(), f.labels(), f.issueType(),
			f.parentID(), m.user,
		)
		if err != nil {
			return issueCreatedMsg{title: "error: " + err.Error()}
		}
		return issueCreatedMsg{id: issue.ID[:min(6, len(issue.ID))], title: issue.Title}
	}
}

func (m rootModel) openEditor() (tea.Model, tea.Cmd) {
	id, _, ok := m.selectedIssue()
	if !ok {
		return m, nil
	}
	issue, err := m.tracker.LoadIssue(id)
	if err != nil {
		m.statusMsg = "Load: " + err.Error()
		return m, nil
	}

	tmpFile, err := os.CreateTemp("", "work-edit-*.md")
	if err != nil {
		m.statusMsg = "Temp file: " + err.Error()
		return m, nil
	}
	content := editor.MarshalIssue(issue)
	if _, err := tmpFile.WriteString(content); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
		m.statusMsg = "Write: " + err.Error()
		return m, nil
	}
	_ = tmpFile.Close()

	m.prevScreen = m.screen
	path := tmpFile.Name()
	issueID := issue.ID

	c := exec.Command(m.editor, path)
	return m, tea.ExecProcess(c, func(err error) tea.Msg {
		defer func() { _ = os.Remove(path) }()
		if err != nil {
			return editorDoneMsg{issueID: issueID, err: err}
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return editorDoneMsg{issueID: issueID, err: readErr}
		}
		title, desc, issueType, assignee, priority, labels, parseErr := editor.UnmarshalIssue(string(data))
		if parseErr != nil {
			return editorDoneMsg{issueID: issueID, err: parseErr}
		}
		issue.Title = title
		issue.Description = desc
		issue.Type = issueType
		issue.Assignee = assignee
		issue.Priority = priority
		issue.Labels = labels
		issue.Updated = time.Now().UTC()
		if saveErr := m.tracker.SaveIssue(issue); saveErr != nil {
			return editorDoneMsg{issueID: issueID, err: saveErr}
		}
		return editorDoneMsg{issueID: issueID}
	})
}

func (m rootModel) openLinkInput() (tea.Model, tea.Cmd) {
	id, _, ok := m.selectedIssue()
	if !ok {
		return m, nil
	}
	m.prevScreen = m.screen
	m.screen = screenLink
	m.linkInput = newLinkModel(id)
	return m, m.linkInput.input.Focus()
}

func (m rootModel) executeUnlink() (tea.Model, tea.Cmd) {
	id, _, ok := m.selectedIssue()
	if !ok {
		return m, nil
	}
	return m, func() tea.Msg {
		_, err := m.tracker.UnlinkIssue(id, m.user)
		return issueUnlinkMsg{childID: id, err: err}
	}
}

func (m rootModel) executeLinkIssue(childID, parentID string) tea.Cmd {
	return func() tea.Msg {
		_, err := m.tracker.LinkIssue(childID, parentID, m.user)
		return issueLinkMsg{childID: childID, err: err}
	}
}

func (m rootModel) executeStatusChange(issueID, newStatus string) tea.Cmd {
	return func() tea.Msg {
		_, err := m.tracker.SetStatus(issueID, newStatus, m.user)
		if err != nil {
			return statusChangedMsg{issueID: issueID, status: "error: " + err.Error()}
		}
		if newStatus == "done" || newStatus == "cancelled" {
			_ = m.tracker.CompactIssue(issueID)
		}
		return statusChangedMsg{issueID: issueID, status: newStatus}
	}
}

func (m rootModel) View() string {
	if m.help.visible {
		header := titleStyle.Render("work — help")
		return header + m.help.View(m.width)
	}

	var header, body, footer string

	switch m.screen {
	case screenList:
		header = titleStyle.Render("work — issues")
		body = m.list.View()
		footer = helpStyle.Render("  ?:help  n:new  s:status  /:search  q:quit")
	case screenDetail:
		header = titleStyle.Render("work — " + m.detail.issue.Title)
		body = m.detail.View()
		footer = helpStyle.Render("  ?:help  esc:back  s:status  h:history  q:quit")
	case screenStatus:
		header = titleStyle.Render("work — change status")
		body = m.statusPicker.View()
		footer = helpStyle.Render("  j/k:navigate  enter:select  esc:cancel")
	case screenComment:
		header = titleStyle.Render("work — add comment")
		body = m.commentInput.View()
	case screenConfirm:
		header = titleStyle.Render("work — confirm")
		body = m.confirm.View()
	case screenCreate:
		header = titleStyle.Render("work — new issue")
		body = m.createForm.View()
	case screenLink:
		header = titleStyle.Render("work — link parent")
		body = fmt.Sprintf("\n  Parent ID for %s:\n\n  %s\n\n%s",
			m.linkInput.childID,
			m.linkInput.input.View(),
			helpStyle.Render("  enter:link  esc:cancel"))
	case screenHistory:
		header = titleStyle.Render("work — history")
		body = m.history.View()
		footer = helpStyle.Render("  j/k:scroll  esc:back  q:quit")
	}

	if m.statusMsg != "" {
		footer = helpStyle.Render("  " + m.statusMsg)
	}

	return header + "\n" + body + "\n" + footer
}
