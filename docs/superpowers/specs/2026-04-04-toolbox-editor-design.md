# Spec: Toolbox Detail Editor (Side Panel)

**Status:** Draft
**Date:** 2026-04-04

## Goal
Implement a comprehensive, inline editor for the Toolbox node configuration in the right side panel.

## Architecture

### 1. Built-in Tools Editor
- **UI:** A searchable multi-select or list of checkboxes.
- **Tools:** `google_search`, `code_interpreter`, `file_system_browser`.
- **Action:** Toggle tool ID in the `config.tools` array.

### 2. MCP Servers Editor
- **UI:** An editable list of MCP configurations.
- **Fields:** `name`, `command`, `args` (space-separated input), and `env` (key-value pairs).
- **Actions:** "Add MCP Server", "Remove", and "Edit Fields".

### 3. Custom Functions Editor
- **UI:** An editable list of function signatures.
- **Fields:**
  - `name`: String
  - `description`: String
  - `parameters`: JSON Schema Editor (Monaco-like textarea with validation).
- **Actions:** "Add Function", "Remove", and "Edit JSON Schema".

## UI Layout (within Side Panel)
The Toolbox configuration will be split into three accordion-style sections:
1. **Built-in Tools** (Count badge)
2. **MCP Servers** (Count badge)
3. **Custom Functions** (Count badge)

## Next Steps
1. Create `front_end/src/components/editors/ToolListEditor.tsx`.
2. Create `front_end/src/components/editors/MCPServerEditor.tsx`.
3. Create `front_end/src/components/editors/CustomFunctionEditor.tsx`.
4. Integrate these into `SidePanel.tsx`.
