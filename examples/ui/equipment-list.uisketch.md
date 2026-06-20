---
id: screen.equipment-list
type: uisketch
title: Equipment List
screen:
  id: screen.equipment-list
---

# Equipment List

Shows equipment status and the main refresh operation.

## Layout

```yaml
browser:
  id: equipment-list
  title: Equipment List
  address: https://example.internal/equipment
  children:
    - vstack:
        children:
          - hstack:
              children:
                - button:
                    id: refresh
                    action: action.refresh-equipment
                    label: Refresh
                - button:
                    id: create-alert
                    action: action.create-alert
                    label: Create Alert
          - table:
              id: equipment-table
              columns:
                - Equipment
                - Status
          - label: Ready
```
