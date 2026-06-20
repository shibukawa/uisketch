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
                - icon-button:
                    label: CSV
                    action: action.export-equipment
                - floating-action-button:
                    label: Add
                    action: action.create-alert
                - badge-button:
                    label: Alerts
                    badge: 12
                - toggle-button:
                    label: Offline
                - link:
                    label: Detail
                    anchor: equipment-detail
          - table:
              id: equipment-table
              columns:
                - Equipment
                - Status
                - Owner
          - label: Updated just now
          - hint: Verify warning status with operations.
```
