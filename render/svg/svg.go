package svg

import (
	"bytes"
	"fmt"
	"html"
	"math"
	"strings"

	"uisketch/sketch"
)

const (
	defaultWidth  = 960
	defaultHeight = 640
	minWidth      = 48
	minHeight     = 24
	noteBoxWidth  = 180
	noteInnerPad  = 12
	noteLineH     = 16
	noteGap       = 16
)

type Options struct {
	Width  int
	Height int
}

type renderer struct {
	buf          bytes.Buffer
	width        int
	contentWidth int
	height       int
	roughSeq     int
	notes        []noteCallout
}

type rect struct {
	x, y, w, h int
}

type noteCallout struct {
	text   string
	anchor rect
}

type point struct {
	x, y float64
}

type roughOptions struct {
	maxRandomnessOffset float64
	roughness           float64
	bowing              float64
	disableMultiStroke  bool
	preserveVertices    bool
	seed                int64
}

type rougher struct {
	opts roughOptions
	rng  roughRandom
}

type roughRandom struct {
	seed int64
}

func Render(doc *sketch.Document) string {
	return RenderWithOptions(doc, Options{})
}

func RenderWithOptions(doc *sketch.Document, opts Options) string {
	if doc == nil || doc.Root == nil {
		return ""
	}
	w := opts.Width
	if w <= 0 {
		w = defaultWidth
	}
	h := opts.Height
	if h <= 0 {
		h = max(defaultHeight, requiredRootHeight(doc.Root))
	}
	outputW := w
	if noteCount(doc.Root) > 0 {
		outputW += 220
	}
	r := &renderer{width: outputW, contentWidth: w, height: h}
	r.start()
	r.renderRoot(rect{16, 16, r.contentWidth - 32, h - 32}, doc)
	r.renderNotes()
	r.end()
	return r.buf.String()
}

func (r *renderer) start() {
	fmt.Fprintf(&r.buf, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d" role="img">`+"\n", r.width, r.height, r.width, r.height)
	r.buf.WriteString(`<style>
svg{background:#fbfaf7;color:#1f2933;font-family:"Comic Sans MS","Bradley Hand","Segoe Print",cursive,sans-serif}
.line{fill:none;stroke:#1f2933;stroke-width:2.2;stroke-linecap:round;stroke-linejoin:round}
.thin{stroke-width:1.4}
.fill{fill:#fffdf8}
.soft{fill:#f3efe7}
.text{fill:#1f2933;font-size:16px}
.small{font-size:13px;fill:#52606d}
.title{font-size:18px;font-weight:700}
.button{fill:#fdfbf5}
.muted{stroke:#9aa5b1}
.chrome{fill:#f7f7f4}
.accent{fill:#e8f2ff}
.note{fill:#fff2cc;stroke:#d6b656}
.note-connector{fill:none;stroke:#ffd966;stroke-width:2;stroke-linecap:round;stroke-linejoin:round}
.badge-fill{fill:#fffdf8}
.nav{fill:#1f2933;stroke:none}
.nav-stroke{fill:none;stroke:#52606d;stroke-width:3.2;stroke-linecap:round;stroke-linejoin:round}
.fill-only{stroke:none}
</style>
`)
	r.buf.WriteString(`<defs><filter id="wobble"><feTurbulence type="fractalNoise" baseFrequency="0.015" numOctaves="2" seed="8"/><feDisplacementMap in="SourceGraphic" scale="0.7"/></filter></defs>` + "\n")
}

func (r *renderer) end() {
	r.buf.WriteString("</svg>\n")
}

func (r *renderer) renderRoot(bounds rect, doc *sketch.Document) {
	root := doc.Root
	title := rootTitle(doc)
	if root.Type == "mobile" {
		r.renderMobileRoot(bounds, root, title)
		return
	}
	r.box(bounds, "fill")
	switch root.Type {
	case "browser":
		r.renderWindowControls(bounds.x+bounds.w-84, bounds.y+18)
		tabW := min(max(textWidth(title)+70, 180), bounds.w-96)
		r.roundRect(rect{bounds.x + 24, bounds.y + 12, tabW, 36}, 8, "fill", "line thin")
		r.text(bounds.x+72, bounds.y+36, title, "text")
		r.documentIcon(bounds.x+42, bounds.y+20)
		r.line(bounds.x, bounds.y+48, bounds.x+bounds.w, bounds.y+48, "line thin")
		r.renderBrowserNavControls(bounds.x+24, bounds.y+75)
		if root.Address != "" {
			r.roundRect(rect{bounds.x + 120, bounds.y + 58, bounds.w - 140, 34}, 6, "fill", "line thin muted")
			r.documentIcon(bounds.x+134, bounds.y+66)
			r.text(bounds.x+164, bounds.y+80, root.Address, "text")
		}
		r.line(bounds.x, bounds.y+108, bounds.x+bounds.w, bounds.y+108, "line thin")
		r.renderChildrenAsVStack(insetTop(bounds, 120, 16), root.Children)
	case "window":
		r.renderWindowControls(bounds.x+bounds.w-84, bounds.y+24)
		r.text(bounds.x+22, bounds.y+31, title, "text")
		r.line(bounds.x, bounds.y+48, bounds.x+bounds.w, bounds.y+48, "line thin")
		content := insetTop(bounds, 60, 16)
		if len(root.Menu) > 0 && content.h > 48 {
			r.renderMenuBar(rect{bounds.x + 16, bounds.y + 56, bounds.w - 32, 34}, root.Menu)
			content.y += 42
			content.h -= 42
		}
		r.renderChildrenAsVStack(content, root.Children)
	case "dialog":
		r.circle(bounds.x+bounds.w-28, bounds.y+30, 11, "accent", "line")
		r.text(bounds.x+18, bounds.y+31, title, "text")
		r.line(bounds.x, bounds.y+48, bounds.x+bounds.w, bounds.y+48, "line thin")
		r.renderChildrenAsVStack(insetTop(bounds, 60, 16), root.Children)
	case "menu":
		if root.Type == "dialog" {
			r.centeredText(bounds.x, bounds.y+31, bounds.w, title, "text title")
		} else {
			r.text(bounds.x+22, bounds.y+31, title, "text title")
		}
		r.line(bounds.x, bounds.y+48, bounds.x+bounds.w, bounds.y+48, "line thin")
		r.renderChildrenAsVStack(insetTop(bounds, 60, 16), root.Children)
	default:
		r.renderNode(inset(bounds, 16), root)
	}
}

func (r *renderer) renderMobileRoot(bounds rect, root *sketch.Node, title string) {
	phoneH := bounds.h
	phoneW := min(bounds.w, max(220, int(float64(phoneH)*0.48)))
	phone := rect{bounds.x + max(0, (bounds.w-phoneW)/2), bounds.y, phoneW, phoneH}
	r.roundRect(phone, 34, "fill", "line")
	notchW := min(96, max(58, phone.w/3))
	fmt.Fprintf(&r.buf, `<rect x="%d" y="%d" width="%d" height="%d" rx="%d" class="nav fill-only"/>`+"\n", phone.x+(phone.w-notchW)/2, phone.y+8, notchW, 20, 10)
	r.line(phone.x-6, phone.y+80, phone.x-6, phone.y+126, "line thin muted")
	r.line(phone.x+phone.w+6, phone.y+104, phone.x+phone.w+6, phone.y+174, "line thin muted")
	fmt.Fprintf(&r.buf, `<rect x="%d" y="%d" width="%d" height="%d" rx="%d" class="nav fill-only"/>`+"\n", phone.x+phone.w/2-48, phone.y+phone.h-28, 96, 5, 3)
	contentTop := 62
	if title != "" {
		r.centeredText(phone.x+18, phone.y+54, phone.w-36, title, "text title")
		contentTop = 76
	}
	content := rect{phone.x + 22, phone.y + contentTop, phone.w - 44, max(0, phone.h-contentTop-46)}
	if len(root.Menu) > 0 && content.h > 64 {
		menu := rect{phone.x + 22, phone.y + phone.h - 70, phone.w - 44, 38}
		r.renderMenuBar(menu, root.Menu)
		content.h = max(0, menu.y-content.y-12)
	}
	r.renderChildrenAsVStack(content, root.Children)
}

func (r *renderer) renderWindowControls(x, y int) {
	r.circle(x, y, 10, "fill", "line thin")
	r.circle(x+28, y, 10, "fill", "line thin")
	r.circle(x+56, y, 10, "accent", "line thin")
}

func (r *renderer) renderBrowserNavControls(x, centerY int) {
	// Draw browser navigation as vector shapes instead of font glyphs so sizing
	// remains stable across renderers and browsers.
	arrowX := x - 14
	r.solidPolygon([]point{
		{float64(arrowX + 2), float64(centerY)},
		{float64(arrowX + 20), float64(centerY - 12)},
		{float64(arrowX + 20), float64(centerY + 12)},
	}, "nav")
	r.solidPolygon([]point{
		{float64(arrowX + 36), float64(centerY - 12)},
		{float64(arrowX + 54), float64(centerY)},
		{float64(arrowX + 36), float64(centerY + 12)},
	}, "nav")
	r.reloadIcon(x+70, centerY+12, 12)
}

func (r *renderer) renderMenuBar(bounds rect, items []string) {
	r.roundRect(bounds, 4, "chrome", "line thin muted")
	if len(items) == 0 {
		return
	}
	x := bounds.x + 14
	for _, item := range items {
		r.text(x, bounds.y+23, item, "text small")
		x += textWidth(item) + 28
		if x > bounds.x+bounds.w-20 {
			break
		}
	}
}

func (r *renderer) reloadIcon(cx, cy, radius int) {
	endX := cx + radius - 2
	endY := cy - radius + 2
	fmt.Fprintf(&r.buf, `<path d="M %d %d A %d %d 0 1 1 %d %d" class="nav-stroke"/>`+"\n", cx-radius+2, cy, radius, radius, endX, endY)
	r.solidPolygon([]point{
		{float64(endX), float64(endY + 7)},
		{float64(endX - 8), float64(endY - 1)},
		{float64(endX + 4), float64(endY - 3)},
	}, "nav")
}

func (r *renderer) documentIcon(x, y int) {
	r.roundRect(rect{x, y, 18, 22}, 2, "fill", "line thin muted")
	r.line(x+12, y, x+18, y+6, "line thin muted")
}

func (r *renderer) renderNode(bounds rect, n *sketch.Node) {
	if n == nil || bounds.w <= 0 || bounds.h <= 0 {
		return
	}
	switch n.Type {
	case "spacer":
		return
	case "vstack":
		r.renderChildrenAsVStackWithHeights(bounds, n.Children, n.Heights)
	case "hstack":
		r.renderChildrenAsHStack(bounds, n.Children, n.Widths)
	case "grid":
		r.renderGrid(bounds, n)
	case "splitter":
		r.renderSplitter(bounds, n)
	case "tabs":
		r.renderTabs(bounds, n)
	case "section":
		r.box(bounds, "fill")
		label := labelFor(n)
		labelW := min(bounds.w-24, textWidth(label)+28)
		if labelW > 0 {
			fmt.Fprintf(&r.buf, `<rect x="%d" y="%d" width="%d" height="%d" class="fill fill-only"/>`+"\n", bounds.x+14, bounds.y-13, labelW, 24)
			r.text(bounds.x+22, bounds.y+6, label, "text")
		}
		r.renderChildrenAsVStack(inset(bounds, 18), n.Children)
	case "button":
		r.renderButton(bounds, n)
	case "label":
		r.text(bounds.x, bounds.y+20, labelFor(n), "text")
	case "input":
		r.renderInput(bounds, n)
	case "textarea":
		r.renderTextarea(bounds, n)
	case "combobox":
		r.renderCombobox(bounds, n)
	case "checkbox":
		r.square(bounds.x, bounds.y+4, 18)
		r.text(bounds.x+28, bounds.y+20, labelFor(n), "text")
	case "radio":
		r.circle(bounds.x+9, bounds.y+13, 9, "fill", "line")
		r.text(bounds.x+28, bounds.y+20, labelFor(n), "text")
	case "toggle":
		r.roundRect(rect{bounds.x, bounds.y + 4, 48, 24}, 12, "soft", "line")
		r.circle(bounds.x+14, bounds.y+16, 8, "fill", "line thin")
		r.text(bounds.x+60, bounds.y+21, labelFor(n), "text")
	case "slider":
		r.text(bounds.x, bounds.y+18, labelFor(n), "text")
		y := bounds.y + 42
		r.line(bounds.x, y, bounds.x+bounds.w-20, y, "line muted")
		r.circle(bounds.x+bounds.w/2, y, 9, "fill", "line")
	case "table":
		r.renderTable(bounds, n)
	case "list":
		r.renderList(bounds, n, "List")
	case "tree":
		r.renderList(bounds, n, "Tree")
	case "calendar":
		r.renderList(bounds, n, "Calendar")
	case "badge":
		r.roundRect(rect{bounds.x, bounds.y + 4, max(54, textWidth(labelFor(n))+22), 28}, 14, "soft", "line thin")
		r.centeredText(bounds.x, bounds.y+23, max(54, textWidth(labelFor(n))+22), labelFor(n), "text small")
	case "image", "custom":
		r.renderImagePlaceholder(bounds, n)
	default:
		r.box(bounds, "fill")
		r.centeredText(bounds.x, bounds.y+bounds.h/2+5, bounds.w, labelFor(n), "text")
	}
	if n.Note != "" {
		r.notes = append(r.notes, noteCallout{text: n.Note, anchor: bounds})
	}
}

func (r *renderer) renderChildrenAsVStack(bounds rect, children []*sketch.Node) {
	r.renderChildrenAsVStackWithHeights(bounds, children, nil)
}

func (r *renderer) renderChildrenAsVStackWithHeights(bounds rect, children []*sketch.Node, slots []sketch.SizeSlot) {
	if len(children) == 0 || bounds.w <= 0 || bounds.h <= 0 {
		return
	}
	heights := distributeVertical(bounds.h, children, slots)
	y := bounds.y
	for i, child := range children {
		h := heights[i]
		r.renderNode(rect{bounds.x, y, bounds.w, h}, child)
		y += h
	}
}

func (r *renderer) renderChildrenAsHStack(bounds rect, children []*sketch.Node, slots []sketch.SizeSlot) {
	r.renderChildrenAsHStackWithMode(bounds, children, slots, false)
}

func (r *renderer) renderChildrenAsHStackFill(bounds rect, children []*sketch.Node, slots []sketch.SizeSlot) {
	r.renderChildrenAsHStackWithMode(bounds, children, slots, true)
}

func (r *renderer) renderChildrenAsHStackWithMode(bounds rect, children []*sketch.Node, slots []sketch.SizeSlot, fillRemainder bool) {
	if len(children) == 0 || bounds.w <= 0 || bounds.h <= 0 {
		return
	}
	gap := 14
	if len(children) == 1 {
		gap = 0
	}
	widths := distributeHorizontal(max(1, bounds.w-gap*(len(children)-1)), children, slots, fillRemainder)
	x := bounds.x
	for i, child := range children {
		w := widths[i]
		r.renderNode(rect{x, bounds.y, w, bounds.h}, child)
		x += w + gap
	}
}

func (r *renderer) renderSingleChild(bounds rect, n *sketch.Node) {
	if len(n.Children) > 0 {
		r.renderNode(bounds, n.Children[0])
	}
}

func (r *renderer) renderGrid(bounds rect, n *sketch.Node) {
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
	cellW := max(1, bounds.w/cols)
	cellH := max(1, bounds.h/rows)
	for i, child := range children {
		r.renderNode(rect{bounds.x + (i%cols)*cellW, bounds.y + (i/cols)*cellH, cellW - 8, cellH - 8}, child)
	}
}

func (r *renderer) renderSplitter(bounds rect, n *sketch.Node) {
	if n.Orientation == "vertical" {
		r.renderChildrenAsVStackWithHeights(bounds, n.Children, defaultSizes(n.Sizes))
		return
	}
	r.renderChildrenAsHStackFill(bounds, n.Children, defaultSizes(n.Sizes))
}

func defaultSizes(slots []sketch.SizeSlot) []sketch.SizeSlot {
	if len(slots) > 0 {
		return slots
	}
	return []sketch.SizeSlot{{Percent: 25}, {Percent: 75}}
}

func (r *renderer) renderTabs(bounds rect, n *sketch.Node) {
	x := bounds.x
	labels := tabLabels(n)
	active := activeTabIndex(labels)
	var activeBounds rect
	for i, label := range labels {
		text := label.Text
		class := "text"
		if i == active {
			class = "text title"
		}
		w := min(max(textWidth(text)+30, 96), max(96, bounds.w/3))
		r.roundRect(rect{x, bounds.y, w, 38}, 8, "button", "line")
		r.centeredText(x, bounds.y+25, w, text, class)
		if i == active {
			activeBounds = rect{x, bounds.y, w, 38}
		}
		x += w - 2
		if x >= bounds.x+bounds.w {
			break
		}
	}
	panel := rect{bounds.x, bounds.y + 37, bounds.w, bounds.h - 37}
	r.box(panel, "fill")
	if activeBounds.w > 0 {
		fmt.Fprintf(&r.buf, `<rect x="%d" y="%d" width="%d" height="%d" class="button fill-only"/>`+"\n", activeBounds.x+2, panel.y-2, activeBounds.w-4, 6)
	}
	r.renderActiveTabBody(inset(panel, 18), n, active)
}

func (r *renderer) renderButton(bounds rect, n *sketch.Node) {
	label := labelFor(n)
	w := min(max(textWidth(label)+34, minWidth), bounds.w)
	if n.Badge != "" {
		w = min(max(w, textWidth(label)+textWidth(n.Badge)+54), bounds.w)
	}
	h := min(40, max(minHeight, bounds.h))
	box := rect{bounds.x + max(0, bounds.w-w), bounds.y + max(0, bounds.h-h)/2, w, h}
	r.roundRect(box, 10, "button", "line")
	r.centeredText(box.x, box.y+25, box.w, label, "text")
	if n.Badge != "" {
		r.renderBadgeMarker(box.x+box.w-9, box.y+1, n.Badge)
	}
}

func (r *renderer) renderInput(bounds rect, n *sketch.Node) {
	r.text(bounds.x, bounds.y+18, labelFor(n), "text")
	r.roundRect(rect{bounds.x, bounds.y + 28, min(bounds.w, max(160, textWidth(labelFor(n))+80)), 38}, 8, "fill", "line")
}

func (r *renderer) renderTextarea(bounds rect, n *sketch.Node) {
	r.text(bounds.x, bounds.y+18, labelFor(n), "text")
	h := min(max(72, bounds.h-30), max(72, bounds.h))
	r.roundRect(rect{bounds.x, bounds.y + 28, min(bounds.w, max(180, textWidth(labelFor(n))+80)), h}, 8, "fill", "line")
}

func (r *renderer) renderCombobox(bounds rect, n *sketch.Node) {
	r.text(bounds.x, bounds.y+18, labelFor(n), "text")
	box := rect{bounds.x, bounds.y + 28, min(bounds.w, max(160, textWidth(labelFor(n))+80)), 38}
	r.roundRect(box, 8, "fill", "line")
	r.solidPolygon([]point{
		{float64(box.x + box.w - 28), float64(box.y + 16)},
		{float64(box.x + box.w - 16), float64(box.y + 16)},
		{float64(box.x + box.w - 22), float64(box.y + 24)},
	}, "nav")
}

func (r *renderer) renderTable(bounds rect, n *sketch.Node) {
	r.box(bounds, "fill")
	if len(n.Columns) == 0 {
		r.text(bounds.x+18, bounds.y+28, labelFor(n), "text")
		return
	}
	colW := max(1, bounds.w/len(n.Columns))
	for i, column := range n.Columns {
		cellX := bounds.x + i*colW
		if i > 0 {
			r.line(cellX, bounds.y, cellX, bounds.y+bounds.h, "line thin")
		}
		r.text(cellX+18, bounds.y+28, column, "text")
	}
	r.line(bounds.x, bounds.y+42, bounds.x+bounds.w, bounds.y+42, "line thin")
	for y := bounds.y + 76; y < bounds.y+bounds.h-16; y += 36 {
		r.line(bounds.x+12, y, bounds.x+bounds.w-12, y+roughOffset(y), "line thin muted")
	}
}

func (r *renderer) renderList(bounds rect, n *sketch.Node, kind string) {
	r.box(bounds, "fill")
	r.text(bounds.x+18, bounds.y+28, kind+": "+labelFor(n), "text")
	for y := bounds.y + 58; y < bounds.y+bounds.h-16; y += 30 {
		r.circle(bounds.x+24, y-4, 3, "soft", "line thin")
		r.line(bounds.x+38, y-4, bounds.x+bounds.w-24, y+roughOffset(y), "line thin muted")
	}
}

func (r *renderer) renderImagePlaceholder(bounds rect, n *sketch.Node) {
	r.box(bounds, "soft")
	r.line(bounds.x+12, bounds.y+12, bounds.x+bounds.w-12, bounds.y+bounds.h-12, "line thin muted")
	r.line(bounds.x+bounds.w-12, bounds.y+12, bounds.x+12, bounds.y+bounds.h-12, "line thin muted")
	r.centeredText(bounds.x, bounds.y+bounds.h/2+5, bounds.w, labelFor(n), "text")
}

func (r *renderer) renderBadgeMarker(cx, cy int, text string) {
	r.circle(cx, cy, 12, "badge-fill", "line")
	r.centeredText(cx-11, cy+5, 22, strings.TrimSpace(text), "text small")
}

func (r *renderer) renderNotes() {
	if len(r.notes) == 0 {
		return
	}
	x := r.contentWidth + 20
	y := 24
	w := min(noteBoxWidth, r.width-x-16)
	for _, note := range r.notes {
		lines := wrapText(note.text, max(1, w-noteInnerPad*2))
		h := noteBoxHeight(lines)
		box := rect{x, y, w, h}
		anchorX := note.anchor.x + note.anchor.w
		anchorY := note.anchor.y + max(0, min(note.anchor.h/2, 28))
		r.line(anchorX, anchorY, box.x, box.y+box.h/2, "note-connector")
		fmt.Fprintf(&r.buf, `<rect x="%d" y="%d" width="%d" height="%d" rx="6" class="note"/>`+"\n", box.x, box.y, box.w, box.h)
		r.multilineText(box.x+noteInnerPad, box.y+22, box.w-noteInnerPad*2, lines, "text small")
		y += h + noteGap
	}
}

func (r *renderer) box(bounds rect, fillClass string) {
	r.roundRect(bounds, 6, fillClass, "line")
}

func (r *renderer) roundRect(bounds rect, radius int, fillClass, strokeClass string) {
	if bounds.w <= 0 || bounds.h <= 0 {
		return
	}
	fmt.Fprintf(&r.buf, `<rect x="%d" y="%d" width="%d" height="%d" rx="%d" class="%s fill-only"/>`+"\n", bounds.x, bounds.y, bounds.w, bounds.h, radius, fillClass)
	r.roughPath(roughPolygon([]point{
		{float64(bounds.x), float64(bounds.y)},
		{float64(bounds.x + bounds.w), float64(bounds.y)},
		{float64(bounds.x + bounds.w), float64(bounds.y + bounds.h)},
		{float64(bounds.x), float64(bounds.y + bounds.h)},
	}, r.nextRoughOptions()), strokeClass)
}

func (r *renderer) square(x, y, size int) {
	r.roundRect(rect{x, y, size, size}, 3, "fill", "line")
}

func (r *renderer) circle(x, y, radius int, fillClass, strokeClass string) {
	fmt.Fprintf(&r.buf, `<circle cx="%d" cy="%d" r="%d" class="%s fill-only"/>`+"\n", x, y, radius, fillClass)
	r.roughPath(roughEllipse(float64(x), float64(y), float64(radius), r.nextRoughOptions()), strokeClass)
}

func (r *renderer) line(x1, y1, x2, y2 int, class string) {
	r.roughPath(roughLine(float64(x1), float64(y1), float64(x2), float64(y2), r.nextRoughOptions()), class)
}

func (r *renderer) roughPath(d, class string) {
	if d == "" {
		return
	}
	fmt.Fprintf(&r.buf, `<path d="%s" class="%s"/>`+"\n", d, class)
}

func (r *renderer) solidPolygon(points []point, class string) {
	var b strings.Builder
	for i, p := range points {
		if i > 0 {
			b.WriteByte(' ')
		}
		fmt.Fprintf(&b, "%s,%s", formatFloat(p.x), formatFloat(p.y))
	}
	fmt.Fprintf(&r.buf, `<polygon points="%s" class="%s"/>`+"\n", b.String(), class)
}

func (r *renderer) nextRoughOptions() roughOptions {
	r.roughSeq++
	return roughOptions{
		maxRandomnessOffset: 1.8,
		roughness:           1,
		bowing:              1,
		seed:                int64(983 + r.roughSeq*7919),
	}
}

func (r *renderer) text(x, y int, value, class string) {
	if strings.TrimSpace(value) == "" {
		return
	}
	fmt.Fprintf(&r.buf, `<text x="%d" y="%d" class="%s">%s</text>`+"\n", x, y, class, html.EscapeString(value))
}

func (r *renderer) centeredText(x, y, width int, value, class string) {
	r.text(x+max(0, (width-textWidth(value))/2), y, value, class)
}

func (r *renderer) multilineText(x, y, width int, lines []string, class string) {
	for i, line := range lines {
		r.text(x+max(0, (width-textWidth(line))/2), y+i*noteLineH, line, class)
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
			return textWidth(labelFor(n)) + textWidth(n.Badge) + 60
		}
		return textWidth(labelFor(n)) + 44
	case "checkbox", "radio", "toggle", "slider", "label", "badge":
		return textWidth(labelFor(n)) + 90
	case "input", "textarea", "combobox":
		return textWidth(labelFor(n)) + 120
	default:
		return 260
	}
}

func intrinsicHeight(n *sketch.Node) int {
	switch n.Type {
	case "spacer":
		return 1
	case "button":
		return 48
	case "input", "combobox":
		return 82
	case "textarea":
		return 128
	case "hstack":
		return max(52, maxChildHeight(n.Children))
	case "vstack":
		return childrenVStackHeight(n.Children)
	case "label", "checkbox", "radio", "toggle", "badge":
		return 34
	case "slider":
		return 64
	case "table":
		return 72
	case "tabs":
		return max(240, activeTabHeight(n)+72)
	case "grid":
		return gridHeight(n)
	case "section":
		if allFlexibleDisplay(n.Children) {
			return max(120, childrenVStackHeight(n.Children)+36)
		}
		return max(180, childrenVStackHeight(n.Children)+36)
	case "list", "tree", "calendar", "image", "custom":
		return 72
	case "splitter":
		if n.Orientation == "vertical" {
			return max(240, childrenVStackHeight(n.Children))
		}
		return max(240, maxChildHeight(n.Children))
	default:
		return 96
	}
}

func allFlexibleDisplay(children []*sketch.Node) bool {
	if len(children) == 0 {
		return false
	}
	for _, child := range children {
		if !isFlexibleDisplay(child) {
			return false
		}
	}
	return true
}

func isFlexibleDisplay(n *sketch.Node) bool {
	if n == nil {
		return true
	}
	switch n.Type {
	case "spacer", "table", "list", "tree", "calendar", "image", "custom":
		return true
	case "grid", "vstack", "hstack", "splitter":
		return allFlexibleDisplay(n.Children)
	default:
		return false
	}
}

func requiredRootHeight(root *sketch.Node) int {
	content := childrenVStackHeight(root.Children)
	notes := requiredNotesHeight(root)
	var body int
	switch root.Type {
	case "browser":
		body = content + 168
	case "window":
		menu := 0
		if len(root.Menu) > 0 {
			menu = 42
		}
		body = content + menu + 92
	case "dialog", "menu":
		body = content + 92
	case "mobile":
		menu := 0
		if len(root.Menu) > 0 {
			menu = 64
		}
		body = content + menu + 132
	default:
		body = intrinsicHeight(root) + 64
	}
	return max(body, notes)
}

func requiredNotesHeight(root *sketch.Node) int {
	heights := noteHeights(root)
	if len(heights) == 0 {
		return 0
	}
	total := 48
	for i, h := range heights {
		total += h
		if i < len(heights)-1 {
			total += noteGap
		}
	}
	return total
}

func noteHeights(n *sketch.Node) []int {
	if n == nil {
		return nil
	}
	var out []int
	if n.Note != "" {
		out = append(out, noteBoxHeight(wrapText(n.Note, noteBoxWidth-noteInnerPad*2)))
	}
	for _, child := range n.Children {
		out = append(out, noteHeights(child)...)
	}
	return out
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
		return 72
	}
	cols := n.GridColumns
	if cols <= 0 {
		cols = 2
	}
	rows := (len(n.Children) + cols - 1) / cols
	if allFlexibleDisplay(n.Children) {
		return max(72, rows*max(1, maxChildHeight(n.Children)))
	}
	return max(180, rows*(max(1, maxChildHeight(n.Children))+8))
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

func noteCount(n *sketch.Node) int {
	if n == nil {
		return 0
	}
	count := 0
	if n.Note != "" {
		count++
	}
	for _, child := range n.Children {
		count += noteCount(child)
	}
	return count
}

func fixedRect(bounds rect) rect {
	return rect{bounds.x, bounds.y, min(bounds.w, 220), min(bounds.h, 160)}
}

func inset(bounds rect, amount int) rect {
	return rect{bounds.x + amount, bounds.y + amount, max(0, bounds.w-amount*2), max(0, bounds.h-amount*2)}
}

func insetTop(bounds rect, top, side int) rect {
	return rect{bounds.x + side, bounds.y + top, max(0, bounds.w-side*2), max(0, bounds.h-top-side)}
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

func activeTabIndex(labels []sketch.TabLabel) int {
	for i, label := range labels {
		if label.Selected {
			return i
		}
	}
	return 0
}

func (r *renderer) renderActiveTabBody(bounds rect, n *sketch.Node, active int) {
	if len(n.Labels) == 0 && len(n.Children) > active {
		r.renderNode(bounds, n.Children[active])
		return
	}
	if len(n.Children) == 1 {
		r.renderNode(bounds, n.Children[0])
		return
	}
	r.renderChildrenAsVStack(bounds, n.Children)
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

func textWidth(value string) int {
	return int(math.Ceil(float64(len([]rune(value))) * 8.5))
}

func noteBoxHeight(lines []string) int {
	return max(56, len(lines)*noteLineH+24)
}

func wrapText(value string, maxWidth int) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return []string{""}
	}
	var lines []string
	for _, paragraph := range strings.Split(value, "\n") {
		paragraph = strings.TrimSpace(paragraph)
		if paragraph == "" {
			lines = append(lines, "")
			continue
		}
		lines = append(lines, wrapParagraph(paragraph, maxWidth)...)
	}
	return lines
}

func wrapParagraph(value string, maxWidth int) []string {
	words := strings.Fields(value)
	if len(words) <= 1 {
		return splitRunesToWidth(value, maxWidth)
	}
	var lines []string
	current := ""
	for _, word := range words {
		if current == "" {
			if textWidth(word) > maxWidth {
				chunks := splitRunesToWidth(word, maxWidth)
				lines = append(lines, chunks[:max(0, len(chunks)-1)]...)
				current = chunks[len(chunks)-1]
			} else {
				current = word
			}
			continue
		}
		next := current + " " + word
		if textWidth(next) <= maxWidth {
			current = next
			continue
		}
		lines = append(lines, current)
		if textWidth(word) > maxWidth {
			chunks := splitRunesToWidth(word, maxWidth)
			lines = append(lines, chunks[:max(0, len(chunks)-1)]...)
			current = chunks[len(chunks)-1]
		} else {
			current = word
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func splitRunesToWidth(value string, maxWidth int) []string {
	rs := []rune(value)
	if len(rs) == 0 {
		return []string{""}
	}
	var lines []string
	var current []rune
	for _, r := range rs {
		next := append(append([]rune{}, current...), r)
		if len(current) > 0 && textWidth(string(next)) > maxWidth {
			lines = append(lines, string(current))
			current = []rune{r}
			continue
		}
		current = next
	}
	if len(current) > 0 {
		lines = append(lines, string(current))
	}
	return lines
}

func roughOffset(seed int) int {
	return seed%5 - 2
}

func roughPolygon(points []point, opts roughOptions) string {
	if len(points) < 2 {
		return ""
	}
	rr := newRougher(opts)
	var ops []roughOp
	for i := 0; i < len(points); i++ {
		next := (i + 1) % len(points)
		ops = append(ops, rr.doubleLine(points[i].x, points[i].y, points[next].x, points[next].y)...)
	}
	return opsToPath(ops)
}

func roughEllipse(cx, cy, radius float64, opts roughOptions) string {
	steps := 16
	points := make([]point, 0, steps)
	for i := 0; i < steps; i++ {
		angle := (math.Pi * 2 * float64(i)) / float64(steps)
		points = append(points, point{
			x: cx + math.Cos(angle)*radius,
			y: cy + math.Sin(angle)*radius,
		})
	}
	return roughPolygon(points, opts)
}

func roughLine(x1, y1, x2, y2 float64, opts roughOptions) string {
	return opsToPath(newRougher(opts).doubleLine(x1, y1, x2, y2))
}

func newRougher(opts roughOptions) *rougher {
	if opts.maxRandomnessOffset == 0 {
		opts.maxRandomnessOffset = 1
	}
	if opts.roughness == 0 {
		opts.roughness = 1
	}
	if opts.seed == 0 {
		opts.seed = 1
	}
	return &rougher{opts: opts, rng: roughRandom{seed: opts.seed}}
}

type roughOp struct {
	kind string
	data []float64
}

func (r *rougher) doubleLine(x1, y1, x2, y2 float64) []roughOp {
	ops := r.line(x1, y1, x2, y2, true, false)
	if r.opts.disableMultiStroke {
		return ops
	}
	return append(ops, r.line(x1, y1, x2, y2, true, true)...)
}

func (r *rougher) line(x1, y1, x2, y2 float64, move, overlay bool) []roughOp {
	lengthSq := math.Pow(x1-x2, 2) + math.Pow(y1-y2, 2)
	length := math.Sqrt(lengthSq)
	roughnessGain := 1.0
	if length > 500 {
		roughnessGain = 0.4
	} else if length >= 200 {
		roughnessGain = -0.0016668*length + 1.233334
	}

	offset := r.opts.maxRandomnessOffset
	if offset*offset*100 > lengthSq {
		offset = length / 10
	}
	halfOffset := offset / 2
	divergePoint := 0.2 + r.random()*0.2
	midDispX := r.opts.bowing * r.opts.maxRandomnessOffset * (y2 - y1) / 200
	midDispY := r.opts.bowing * r.opts.maxRandomnessOffset * (x1 - x2) / 200
	midDispX = r.offsetOpt(midDispX, roughnessGain)
	midDispY = r.offsetOpt(midDispY, roughnessGain)

	var ops []roughOp
	randomHalf := func() float64 { return r.offsetOpt(halfOffset, roughnessGain) }
	randomFull := func() float64 { return r.offsetOpt(offset, roughnessGain) }
	preserve := r.opts.preserveVertices
	if move {
		if overlay {
			ops = append(ops, roughOp{kind: "move", data: []float64{
				x1 + chooseOffset(preserve, randomHalf),
				y1 + chooseOffset(preserve, randomHalf),
			}})
		} else {
			ops = append(ops, roughOp{kind: "move", data: []float64{
				x1 + chooseOffset(preserve, randomFull),
				y1 + chooseOffset(preserve, randomFull),
			}})
		}
	}
	offsetFn := randomFull
	if overlay {
		offsetFn = randomHalf
	}
	ops = append(ops, roughOp{kind: "bcurveTo", data: []float64{
		midDispX + x1 + (x2-x1)*divergePoint + offsetFn(),
		midDispY + y1 + (y2-y1)*divergePoint + offsetFn(),
		midDispX + x1 + 2*(x2-x1)*divergePoint + offsetFn(),
		midDispY + y1 + 2*(y2-y1)*divergePoint + offsetFn(),
		x2 + chooseOffset(preserve, offsetFn),
		y2 + chooseOffset(preserve, offsetFn),
	}})
	return ops
}

func chooseOffset(preserve bool, fn func() float64) float64 {
	if preserve {
		return 0
	}
	return fn()
}

func (r *rougher) random() float64 {
	return r.rng.next()
}

func (r *rougher) offset(min, max, roughnessGain float64) float64 {
	return r.opts.roughness * roughnessGain * ((r.random() * (max - min)) + min)
}

func (r *rougher) offsetOpt(x, roughnessGain float64) float64 {
	return r.offset(-x, x, roughnessGain)
}

func (r *roughRandom) next() float64 {
	r.seed = (48271 * r.seed) & (1<<31 - 1)
	return float64(r.seed) / float64(1<<31)
}

func opsToPath(ops []roughOp) string {
	var b strings.Builder
	for _, op := range ops {
		switch op.kind {
		case "move":
			fmt.Fprintf(&b, "M%s %s ", formatFloat(op.data[0]), formatFloat(op.data[1]))
		case "bcurveTo":
			fmt.Fprintf(&b, "C%s %s, %s %s, %s %s ", formatFloat(op.data[0]), formatFloat(op.data[1]), formatFloat(op.data[2]), formatFloat(op.data[3]), formatFloat(op.data[4]), formatFloat(op.data[5]))
		}
	}
	return strings.TrimSpace(b.String())
}

func formatFloat(v float64) string {
	if math.Abs(v) < 0.0001 {
		v = 0
	}
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", v), "0"), ".")
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
