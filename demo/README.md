# Demo Directory

This directory contains VHS (Video Handling System) demo scripts for creating animated demonstrations of the provider-explorer TUI.

## Prerequisites

Install VHS and dependencies by Charm:
```bash
# Using Go
go install github.com/charmbracelet/vhs@latest

# Using Homebrew (macOS)
brew install vhs
# also install dependencies ffmpeg, ttyd

# Using mise (this repo .mise.toml)
mise install # includes vhs and ffmpeg, ttyd only works for linux
brew install ttyd

# Using package managers (Linux)
# See: https://github.com/charmbracelet/vhs#installation
```

## Recording Demos

### Generate the Demo GIF
```bash
# From the project root
vhs demo/demo.tape
```

This will create `demo/demo.gif` showing the complete provider-explorer workflow.

### Demo Script Overview

The `demo.tape` script demonstrates:

1. **Startup**: Load AWS provider configuration
2. **Navigation**: Browse from providers → resource types → specific resources
3. **Filtering**: Use fuzzy search to find S3 buckets and EC2 instances
4. **Tree Views**: Switch between Arguments and Attributes views
5. **Selection**: Select multiple schema attributes/arguments
6. **Export**: Generate HCL code (outputs and variables)
7. **Workflow**: Complete end-to-end usage scenarios

### Customizing the Demo

Edit `demo.tape` to:
- Change terminal dimensions (`Set Width/Height`)
- Adjust timing (`Sleep` durations)
- Modify font size (`Set FontSize`)
- Add/remove workflow steps

Key VHS commands used:
- `Type "text"` - Simulate typing
- `Enter` - Press Enter key
- `Sleep 1s` - Pause for timing
- `Ctrl+a` - Send control sequences
- `Escape` - Send Escape key
- `Backspace` - Delete characters

## Demo Maintenance

### When to Update
- After UI changes that affect the demo workflow
- When adding new features to showcase
- After significant navigation or interaction changes

### Best Practices
- Keep demos under 60 seconds when possible
- Use consistent timing for professional appearance
- Test the demo script before committing changes
- Ensure the fixture (`fixtures/example-aws`) supports the demo workflow

### Troubleshooting

**Demo doesn't complete properly:**
- Check that `fixtures/example-aws` contains required resources
- Verify timing - some operations may need longer `Sleep` durations
- Ensure provider-explorer binary is built and accessible

**Output looks wrong:**
- Check terminal dimensions match your target display
- Adjust font size for readability
- Verify VHS is using the correct shell

## Integration

The generated `demo.gif` is designed to be embedded in:
- Main project README
- Documentation sites
- Social media demonstrations
- Project presentations

## Reference

For more VHS documentation and examples, see:
- [VHS Repository](https://github.com/charmbracelet/vhs)
- [Building Bubble Tea Programs Blog](https://leg100.github.io/en/posts/building-bubbletea-programs/#10-record-demos-and-screenshots-on-vhs)