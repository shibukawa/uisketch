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
  children:
    - menubar:
        children:
          - label: File
          - label: Edit
          - label: View
    - split-pane:
        children:
          - sidebar:
              title: Filters
              children:
                - list:
                    label: Saved Filters
                - checkbox:
                    label: Critical only
          - section:
              title: Equipment Detail
              children:
                - table-layout:
                    children:
                      - input:
                          label: Equipment Name
                      - switch:
                          label: Enabled
                      - slider:
                          label: Priority
                - image:
                    label: Equipment Photo
```
