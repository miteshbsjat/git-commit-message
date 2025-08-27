# Git Commit Message using AI (Ollama)

The Go program that generates a single-line Git commit message based on your staged changes using Ollama.

The program works by:

1.  Reading your Ollama configuration from `~/.config/git_commit_message/config.yaml`.
2.  Executing the `git diff --staged` command to get the code changes.
3.  Sending these changes to your local Ollama instance with a specific prompt.
4.  Printing the cleaned, single-line commit message returned by the model.

-----

### \#\# Prerequisites

1.  **Go:** You need to have the Go programming language installed.
2.  **Git:** The program must be run within a Git repository.
3.  **Ollama:** You must have a running Ollama instance with a model downloaded (e.g., `ollama pull llama3`).

-----

### \#\# Step 1: Create the Configuration File

First, create the directory and the configuration file. This file will store your Ollama settings so you don't have to hardcode them.

Open your terminal and run the following commands:

```bash
# Create the directory
mkdir -p ~/.config/git_commit_message

# Create and open the config file with a text editor
vi ~/.config/git_commit_message/config.yaml
```

Now, paste the following content into the `config.yaml` file. Adjust the values to match your setup.

```yaml
ollama_url: "http://localhost:11434"
model: "llama3" # Or any other model you prefer, like 'mistral' or 'codellama'
temperature: 0.5 # A value between 0.0 (deterministic) and 1.0 (creative)
```

-----

### \#\# Step 2: Build and Use the Program

1.  **Build the Executable:**

    ```bash
    go build -o git-commit-message .
    ```

    This creates an executable file named `git-commit`. You can move this to a location in your system's `PATH` (like `/usr/local/bin`) to make it accessible everywhere.

    ```bash
    # Optional: Move to a universal location
    sudo cp git-commit-message /usr/local/bin/
    ```

2.  **Usage:**
    Now, from any git repository on your machine:

      * Stage your changes (`git add .`).
      * Run the program: `git-commit`
      * It will print a suggested commit message. You can then copy it and use it with `git commit -m "..."`.

#### **Git Alias for Quick Commits**

For even faster workflow, you can create a Git alias that runs the program and immediately creates the commit. Add this to your global `.gitconfig` file:

```bash
git config --global alias.aic '! "f () { git commit -am \"$(/usr/local/bin/git-commit-message | tail -1)\"; }; f"'
```

Now, you can simply run `git aic` to stage all changes and automatically generate and apply the commit message. 🚀
