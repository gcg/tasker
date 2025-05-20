# Go Charm Todo CLI

A simple command-line interface (CLI) application for managing your todo list, built with Go and the Charm libraries (Bubble Tea, Lipgloss). Tasks are stored in a `todo.yaml` file in the directory where the application is run.

## Features

- Create tasks and subtasks.
- Edit existing tasks.
- Mark tasks as complete or pending.
- Delete tasks.
- Interactive TUI powered by Bubble Tea.
- Data persistence in `todo.yaml`.

## Prerequisites

- Go (version 1.18 or later recommended)
- Make (for using Makefile commands)

## Installation

1.  **Clone the repository (if you haven't already):**
    ```bash
    # git clone <repository-url>
    # cd todo-cli
    ```

2.  **Install using the Makefile:**
    This command will compile and install the binary into your Go bin directory (e.g., `$GOPATH/bin` or `$HOME/go/bin`).
    ```bash
    make install
    ```
    Ensure your Go bin directory is in your system's PATH.

## Usage

### Running the Application

-   If installed via `make install`, you can run it directly:
    ```bash
    todo-cli
    ```
-   To run from the source code directory for development:
    ```bash
    make run
    ```
    or
    ```bash
    go run main.go
    ```

The application will create or use a `todo.yaml` file in the current working directory to store tasks.

### Keybindings (TUI)

Once the application is running:

-   **Navigation:**
    -   `↑`/`k`: Move cursor up
    -   `↓`/`j`: Move cursor down
    -   `pgup`/`pgdn`: Page up/down

-   **Task Operations:**
    -   `a`: Add a new top-level task.
    -   `s`: Add a subtask to the currently selected task (cannot add a subtask to another subtask).
    -   `e`: Edit the selected task's description.
    -   `enter` / `space`: Toggle the selected task's status (pending/completed).
    -   `d` / `backspace`: Delete the selected task (and its subtasks if any).

-   **Application:**
    -   `q` / `ctrl+c`: Quit the application.
    -   `ctrl+s`: Manually save tasks (tasks are also saved automatically after most operations).

### Input Mode (when adding/editing tasks)

-   `enter`: Confirm and save the task description.
-   `esc`: Cancel adding/editing and return to navigation.

## Development

-   **Update dependencies:**
    ```bash
    make update
    ```
-   **Clean build artifacts:**
    ```bash
    make clean
    ```

## Contributing

Feel free to open issues or submit pull requests!
