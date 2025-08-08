# Attribution

This project includes code derived from other open source projects. We acknowledge and are grateful for their contributions:

## Bubbly Tree Component

The tree component in `internal/ui/tree/` contains code derived from the [Bubbly project](https://github.com/anchore/bubbly):

- **Project**: Anchore Bubbly
- **URL**: https://github.com/anchore/bubbly
- **License**: Apache License 2.0
- **Files adapted**:
  - `internal/ui/tree/model.go` - derived from `bubbles/tree/model.go`
  - `internal/ui/tree/visible_model.go` - derived from `visible_model.go`
  - `internal/ui/tree/testutil/run_model.go` - derived from `bubbles/internal/testutil/run_model.go`
  - `internal/ui/tree/model_test.go` - test structure derived from `bubbles/tree/model_test.go`

### Changes Made

We have extended the original bubbly tree component with the following enhancements:
- Added scrolling support with viewport management for height-constrained displays
- Implemented selection indicators with checkboxes ("[ ]" and "[x]")
- Added schema-specific node types for Terraform provider schemas
- Enhanced keyboard navigation and interaction patterns
- Comprehensive test coverage using go-snaps

### Original License

The original code from Bubbly is licensed under the Apache License 2.0:

```
Copyright (c) 2024 Anchore, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```

All modified files retain the original copyright headers as required by the Apache License 2.0.