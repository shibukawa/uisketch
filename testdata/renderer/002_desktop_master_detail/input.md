---
id: renderer.002_desktop_master_detail
type: uisketch
title: Desktop Workbench
---

# Desktop Workbench

## Layout

```uisketch
window:
  id: desktop-workbench
  title: Desktop Workbench
  menu:
    - File
    - Edit
    - View
  children:
    - splitter:
        orientation: horizontal
        sizes: [30, 70]
        children:
          - section:
              title: Filters
              children:
                - list:
                    label: Saved Filters
                - checkbox:
                    label: Critical only
          - section:
              title: Equipment Detail
              children:
                - grid:
                    columns: 2
                    children:
                      - input:
                          label: Equipment Name
                      - textarea:
                          label: Notes
                      - combobox:
                          label: Status
                          options:
                            - Enabled
                            - Disabled
                      - radio:
                          label: Manual mode
                      - toggle:
                          label: Enabled
                      - slider:
                          label: Priority
                - image:
                    label: Equipment Photo
```
