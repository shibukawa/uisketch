---
id: renderer.001_browser_dashboard
type: uisketch
title: Equipment Dashboard
---

# Equipment Dashboard

## Layout

```uisketch
browser:
  id: equipment-dashboard
  title: Equipment Dashboard
  address: https://example.internal/equipment
  children:
    - vstack:
        children:
          - hstack:
              children:
                - button:
                    label: Refresh
                    action: action.refresh-equipment
                - button:
                    label: CSV
                    action: action.export-equipment
                - button:
                    label: Add
                    action: action.create-alert
                - button:
                    label: Alerts
                    badge: 12
                - toggle:
                    label: Offline
                - button:
                    label: Detail
                    anchor: equipment-detail
          - table:
              id: equipment-table
              columns:
                - Equipment
                - Status
                - Owner
          - label: Updated just now
          - label:
              label: Verify warning status with operations.
              note: Confirm thresholds with operations.
```
