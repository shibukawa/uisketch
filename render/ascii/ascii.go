package ascii

import (
	"strconv"
	"strings"

	"uisketch/sketch"
)

const (
	defaultWidth  = 80
	defaultHeight = 24
	minWidth      = 6
	minHeight     = 3
)

type canvas struct {
	w, h  int
	cells [][]rune
}

type rect struct {
	x, y, w, h int
}

type notes []string

func Render(doc *sketch.Document) string {
	if doc == nil || doc.Root == nil {
		return ""
	}
	c := newCanvas(defaultWidth, max(defaultHeight, requiredRootHeight(doc.Root)))
	var ns notes
	renderRoot(c, rect{0, 0, c.w, c.h}, doc, &ns)
	out := c.String()
	if len(ns) == 0 {
		return out
	}
	var b strings.Builder
	b.WriteString(out)
	for i, note := range ns {
		b.WriteString("\n")
		b.WriteString("[")
		b.WriteString(intString(i + 1))
		b.WriteString("] ")
		b.WriteString(note)
	}
	return b.String()
}

func requiredRootHeight(root *sketch.Node) int {
	content := childrenVStackHeight(root.Children)
	switch root.Type {
	case "browser":
		return content + 5
	case "window":
		menu := 0
		if len(root.Menu) > 0 {
			menu = 2
		}
		return content + menu + 4
	case "dialog", "menu":
		return content + 4
	case "mobile":
		menu := 0
		if len(root.Menu) > 0 {
			menu = 3
		}
		return content + menu + 7
	default:
		return intrinsicHeight(root) + 2
	}
}

func renderRoot(c *canvas, r rect, doc *sketch.Document, ns *notes) {
	root := doc.Root
	if root.Type == "mobile" {
		renderMobileRoot(c, r, doc, ns)
		return
	}
	drawBox(c, r)
	title := rootTitle(doc)
	switch root.Type {
	case "browser":
		writeText(c, r.x+2, r.y+1, title)
		writeRight(c, r.x+r.w-2, r.y+1, "○ ○ ○")
		writeText(c, r.x+2, r.y+2, "← →")
		if root.Address != "" {
			writeText(c, r.x+8, r.y+2, "[ "+root.Address+" ]")
		}
		drawHLine(c, r.x, r.y+3, r.w)
		renderChildrenAsVStack(c, rect{r.x + 1, r.y + 4, r.w - 2, r.h - 5}, root.Children, ns)
	case "window":
		writeText(c, r.x+2, r.y+1, title)
		writeRight(c, r.x+r.w-2, r.y+1, "○ ○ ○")
		drawHLine(c, r.x, r.y+2, r.w)
		content := rect{r.x + 1, r.y + 3, r.w - 2, r.h - 4}
		if len(root.Menu) > 0 && content.h > 2 {
			writeText(c, content.x+1, content.y, strings.Join(root.Menu, "   "))
			drawHLine(c, r.x, content.y+1, r.w)
			content.y += 2
			content.h -= 2
		}
		renderChildrenAsVStack(c, content, root.Children, ns)
	case "dialog":
		writeCentered(c, r.x+1, r.y+1, r.w-2, title)
		writeRight(c, r.x+r.w-2, r.y+1, "○")
		drawHLine(c, r.x, r.y+2, r.w)
		renderChildrenAsVStack(c, rect{r.x + 1, r.y + 3, r.w - 2, r.h - 4}, root.Children, ns)
	case "menu":
		writeText(c, r.x+2, r.y+1, title)
		drawHLine(c, r.x, r.y+2, r.w)
		renderChildrenAsVStack(c, rect{r.x + 1, r.y + 3, r.w - 2, r.h - 4}, root.Children, ns)
	default:
		renderNode(c, inset(r, 1), root, ns)
	}
}

func renderMobileRoot(c *canvas, r rect, doc *sketch.Document, ns *notes) {
	root := doc.Root
	title := rootTitle(doc)
	phoneW := min(32, max(12, r.w-4))
	phone := rect{r.x + max(0, (r.w-phoneW)/2), r.y, phoneW, r.h}
	drawBox(c, phone)
	writeCentered(c, phone.x+1, phone.y, phone.w-2, "━━━━")
	if title != "" {
		writeCentered(c, phone.x+1, phone.y+2, phone.w-2, title)
	}
	homeY := phone.y + phone.h - 2
	if homeY > phone.y {
		writeCentered(c, phone.x+1, homeY, phone.w-2, "────")
	}
	contentTop := 4
	if title == "" {
		contentTop = 3
	}
	content := rect{phone.x + 2, phone.y + contentTop, phone.w - 4, max(0, phone.h-contentTop-3)}
	if len(root.Menu) > 0 && content.h > 3 {
		menuY := content.y + content.h - 2
		drawHLine(c, phone.x+1, menuY-1, phone.w-2)
		writeCentered(c, phone.x+2, menuY, phone.w-4, strings.Join(root.Menu, " | "))
		content.h -= 3
	}
	renderChildrenAsVStack(c, content, root.Children, ns)
}

func renderNode(c *canvas, r rect, n *sketch.Node, ns *notes) {
	if n == nil || r.w <= 0 || r.h <= 0 {
		return
	}
	switch n.Type {
	case "spacer":
		return
	case "vstack":
		renderChildrenAsVStackWithHeights(c, r, n.Children, n.Heights, ns)
	case "hstack":
		renderChildrenAsHStack(c, r, n.Children, n.Widths, ns)
	case "grid":
		renderGrid(c, r, n, ns)
	case "splitter":
		renderSplitter(c, r, n, ns)
	case "tabs":
		renderTabs(c, r, n, ns)
	case "section":
		drawBox(c, r)
		writeText(c, r.x+2, r.y, " "+labelFor(n)+" ")
		renderChildrenAsVStack(c, inset(r, 1), n.Children, ns)
	case "button":
		renderButton(c, r, n)
	case "label":
		if n.Note != "" {
			*ns = append(*ns, n.Note)
			writeTextClipped(c, r.x, r.y, r.w, "["+intString(len(*ns))+"]"+labelFor(n))
			return
		}
		writeTextClipped(c, r.x, r.y, r.w, labelFor(n))
	case "input":
		renderInput(c, r, n)
	case "checkbox":
		writeText(c, r.x, r.y, "[ ] "+labelFor(n))
	case "radio":
		writeText(c, r.x, r.y, "( ) "+labelFor(n))
	case "toggle":
		writeText(c, r.x, r.y, "[off] "+labelFor(n))
	case "textarea":
		renderTextarea(c, r, n)
	case "combobox":
		renderCombobox(c, r, n)
	case "slider":
		writeText(c, r.x, r.y, labelFor(n)+" ─────○────")
	case "table":
		renderTable(c, r, n)
	case "list":
		renderList(c, r, n, "List")
	case "tree":
		renderList(c, r, n, "Tree")
	case "calendar":
		renderList(c, r, n, "Calendar")
	case "badge":
		writeText(c, r.x, r.y, "("+labelFor(n)+")")
	case "image", "custom":
		renderImagePlaceholder(c, r, n)
	default:
		drawBox(c, r)
		writeCentered(c, r.x+1, r.y+r.h/2, r.w-2, labelFor(n))
	}
	if n.Note != "" {
		*ns = append(*ns, n.Note)
		renderNoteMarker(c, noteTargetRect(r, n), len(*ns), n.Type == "button")
	}
}

func renderChildrenAsVStack(c *canvas, r rect, children []*sketch.Node, ns *notes) {
	renderChildrenAsVStackWithHeights(c, r, children, nil, ns)
}

func renderChildrenAsVStackWithHeights(c *canvas, r rect, children []*sketch.Node, slots []sketch.SizeSlot, ns *notes) {
	if len(children) == 0 || r.w <= 0 || r.h <= 0 {
		return
	}
	heights := distributeVertical(r.h, children, slots)
	y := r.y
	for i, child := range children {
		h := heights[i]
		renderNode(c, rect{r.x, y, r.w, h}, child, ns)
		y += h
	}
}

func renderChildrenAsHStack(c *canvas, r rect, children []*sketch.Node, slots []sketch.SizeSlot, ns *notes) {
	renderChildrenAsHStackWithMode(c, r, children, slots, false, ns)
}

func renderChildrenAsHStackFill(c *canvas, r rect, children []*sketch.Node, slots []sketch.SizeSlot, ns *notes) {
	renderChildrenAsHStackWithMode(c, r, children, slots, true, ns)
}

func renderChildrenAsHStackWithMode(c *canvas, r rect, children []*sketch.Node, slots []sketch.SizeSlot, fillRemainder bool, ns *notes) {
	if len(children) == 0 || r.w <= 0 || r.h <= 0 {
		return
	}
	gap := 1
	if len(children) == 1 {
		gap = 0
	}
	widths := distributeHorizontal(max(1, r.w-gap*(len(children)-1)), children, slots, fillRemainder)
	x := r.x
	for i, child := range children {
		w := widths[i]
		renderNode(c, rect{x, r.y, w, r.h}, child, ns)
		x += w + gap
	}
}

func renderSingleChild(c *canvas, r rect, n *sketch.Node) {
	if len(n.Children) == 0 {
		return
	}
	var ns notes
	renderNode(c, r, n.Children[0], &ns)
}

func renderGrid(c *canvas, r rect, n *sketch.Node, ns *notes) {
	children := n.Children
	if len(children) == 0 {
		return
	}
	cols := 2
	if n.Type == "grid" && n.GridColumns > 0 {
		cols = n.GridColumns
	} else if len(children) == 1 {
		cols = 1
	}
	rows := (len(children) + cols - 1) / cols
	cellH := max(1, r.h/rows)
	cellW := max(1, r.w/cols)
	for i, child := range children {
		col := i % cols
		row := i / cols
		renderNode(c, rect{r.x + col*cellW, r.y + row*cellH, cellW, cellH}, child, ns)
	}
}

func renderSplitter(c *canvas, r rect, n *sketch.Node, ns *notes) {
	if n.Orientation == "vertical" {
		renderChildrenAsVStackWithHeights(c, r, n.Children, defaultSizes(n.Sizes), ns)
		return
	}
	renderChildrenAsHStackFill(c, r, n.Children, defaultSizes(n.Sizes), ns)
}

func defaultSizes(slots []sketch.SizeSlot) []sketch.SizeSlot {
	if len(slots) > 0 {
		return slots
	}
	return []sketch.SizeSlot{{Percent: 25}, {Percent: 75}}
}

func renderTabs(c *canvas, r rect, n *sketch.Node, ns *notes) {
	if r.h < 4 {
		writeText(c, r.x, r.y, strings.Join(tabLabelTexts(n), " | "))
		return
	}
	labels := tabLabels(n)
	active := activeTabIndex(labels)
	panel := rect{r.x, r.y + 2, r.w, r.h - 2}
	drawBox(c, panel)
	redrawTabsOverPanel(c, r, labels, active)
	renderActiveTabBody(c, inset(panel, 1), n, active, ns)
}

func redrawTabsOverPanel(c *canvas, r rect, labels []sketch.TabLabel, active int) {
	tabX := r.x
	for i, label := range labels {
		text := label.Text
		if i == active {
			text = "*" + text + "*"
		}
		w := min(max(len([]rune(text))+4, 10), max(10, r.w/3))
		drawTab(c, rect{tabX, r.y, w, 3}, i == active, i == 0)
		writeCentered(c, tabX+1, r.y+1, w-2, text)
		tabX += w
		if tabX >= r.x+r.w {
			break
		}
	}
}

func drawTab(c *canvas, r rect, active, first bool) {
	if r.w < 2 || r.h < 2 {
		return
	}
	x2 := r.x + r.w - 1
	y2 := r.y + r.h - 1
	put(c, r.x, r.y, '┌')
	put(c, x2, r.y, '┐')
	for x := r.x + 1; x < x2; x++ {
		put(c, x, r.y, '─')
	}
	put(c, r.x, r.y+1, '│')
	put(c, x2, r.y+1, '│')
	if active {
		if first {
			put(c, r.x, y2, '│')
		} else {
			put(c, r.x, y2, '┘')
		}
		for x := r.x + 1; x < x2; x++ {
			put(c, x, y2, ' ')
		}
		put(c, x2, y2, '└')
		return
	}
	if first {
		put(c, r.x, y2, '├')
	} else {
		put(c, r.x, y2, '┴')
	}
	put(c, x2, y2, '┴')
	for x := r.x + 1; x < x2; x++ {
		put(c, x, y2, '─')
	}
}

func renderButton(c *canvas, r rect, n *sketch.Node) {
	box := buttonRect(r, n)
	label := labelFor(n)
	if box.h <= 1 {
		writeText(c, box.x, box.y, "└─"+label+"─┘")
		return
	}
	drawBox(c, box)
	writeCentered(c, box.x+1, box.y+1, box.w-2, label)
	if n.Badge != "" {
		writeRight(c, box.x+box.w-1, box.y, "("+n.Badge+")")
	}
}

func buttonRect(r rect, n *sketch.Node) rect {
	label := labelFor(n)
	w := min(max(len([]rune(label))+4, minWidth), r.w)
	if n.Badge != "" {
		w = min(max(w, len([]rune(label))+len([]rune(n.Badge))+6), r.w)
	}
	h := min(3, r.h)
	return rect{r.x + max(0, r.w-w), r.y + max(0, r.h-h), w, h}
}

func noteTargetRect(r rect, n *sketch.Node) rect {
	if n.Type == "button" {
		return buttonRect(r, n)
	}
	return r
}

func renderNoteMarker(c *canvas, target rect, index int, outside bool) {
	marker := "[" + intString(index) + "]"
	y := target.y
	markerLen := len([]rune(marker))
	if !outside {
		writeText(c, target.x, y, marker)
		return
	}
	markerX := target.x + max(0, target.w-markerLen-10)
	markerX = max(0, target.x-markerLen-8)
	writeText(c, markerX, y, marker)
	lineStart := markerX + markerLen
	lineEnd := target.x + target.w - 1
	if lineEnd > lineStart {
		for x := lineStart; x < lineEnd; x++ {
			put(c, x, y, '─')
		}
		put(c, lineEnd, y, '┐')
	}
}

func renderInput(c *canvas, r rect, n *sketch.Node) {
	label := labelFor(n)
	writeText(c, r.x, r.y, label)
	if r.h >= 2 {
		drawBox(c, rect{r.x, r.y + 1, min(r.w, max(12, len([]rune(label))+8)), 3})
	}
}

func renderTextarea(c *canvas, r rect, n *sketch.Node) {
	label := labelFor(n)
	writeText(c, r.x, r.y, label)
	if r.h >= 2 {
		drawBox(c, rect{r.x, r.y + 1, min(r.w, max(16, len([]rune(label))+8)), min(5, max(3, r.h-1))})
	}
}

func renderCombobox(c *canvas, r rect, n *sketch.Node) {
	label := labelFor(n)
	writeText(c, r.x, r.y, label)
	if r.h >= 2 {
		box := rect{r.x, r.y + 1, min(r.w, max(16, len([]rune(label))+8)), 3}
		drawBox(c, box)
		writeRight(c, box.x+box.w-2, box.y+1, "v")
	}
}

func renderTable(c *canvas, r rect, n *sketch.Node) {
	drawBox(c, r)
	if len(n.Columns) == 0 {
		writeText(c, r.x+2, r.y+1, labelFor(n))
		return
	}
	writeText(c, r.x+2, r.y+1, strings.Join(n.Columns, " │ "))
	if r.h > 3 {
		drawHLine(c, r.x, r.y+2, r.w)
	}
}

func renderImagePlaceholder(c *canvas, r rect, n *sketch.Node) {
	drawBox(c, r)
	writeCentered(c, r.x+1, r.y+r.h/2, r.w-2, labelFor(n))
	if r.h >= 4 && r.w >= 8 {
		put(c, r.x+2, r.y+1, '╲')
		put(c, r.x+r.w-3, r.y+1, '╱')
		put(c, r.x+2, r.y+r.h-2, '╱')
		put(c, r.x+r.w-3, r.y+r.h-2, '╲')
	}
}

func renderList(c *canvas, r rect, n *sketch.Node, kind string) {
	drawBox(c, r)
	writeText(c, r.x+2, r.y+1, kind+": "+labelFor(n))
	for y := r.y + 2; y < r.y+r.h-1; y++ {
		writeText(c, r.x+2, y, "·")
	}
}

func distributeVertical(total int, children []*sketch.Node, slots []sketch.SizeSlot) []int {
	if proportional := distributeSlots(total, len(children), slots); len(proportional) > 0 {
		return proportional
	}
	out := make([]int, len(children))
	for i, child := range children {
		out[i] = min(total, intrinsicHeight(child))
	}
	fixed := sumInts(out)
	remaining := max(0, total-fixed)
	if fixed > total {
		return shrinkToFit(out, total)
	}
	if len(out) > 0 {
		out[len(out)-1] += remaining
	}
	return out
}

func distributeHorizontal(total int, children []*sketch.Node, slots []sketch.SizeSlot, fillRemainder bool) []int {
	if proportional := distributeSlots(total, len(children), slots); len(proportional) > 0 {
		return proportional
	}
	if spacerCount := countSpacers(children); spacerCount > 0 {
		return distributeWithSpacers(total, children, spacerCount)
	}
	out := make([]int, len(children))
	for i, child := range children {
		out[i] = min(total, intrinsicWidth(child))
	}
	fixed := sumInts(out)
	remaining := max(0, total-fixed)
	if fixed > total {
		return shrinkToFit(out, total)
	}
	if fillRemainder && len(out) > 0 {
		out[len(out)-1] += remaining
	}
	return out
}

func countSpacers(children []*sketch.Node) int {
	count := 0
	for _, child := range children {
		if child != nil && child.Type == "spacer" {
			count++
		}
	}
	return count
}

func distributeWithSpacers(total int, children []*sketch.Node, spacerCount int) []int {
	out := make([]int, len(children))
	for i, child := range children {
		if child == nil || child.Type == "spacer" {
			continue
		}
		out[i] = min(total, intrinsicWidth(child))
	}
	fixed := sumInts(out)
	if fixed > total {
		return shrinkToFit(out, total)
	}
	remaining := max(0, total-fixed)
	each := remaining / spacerCount
	extra := remaining % spacerCount
	for i, child := range children {
		if child == nil || child.Type != "spacer" {
			continue
		}
		out[i] = each
		if extra > 0 {
			out[i]++
			extra--
		}
	}
	return out
}

func distributeSlots(total, count int, slots []sketch.SizeSlot) []int {
	if count == 0 || len(slots) != count {
		return nil
	}
	out := make([]int, count)
	numeric := 0
	stars := 0
	for _, slot := range slots {
		if slot.Star {
			stars++
			continue
		}
		numeric += slot.Percent
	}
	remainingPercent := max(0, 100-numeric)
	remainingPixels := total
	for i, slot := range slots {
		if slot.Star {
			continue
		}
		out[i] = max(1, total*slot.Percent/100)
		remainingPixels -= out[i]
	}
	if stars > 0 {
		eachPercent := remainingPercent / stars
		for i, slot := range slots {
			if slot.Star {
				out[i] = max(1, total*eachPercent/100)
				remainingPixels -= out[i]
			}
		}
	}
	if len(out) > 0 {
		out[len(out)-1] += remainingPixels
	}
	return shrinkToFit(out, total)
}

func sumInts(values []int) int {
	sum := 0
	for _, value := range values {
		sum += value
	}
	return sum
}

func shrinkToFit(sizes []int, total int) []int {
	sum := 0
	for _, size := range sizes {
		sum += size
	}
	if sum <= total {
		return sizes
	}
	out := make([]int, len(sizes))
	remaining := total
	for i, size := range sizes {
		next := max(1, total*size/sum)
		if i == len(sizes)-1 {
			next = remaining
		}
		out[i] = max(1, next)
		remaining -= out[i]
	}
	return out
}

func intrinsicWidth(n *sketch.Node) int {
	switch n.Type {
	case "spacer":
		return 0
	case "button":
		if n.Badge != "" {
			return len([]rune(labelFor(n))) + len([]rune(n.Badge)) + 6
		}
		return len([]rune(labelFor(n))) + 4
	case "checkbox", "radio", "toggle", "slider", "label", "badge":
		return len([]rune(labelFor(n))) + 8
	case "input", "textarea", "combobox":
		return len([]rune(labelFor(n))) + 12
	default:
		return 24
	}
}

func intrinsicHeight(n *sketch.Node) int {
	switch n.Type {
	case "spacer":
		return 1
	case "button":
		return 4
	case "input", "combobox":
		return 3
	case "textarea":
		return 5
	case "hstack":
		return max(4, maxChildHeight(n.Children))
	case "vstack":
		return childrenVStackHeight(n.Children)
	case "label", "checkbox", "radio", "toggle", "slider", "badge":
		return 1
	case "table":
		return 5
	case "tabs":
		return max(12, activeTabHeight(n)+4)
	case "grid":
		return gridHeight(n)
	case "section":
		return max(8, childrenVStackHeight(n.Children)+2)
	case "list", "tree", "calendar", "image", "custom":
		return 8
	case "splitter":
		if n.Orientation == "vertical" {
			return max(12, childrenVStackHeight(n.Children))
		}
		return max(12, maxChildHeight(n.Children))
	default:
		return 4
	}
}

func childrenVStackHeight(children []*sketch.Node) int {
	total := 0
	for _, child := range children {
		total += intrinsicHeight(child)
	}
	return total
}

func maxChildHeight(children []*sketch.Node) int {
	maxHeight := 0
	for _, child := range children {
		maxHeight = max(maxHeight, intrinsicHeight(child))
	}
	return maxHeight
}

func gridHeight(n *sketch.Node) int {
	if len(n.Children) == 0 {
		return 8
	}
	cols := n.GridColumns
	if cols <= 0 {
		cols = 2
	}
	rows := (len(n.Children) + cols - 1) / cols
	return max(8, rows*max(1, maxChildHeight(n.Children)))
}

func activeTabHeight(n *sketch.Node) int {
	active := activeTabIndex(n.Labels)
	if len(n.Labels) == 0 && len(n.Children) > active {
		return intrinsicHeight(n.Children[active])
	}
	if len(n.Children) == 1 {
		return intrinsicHeight(n.Children[0])
	}
	return childrenVStackHeight(n.Children)
}

func distribute(total int, parts int) []int {
	out := make([]int, parts)
	if parts == 0 {
		return out
	}
	remaining := total
	for i := 0; i < parts; i++ {
		size := total / parts
		if i == parts-1 {
			size = remaining
		}
		out[i] = max(1, size)
		remaining -= out[i]
	}
	return out
}

func isCompact(n *sketch.Node) bool {
	switch n.Type {
	case "label", "checkbox", "radio", "toggle", "slider", "badge":
		return true
	case "hstack":
		return true
	default:
		return false
	}
}

func intString(value int) string {
	return strconv.Itoa(value)
}

func tabLabels(n *sketch.Node) []sketch.TabLabel {
	if len(n.Labels) > 0 {
		return n.Labels
	}
	labels := make([]sketch.TabLabel, 0, len(n.Children))
	for i, child := range n.Children {
		labels = append(labels, sketch.TabLabel{Text: labelFor(child), Selected: i == 0})
	}
	return labels
}

func tabLabelTexts(n *sketch.Node) []string {
	labels := tabLabels(n)
	out := make([]string, len(labels))
	for i, label := range labels {
		out[i] = label.Text
	}
	return out
}

func activeTabIndex(labels []sketch.TabLabel) int {
	for i, label := range labels {
		if label.Selected {
			return i
		}
	}
	return 0
}

func renderActiveTabBody(c *canvas, r rect, n *sketch.Node, active int, ns *notes) {
	if len(n.Labels) == 0 && len(n.Children) > active {
		renderNode(c, r, n.Children[active], ns)
		return
	}
	if len(n.Children) == 1 {
		renderNode(c, r, n.Children[0], ns)
		return
	}
	renderChildrenAsVStack(c, r, n.Children, ns)
}

func labelFor(n *sketch.Node) string {
	switch {
	case n.Title != "":
		return n.Title
	case n.Label != "":
		return n.Label
	case n.Name != "":
		return n.Name
	case n.Hint != "":
		return n.Hint
	case n.Purpose != "":
		return n.Purpose
	case n.Action != "":
		return n.Action
	case n.Type != "":
		return n.Type
	default:
		return "item"
	}
}

func rootTitle(doc *sketch.Document) string {
	if doc == nil || doc.Root == nil {
		return ""
	}
	if doc.Root.Title != "" {
		return doc.Root.Title
	}
	if doc.Title != "" {
		return doc.Title
	}
	return labelFor(doc.Root)
}

func newCanvas(w, h int) *canvas {
	c := &canvas{w: w, h: h, cells: make([][]rune, h)}
	for y := range c.cells {
		c.cells[y] = make([]rune, w)
		for x := range c.cells[y] {
			c.cells[y][x] = ' '
		}
	}
	return c
}

func (c *canvas) String() string {
	lines := make([]string, c.h)
	for y := range c.cells {
		lines[y] = strings.TrimRight(string(c.cells[y]), " ")
	}
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return strings.Join(lines, "\n")
}

func drawBox(c *canvas, r rect) {
	if r.w < 2 || r.h < 2 {
		return
	}
	x2 := r.x + r.w - 1
	y2 := r.y + r.h - 1
	put(c, r.x, r.y, '┌')
	put(c, x2, r.y, '┐')
	put(c, r.x, y2, '└')
	put(c, x2, y2, '┘')
	for x := r.x + 1; x < x2; x++ {
		put(c, x, r.y, '─')
		put(c, x, y2, '─')
	}
	for y := r.y + 1; y < y2; y++ {
		put(c, r.x, y, '│')
		put(c, x2, y, '│')
	}
}

func drawHLine(c *canvas, x, y, w int) {
	if w <= 0 {
		return
	}
	put(c, x, y, '├')
	for i := 1; i < w-1; i++ {
		put(c, x+i, y, '─')
	}
	if w > 1 {
		put(c, x+w-1, y, '┤')
	}
}

func writeCentered(c *canvas, x, y, w int, s string) {
	rs := []rune(s)
	if len(rs) > w {
		rs = rs[:max(0, w)]
	}
	start := x + max(0, (w-len(rs))/2)
	writeRunes(c, start, y, rs)
}

func writeText(c *canvas, x, y int, s string) {
	writeRunes(c, x, y, []rune(s))
}

func writeTextClipped(c *canvas, x, y, w int, s string) {
	rs := []rune(s)
	if len(rs) > w {
		rs = rs[:max(0, w)]
	}
	writeRunes(c, x, y, rs)
}

func writeRight(c *canvas, rightX, y int, s string) {
	rs := []rune(s)
	writeRunes(c, rightX-len(rs)+1, y, rs)
}

func writeRunes(c *canvas, x, y int, rs []rune) {
	if y < 0 || y >= c.h {
		return
	}
	for i, r := range rs {
		put(c, x+i, y, r)
	}
}

func put(c *canvas, x, y int, r rune) {
	if x < 0 || x >= c.w || y < 0 || y >= c.h {
		return
	}
	c.cells[y][x] = r
}

func inset(r rect, n int) rect {
	return rect{r.x + n, r.y + n, max(0, r.w-2*n), max(0, r.h-2*n)}
}

func fixedRect(r rect) rect {
	return rect{r.x, r.y, min(r.w, 18), min(r.h, 10)}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
