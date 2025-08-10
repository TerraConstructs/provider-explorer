Output ./demo.gif
Set Shell "bash"
Set FontSize 14
Set Width 1200
Set Height 800
Set Framerate 24
Set Padding 5

Hide
Type `../dist/provider-explorer_darwin_arm64_v8.0/provider-explorer ../fixtures/example-aws` Enter
Sleep 1s
Show

# Navigate to Resources (aws provider auto-selected)
Sleep 2s
Type "j" Sleep 0.5s
Enter Sleep 0.5s

# Filter for bucket
Type "/" Sleep 0.5s
Type "bucket" Sleep 1s
Enter Sleep 0.5s
Enter Sleep 1s

# Switch to Attributes view and select all
Type "a" Sleep 1s
Ctrl+a Sleep 0.5s

# Export attributes with custom name
Type "e" Sleep 1s
Backspace Backspace Backspace Backspace
Type "foo" Sleep 0.5s
Enter Sleep 2s
Type "j" Sleep 0.5s
Type "j" Sleep 0.5s

# Copy to clipboard
Type "c" Sleep 2s

# Navigate back to resources
Escape Sleep 0.5s
Escape Sleep 0.5s
Escape Sleep 0.5s
Escape Sleep 1s

# Filter for instances
Type "/" Sleep 0.5s
Type "inst" Sleep 1s
Enter Sleep 0.5s
Enter Sleep 1s

# Switch to arguments and select items
Type "a" Sleep 1s
Type " " Sleep 0.5s
Type "j" Sleep 0.2s
Type "j" Sleep 0.2s
Type "j" Sleep 0.2s
Type "j" Sleep 0.2s
Type " " Sleep 0.5s

# Export arguments as variables
Type "e" Sleep 2s