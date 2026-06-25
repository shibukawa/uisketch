package renderapi

import (
	"uisketch/layout"
	"uisketch/render/ascii"
	"uisketch/render/svg"
	"uisketch/sketch"
	"uisketch/validate"
)

type Finding struct {
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

type Response struct {
	OK       bool      `json:"ok"`
	Error    string    `json:"error,omitempty"`
	SVG      string    `json:"svg,omitempty"`
	Text     string    `json:"text,omitempty"`
	Findings []Finding `json:"findings,omitempty"`
}

func RenderYAML(source string) Response {
	node, err := layout.ParseYAML(source)
	if err != nil {
		return Response{
			OK:       false,
			Error:    err.Error(),
			Findings: []Finding{{Severity: "error", Message: err.Error()}},
		}
	}
	findings := validate.ValidateLayout(node, nil)
	resp := Response{
		OK:       !validate.HasErrors(findings),
		SVG:      svg.Render(sketch.FromLayout("", node)),
		Text:     ascii.Render(sketch.FromLayout("", node)),
		Findings: fromFindings(findings),
	}
	if !resp.OK {
		resp.Error = "validation failed"
	}
	return resp
}

func fromFindings(findings []validate.Finding) []Finding {
	out := make([]Finding, 0, len(findings))
	for _, finding := range findings {
		out = append(out, Finding{Severity: string(finding.Severity), Message: finding.Message})
	}
	return out
}
