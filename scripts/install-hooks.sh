#!/bin/bash

# Install git hooks for the SBS project
# Run this script from the project root directory

set -e

echo "🔧 Installing git hooks for SBS project..."

# Create hooks directory if it doesn't exist
mkdir -p .git/hooks

# Install pre-commit hook
cat > .git/hooks/pre-commit << 'EOF'
#!/bin/bash

# Pre-commit hook to automatically format Go code using make fmt

set -e  # Exit on any error

echo "🔍 Pre-commit hook: Running code formatter..."

# Check if make command exists
if ! command -v make &> /dev/null; then
    echo "❌ Error: 'make' command not found. Please install make."
    exit 1
fi

# Get list of staged Go files before formatting
staged_go_files=$(git diff --cached --name-only --diff-filter=AM | grep '\.go$' || true)

if [[ -z "$staged_go_files" ]]; then
    echo "ℹ️  No Go files to format."
    exit 0
fi

echo "📝 Formatting Go files..."
make fmt

# Check if any staged files were modified by the formatter
modified_files=""
for file in $staged_go_files; do
    if ! git diff --exit-code --quiet "$file" 2>/dev/null; then
        modified_files="$modified_files $file"
    fi
done

if [[ -n "$modified_files" ]]; then
    echo "✅ Code was reformatted. Adding formatted files to commit..."
    
    # Add the formatted files back to the staging area
    for file in $modified_files; do
        git add "$file"
        echo "   📄 Added formatted file: $file"
    done
    
    echo "✅ Formatted files have been added to the commit."
else
    echo "✅ No formatting changes needed."
fi

echo "🚀 Pre-commit hook completed successfully."
exit 0
EOF

# Make the hook executable
chmod +x .git/hooks/pre-commit

echo "✅ Pre-commit hook installed successfully!"
echo ""
echo "The pre-commit hook will now automatically format Go code before each commit."
echo "To disable temporarily, use: git commit --no-verify"