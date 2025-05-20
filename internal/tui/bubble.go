package tui

import (
	"fmt"
	"os"
	"strings"
	// "time" // Not used directly in this snippet, but good to keep if model.Task uses it

	"todo-cli/internal/model"
	"todo-cli/internal/service" // To call task logic
	"todo-cli/internal/store"   // For saving tasks

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	docStyle         = lipgloss.NewStyle().Margin(1, 2)
	titleStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	itemStyle        = lipgloss.NewStyle().PaddingLeft(2)
	selectedStyle    = lipgloss.NewStyle().PaddingLeft(0).Foreground(lipgloss.Color("75")).Bold(true) // Cyan-ish
	completedStyle   = lipgloss.NewStyle().Strikethrough(true).Foreground(lipgloss.Color("240"))      // Grey
	helpStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Padding(1, 0)
	inputPromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205")) // Magenta/Purple for prompt
	inputValueStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("250")) // Light grey for value
	errorStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)             // Red for errors
)

const (
	defaultTaskFile = "todo.yaml"
	dateTimeLayout  = "2006-01-02 15:04"
)

// item represents a list item for Bubble Tea.
// It wraps a model.Task to implement the list.Item interface.
type item struct {
	task model.Task
	// We might store a reference to the parent task if needed for context
	// parent *model.Task
}

func (i item) Title() string {
	prefix := "  " // For top-level tasks
	if i.task.Status == model.StatusCompleted {
		return completedStyle.Render(prefix + i.task.Description)
	}
	return prefix + i.task.Description
}

// Description for list.Item interface - used for displaying details or filtering.
// We'll show dates here.
func (i item) Description() string {
	var details []string
	details = append(details, "Created: "+i.task.CreatedAt.Format(dateTimeLayout))
	if i.task.CompletedAt != nil {
		details = append(details, "Completed: "+i.task.CompletedAt.Format(dateTimeLayout))
	}
	desc := strings.Join(details, " | ")
	if i.task.Status == model.StatusCompleted {
		return completedStyle.Render("  " + desc)
	}
	return "  " + desc
}
func (i item) FilterValue() string { return i.task.Description }

// subItem is used for subtasks to allow different styling/indentation.
type subItem struct {
	item // Embed item to reuse its methods
}

func (si subItem) Title() string {
	prefix := "    └─ " // Indent subtasks
	if si.item.task.Status == model.StatusCompleted {
		return completedStyle.Render(prefix + si.item.task.Description)
	}
	return prefix + si.item.task.Description
}

func (si subItem) Description() string {
	var details []string
	details = append(details, "Created: "+si.item.task.CreatedAt.Format(dateTimeLayout))
	if si.item.task.CompletedAt != nil {
		details = append(details, "Completed: "+si.item.task.CompletedAt.Format(dateTimeLayout))
	}
	desc := strings.Join(details, " | ")
	if si.item.task.Status == model.StatusCompleted {
		return completedStyle.Render("      " + desc) // Indent subtask description
	}
	return "      " + desc // Indent subtask description
}

// listMode represents the different interaction modes of the TUI.
type listMode int

const (
	modeNavigating listMode = iota // Normal list browsing
	modeAdding                     // Adding a new task/subtask
	modeEditing                    // Editing an existing task
	// modeDeleting // Maybe a confirmation step
)

// listModel is the Bubble Tea model for our todo application.
type listModel struct {
	list         list.Model        // Bubble Tea's list component
	tasks        []model.Task      // Our actual task data
	textInput    textinput.Model   // For adding/editing tasks
	mode         listMode          // Current interaction mode
	editingID    string            // ID of task being edited, or parent ID for new subtask
	errorMessage string            // To display errors to the user
	width, height int               // Terminal dimensions
	quitting     bool
}

// keyMap defines additional keybindings.
type keyMap struct {
	add      key.Binding
	subtask  key.Binding
	edit     key.Binding
	complete key.Binding
	delete   key.Binding
	save     key.Binding
	confirm  key.Binding
	cancel   key.Binding
}

var defaultKeyMap = keyMap{
	add:      key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "add task")),
	subtask:  key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "add subtask")),
	edit:     key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit")),
	complete: key.NewBinding(key.WithKeys("enter", " "), key.WithHelp("enter/space", "toggle complete")),
	delete:   key.NewBinding(key.WithKeys("d", "backspace"), key.WithHelp("d/bksp", "delete")),
	save:     key.NewBinding(key.WithKeys("ctrl+s"), key.WithHelp("ctrl+s", "save tasks")), // Manual save
	confirm:  key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirm")),
	cancel:   key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
}

// InitialModel creates the initial model for the Bubble Tea program.
func InitialModel(tasks []model.Task) listModel {
	l := list.New(nil, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Charm Todo List"
	l.Styles.Title = titleStyle
	// l.SetShowStatusBar(false) // We'll make our own help/status
	l.SetFilteringEnabled(false) // For now
	l.SetShowPagination(true)

	// Setup keybindings for the list itself (navigation)
	// Some default keybindings might be okay, but we can customize
	l.KeyMap.Quit = key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q/ctrl+c", "quit"))

	// Our custom keybindings (subset for the list itself)
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{defaultKeyMap.add, defaultKeyMap.subtask, defaultKeyMap.edit, defaultKeyMap.complete, defaultKeyMap.delete}
	}
	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{defaultKeyMap.add, defaultKeyMap.subtask, defaultKeyMap.edit, defaultKeyMap.complete, defaultKeyMap.delete, defaultKeyMap.save, l.KeyMap.Quit}
	}

	ti := textinput.New()
	ti.Placeholder = "Task description..."
	ti.CharLimit = 156
	ti.Width = 50 // Adjust as needed
	ti.PromptStyle = inputPromptStyle
	ti.TextStyle = inputValueStyle

	m := listModel{list: l, tasks: tasks, textInput: ti, mode: modeNavigating}
	m.syncListItems() // Load initial tasks into the bubble list
	return m
}

// syncListItems converts our []model.Task into []list.Item for the bubble list.
// It handles expanding subtasks.
func (m *listModel) syncListItems() {
	bubbleItems := make([]list.Item, 0)
	for _, task := range m.tasks {
		bubbleItems = append(bubbleItems, item{task: task})
		if len(task.Subtasks) > 0 {
			for _, sub := range task.Subtasks {
				// Create a subItem which embeds item but can have different rendering
				bubbleItems = append(bubbleItems, subItem{item{task: sub}})
			}
		}
	}
	m.list.SetItems(bubbleItems)
}

func (m listModel) Init() tea.Cmd {
	return textinput.Blink // Start the text input blinking (if visible)
}

func (m listModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width-docStyle.GetHorizontalPadding(), msg.Height-docStyle.GetVerticalPadding()-lipgloss.Height(m.headerView())-lipgloss.Height(m.footerView()))
		return m, nil

	case tea.KeyMsg:
		m.errorMessage = "" // Clear error on new key press
		if m.mode == modeAdding || m.mode == modeEditing {
			return m.handleInputMode(msg)
		}
		return m.handleNavigationMode(msg)
	}

	// Propagate updates to the list component if not handled above
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m listModel) handleInputMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch {
	case key.Matches(msg, defaultKeyMap.cancel):
		m.mode = modeNavigating
		m.textInput.Reset()
		m.textInput.Blur()
		return m, nil
	case key.Matches(msg, defaultKeyMap.confirm): // Enter confirms add/edit
		inputValue := strings.TrimSpace(m.textInput.Value())
		if inputValue == "" {
			m.errorMessage = "Description cannot be empty."
			return m, nil
		}

		if m.mode == modeAdding {
			if m.editingID != "" { // If editingID is set, it's a subtask for this parent ID
				m.tasks, _ = service.AddSubtask(m.tasks, m.editingID, inputValue)
				m.editingID = "" // Clear parent ID
			} else {
				m.tasks = service.AddTask(m.tasks, inputValue)
			}
		} else if m.mode == modeEditing {
			m.tasks, _ = service.EditTask(m.tasks, m.editingID, inputValue)
			m.editingID = "" // Clear editing ID
		}

		m.syncListItems()
		err := store.SaveTasks(defaultTaskFile, m.tasks) // Persist changes
		if err != nil {
			m.errorMessage = "Failed to save tasks: " + err.Error()
		}
		m.mode = modeNavigating
		m.textInput.Reset()
		m.textInput.Blur()
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m listModel) handleNavigationMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var currentItem item // To store the actual task data, not just list.Item

	selectedBubbleItem := m.list.SelectedItem()
	if selectedBubbleItem != nil {
		switch selected := selectedBubbleItem.(type) {
		case item:
			currentItem = selected
		case subItem:
			currentItem = selected.item
		}
	}

	switch {
	case key.Matches(msg, m.list.KeyMap.Quit):
		m.quitting = true
		return m, tea.Quit

	case key.Matches(msg, defaultKeyMap.add):
		m.mode = modeAdding
		m.editingID = "" // Clear any previous editing/parent ID
		m.textInput.Placeholder = "New task description..."
		m.textInput.Focus()
		return m, textinput.Blink

	case key.Matches(msg, defaultKeyMap.subtask):
		if selectedBubbleItem == nil {
			m.errorMessage = "Select a parent task first."
			return m, nil
		}
		// Ensure it's not a subItem itself that's selected for adding another subtask (sub-sub-tasks not supported yet by logic)
		if _, ok := selectedBubbleItem.(subItem); ok {
			m.errorMessage = "Cannot add a subtask to another subtask in this version."
			return m, nil
		}
		m.editingID = currentItem.task.ID // Store parent ID
		m.mode = modeAdding
		m.textInput.Placeholder = "New subtask for '" + currentItem.task.Description + "'..."
		m.textInput.Focus()
		return m, textinput.Blink

	case key.Matches(msg, defaultKeyMap.edit):
		if selectedBubbleItem == nil {
			m.errorMessage = "No task selected to edit."
			return m, nil
		}
		m.editingID = currentItem.task.ID
		m.mode = modeEditing
		m.textInput.SetValue(currentItem.task.Description)
		m.textInput.Placeholder = "Edit task description..."
		m.textInput.Focus()
		m.textInput.SetCursor(len(currentItem.task.Description))
		return m, textinput.Blink

	case key.Matches(msg, defaultKeyMap.complete):
		if selectedBubbleItem == nil {
			m.errorMessage = "No task selected."
			return m, nil
		}
		m.tasks, _ = service.ToggleTaskStatus(m.tasks, currentItem.task.ID)
		m.syncListItems()
		err := store.SaveTasks(defaultTaskFile, m.tasks)
		if err != nil {
			m.errorMessage = "Failed to save tasks: " + err.Error()
		}
		return m, nil

	case key.Matches(msg, defaultKeyMap.delete):
		if selectedBubbleItem == nil {
			m.errorMessage = "No task selected to delete."
			return m, nil
		}
		m.tasks, _ = service.RemoveTask(m.tasks, currentItem.task.ID)
		m.syncListItems()
		// Ensure cursor is valid after deletion
		if m.list.Index() >= len(m.list.Items()) && len(m.list.Items()) > 0 {
			m.list.Select(len(m.list.Items()) - 1)
		} else if len(m.list.Items()) == 0 {
			// Handle empty list, maybe move cursor or set a specific state
		}
		err := store.SaveTasks(defaultTaskFile, m.tasks)
		if err != nil {
			m.errorMessage = "Failed to save tasks: " + err.Error()
		}
		return m, nil

	case key.Matches(msg, defaultKeyMap.save):
		err := store.SaveTasks(defaultTaskFile, m.tasks)
		if err != nil {
			m.errorMessage = "Failed to save tasks: " + err.Error()
		} else {
			m.errorMessage = "Tasks saved successfully!" // Provide feedback
		}
		return m, nil // Optionally return a cmd for a timed message
	}

	// Default: propagate to list navigation
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m listModel) headerView() string {
	return titleStyle.Render(m.list.Title)
}

func (m listModel) footerView() string {
	var help strings.Builder
	if m.mode == modeAdding || m.mode == modeEditing {
		help.WriteString(helpStyle.Render(fmt.Sprintf("%s | %s", defaultKeyMap.confirm.Help().Key, defaultKeyMap.confirm.Help().Desc)))
		help.WriteString(helpStyle.Render(fmt.Sprintf(" | %s | %s", defaultKeyMap.cancel.Help().Key, defaultKeyMap.cancel.Help().Desc)))
	} else {
		// Show navigation help keys
		// help.WriteString(m.list.ShortHelpView(m.list.ShortHelpKeys()))
		// Corrected: ShortHelpKeys returns []key.Binding, pass it to list.ShortHelpView
		// However, list.ShortHelpView is not a public method.
		// We need to construct the help string manually or use what list provides if available.
		// For now, let's iterate over the keys we defined for the list.
		// Or, use m.list.HelpView() if it provides a good summary for navigation.
		// The list component internally manages its help view. Let's try accessing its built-in help.
		// This might require m.list.SetShowHelp(true) or similar if not on by default.
		// For a custom approach:
		var navHelp []string
		for _, k := range m.list.AdditionalShortHelpKeys() {
			navHelp = append(navHelp, k.Help().Key+": "+k.Help().Desc)
		}
		navHelp = append(navHelp, m.list.KeyMap.Quit.Help().Key+": "+m.list.KeyMap.Quit.Help().Desc)
		help.WriteString(helpStyle.Render(strings.Join(navHelp, ", ")))

	}
	if m.errorMessage != "" {
		return errorStyle.Render(m.errorMessage) + "\n" + help.String()
	}
	return helpStyle.Render(help.String())
}

func (m listModel) View() string {
	if m.quitting {
		return ""
	}

	var s strings.Builder
	s.WriteString(m.headerView())
	s.WriteString("\n") // Add a newline after the header

	if m.mode == modeAdding || m.mode == modeEditing {
		// Only show text input when adding or editing
		s.WriteString(m.textInput.View())
		s.WriteString("\n\n") // Add some space after the input field
	} else if m.mode == modeNavigating {
		// Only show the list when navigating
		s.WriteString(docStyle.Render(m.list.View()))
	}
    // If there were other modes, they would need handling here or fall through if no specific view.

	s.WriteString(m.footerView())
	return s.String()
}

// StartTeaProgram is the main entry point for the TUI.
func StartTeaProgram(initialTasks []model.Task) error {
	m := InitialModel(initialTasks)
	p := tea.NewProgram(m, tea.WithAltScreen()) // Use AltScreen for cleaner exit

	// It's important to propagate the final model from p.Run()
	// if you need to access data from it after the program exits.
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("error running Bubble Tea program: %w", err)
	}

	// Persist tasks one last time if any changes were made and not caught by quit
	// This is a fallback, ideally saves happen on action.
	if fm, ok := finalModel.(listModel); ok {
		if !fm.quitting { // Avoid saving if quit was intentional and handled
			// This might be redundant if all actions save, but good for safety
			if errSave := store.SaveTasks(defaultTaskFile, fm.tasks); errSave != nil {
				// Use errSave to avoid shadowing the outer err
				fmt.Fprintf(os.Stderr, "Error saving tasks on exit: %v\n", errSave)
			}
		}
	}
	return nil
}
