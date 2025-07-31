#!/bin/bash

LIVLOGS_FUNCTION_NAME="livelogs()"

# Function block to be injected
LIVLOGS_FUNCTION=$(cat << 'EOF'

#####LIVELOGS_INSTALLATION_BEGIN#####
livelogs() {
    if curl -Is https://artifactory.dream11.com/prod/central-log-management/ll-installer.py | head -n 1 | grep -q "200"; then
        curl -s -o ~/.livelogs/livelogs_installer.py https://artifactory.dream11.com/prod/central-log-management/ll-installer.py
    elif [ -f ~/.livelogs/livelogs_installer.py ]; then
        echo "⚠ Failed to download the livelogs installer using existing installer."
    else
        echo "✗ Failed to download the livelogs installer. Press Enter to close..."
        read
        exit 1
    fi
    python ~/.livelogs/livelogs_installer.py "$@"
}
#####LIVELOGS_INSTALLATION_END#####

EOF
)

# List of common shell config files
CONFIG_FILES=(
    "$HOME/.zshrc"
    "$HOME/.bashrc"
    "$HOME/.profile"
)

# Add function block to all config files if not already present
for FILE in "${CONFIG_FILES[@]}"; do
    if [ -f "$FILE" ] && grep -q "$LIVLOGS_FUNCTION_NAME" "$FILE"; then
        echo "ℹ️  LiveLogs already configured in $FILE"
    else
        echo "➕ Adding livelogs function to $FILE"
        echo "$LIVLOGS_FUNCTION" >> "$FILE"
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
        ;;
    *)
        if [ -f "$HOME/.profile" ]; then
            source "$HOME/.profile" >/dev/null 2>&1
        fi
        ;;
esac

echo "✅ Setup complete. You can now run: livelogs"
