---
id: renderer.005-mobile-grid-spacer
type: uisketch
title: Mobile Catalog
---

# Mobile Catalog

## Layout

```uisketch
mobile:
  id: mobile-catalog
  title: Catalog
  menu:
    - Home
    - Search
    - Settings
  children:
    - hstack:
        children:
          - label: Catalog
          - spacer
          - button: Filter
          - button: Sort
    - grid:
        columns: 3
        children:
          - image: Item A
          - image: Item B
          - image: Item C
          - image: Item D
```
