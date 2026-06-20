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
              - note: Requires supervisor review before closing.
    - grid:
        children:
          - custom-component:
              name: alert-severity-map
              purpose: Shows affected equipment by severity.
          - review: Should escalation be automatic after 30 minutes?
          - button:
              label: Approve
          - button:
              label: Cancel
```
