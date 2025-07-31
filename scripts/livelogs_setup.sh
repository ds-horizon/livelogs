#!/bin/bash

LIVELOGS_FUNCTION_NAME="livelogs()"

# Function block to be injected
LIVELOGS_FUNCTION=$(cat << 'EOF'

#####LIVELOGS_INSTALLATION_BEGIN#####
livelogs() {
    if curl -Is https://artifactory.dream11.com/prod/central-log-management/ll-installer.py | head -n 1 | grep -q "200"; then
        curl -s -o ~/.livelogs/livelogs_installer.py https://artifactory.dream11.com/prod/central-log-management/ll-installer.py
    elif [ -f ~/.livelogs/livelogs_installer.py ]; then
        echo "⚠️ Failed to download the livelogs installer using existing installer."
    else
        echo "❌ Failed to download the livelogs installer. Press Enter to close..." >&2
        read
        exit 1
    fi

    if command -v python >/dev/null 2>&1; then
        python ~/.livelogs/livelogs_installer.py "$@"
    elif command -v python3 >/dev/null 2>&1; then
        python3 ~/.livelogs/livelogs_installer.py "$@"
    else
        echo "❌ Python not found. Install it to run livelogs. Press Enter to close..." >&2
        read
        exit 1
    fi
}
#####LIVELOGS_INSTALLATION_END#####

EOF
)

# List of common shell config files
CONFIG_FILES=(
    "$HOME/.zshrc"
    "$HOME/.bashrc"
    "$HOME/.bash_profile"
    "$HOME/.profile"
)

throw_exception() {
    local message="$1"
    echo "$message" >&2
    exit 0
}

# Add function block to all config files if not already present
for FILE in "${CONFIG_FILES[@]}"; do
    if [ -f "$FILE" ] && grep -q "$LIVELOGS_FUNCTION_NAME" "$FILE"; then
        echo "LiveLogs already configured in $FILE"
    elif [ -f "$FILE" ] && [ -w "$FILE" ]; then
        echo "Adding livelogs function to $FILE"
        echo "$LIVELOGS_FUNCTION" >> "$FILE"
    else
        throw_exception "Unable to write to $FILE — permission denied. You may need elevated privileges (try using sudo)."
    fi
done

# Ensure ~/.livelogs directory exists
mkdir -p ~/.livelogs

# Reload the config for the current shell only (silently)
SHELL_NAME=$(basename "$SHELL")
echo "SHELL: $SHELL_NAME"

case "$SHELL_NAME" in
    zsh)
        if [ -f "$HOME/.zshrc" ]; then
            source "$HOME/.zshrc" >/dev/null 2>&1
        fi
        ;;
    bash)
        if [ -f "$HOME/.bashrc" ]; then
            source "$HOME/.bashrc" >/dev/null 2>&1
        fi
        if [ -f "$HOME/.bash_profile" ]; then
            source "$HOME/.bash_profile" >/dev/null 2>&1
        fi
        ;;
    *)
        if [ -f "$HOME/.profile" ]; then
            source "$HOME/.profile" >/dev/null 2>&1
        fi
        ;;
esac

echo "✅ Setup completed. You can now run: livelogs"
