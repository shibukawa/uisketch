---
id: renderer.003_dialog_tabs_review
type: uisketch
title: Alert Review Dialog
---

# Alert Review Dialog

## Layout

```uisketch
dialog:
  id: alert-review-dialog
  title: Alert Review Dialog
  children:
    - tabs:
        labels:
          - [Summary]
          - History
        children:
          section:
            title: Summary
            children:
              - label:
                  label: Supervisor approval required.
                  note: Requires supervisor review before closing.
    - grid:
        children:
          - custom:
              name: alert-severity-map
              purpose: Shows affected equipment by severity.
          - label:
              label: Should escalation be automatic after 30 minutes?
              note: Resolve escalation policy.
  buttons:
    - button:
        label: Cancel
    - button:
        label: Approve
```
