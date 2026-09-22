package renderer

import (
	"regexp"
	"strings"
	"testing"

	"github.com/florent/status-line/internal/domain/model"
)

// stripSGR removes the colour escapes, leaving what the terminal prints.
func stripSGR(s string) string {
	return regexp.MustCompile("\033\\[[0-9;]*m").ReplaceAllString(s, "")
}

// servers builds on enabled and off disabled servers, from mixed scopes.
func servers(on, off int) model.MCPServers {
	out := make(model.MCPServers, 0, on+off)
	scopes := []model.MCPSource{model.MCPSourceCLI, model.MCPSourceUser, model.MCPSourcePlugin}
	for i := range on {
		out = append(out, model.MCPServer{Name: "on" + itoa(i), Enabled: true, Source: scopes[i%len(scopes)]})
	}
	for i := range off {
		out = append(out, model.MCPServer{Name: "off" + itoa(i), Source: model.MCPSourceUser})
	}
	return out
}

func TestSummarizeMCP(t *testing.T) {
	list := servers(3, 2)
	if got := summarizeMCP(list); got != (mcpSummary{on: 3, off: 2}) {
		t.Errorf("summarizeMCP = %+v, want 3 on, 2 off, idle", got)
	}
	list[4].Busy = true
	if got := summarizeMCP(list); !got.busy {
		t.Error("a disabled server being called still lights the pill")
	}
	ghost := servers(1, 0).WithBusy([]string{"ghost"})
	if got := summarizeMCP(ghost); got != (mcpSummary{on: 2, busy: true}) {
		t.Errorf("an undeclared server being called counts as enabled, got %+v", got)
	}
}

func TestPowerline_renderMCPPill(t *testing.T) {
	head := " " + LeftRound + " " + glyphs.MCP + " "
	tests := []struct {
		name    string
		servers model.MCPServers
		want    string
	}{
		{name: "empty draws nothing", servers: model.MCPServers{}, want: ""},
		{name: "the total of enabled servers", servers: servers(7, 0), want: head + "7 " + RightRound},
		{name: "a discreet disabled suffix", servers: servers(7, 1), want: head + "7 ·1 " + RightRound},
		{name: "only disabled", servers: servers(0, 2), want: head + "0 ·2 " + RightRound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sb strings.Builder
			(&Powerline{}).renderMCPPill(&sb, tt.servers)
			if got := stripSGR(sb.String()); got != tt.want {
				t.Errorf("pill = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPowerline_renderMCPPillStyling(t *testing.T) {
	p := &Powerline{}
	idle := p.mcpPill(servers(2, 1))
	if !strings.HasPrefix(idle, " "+FgMCPEnabled+LeftRound+Reset+BgMCPEnabled+FgMCPEnabledText+Bold+" "+glyphs.MCP+" 2") {
		t.Errorf("at rest: bold dark ink on pale teal, got %q", idle)
	}
	if !strings.Contains(idle, BgMCPEnabled+FgMCPMuted+" "+mcpOffMark+StrikeMCP+"1"+Reset) {
		t.Errorf("at rest: the disabled count is muted and crossed out, got %q", idle)
	}
	if !strings.HasSuffix(idle, BgMCPEnabled+" "+Reset+FgMCPEnabled+RightRound+Reset) {
		t.Errorf("at rest: pale teal caps, got %q", idle)
	}

	list := servers(2, 1)
	list[0].Busy = true
	lit := p.mcpPill(list)
	if !strings.HasPrefix(lit, " "+FgMCPEnabledText+LeftRound+Reset+BgMCPLabel+FgWhite+Bold+" "+glyphs.MCP+" 2") {
		t.Errorf("lit: the whole pill is bold white on dark teal, got %q", lit)
	}
	if !strings.Contains(lit, BgMCPLabel+FgMCPEnabled+" "+mcpOffMark+StrikeMCP+"1"+Reset) {
		t.Errorf("lit: the disabled count stays, pale teal on dark teal, got %q", lit)
	}
	if strings.Contains(lit, BgMCPEnabled) {
		t.Errorf("lit: no pale teal ground is left, got %q", lit)
	}
	if strings.Count(lit, LeftRound) != 1 {
		t.Errorf("lit: one pill, no chip inside it, got %q", lit)
	}
}

func TestPowerline_renderMCPPillTextGlyphs(t *testing.T) {
	saved := glyphs
	t.Cleanup(func() { glyphs = saved })
	glyphs = textGlyphs
	if got := stripSGR((&Powerline{}).mcpPill(servers(7, 0))); !strings.Contains(got, " MCP 7 ") {
		t.Errorf("text glyphs: want \"MCP 7\", got %q", got)
	}
}

// withMCPLine sets the pill's line for one test.
func withMCPLine(t *testing.T, line2 bool) {
	t.Helper()
	saved := mcpOnLine2
	t.Cleanup(func() { mcpOnLine2 = saved })
	mcpOnLine2 = line2
}

// lightLine is a session narrow enough to carry the pill at 80 columns.
func lightLine(width int) model.StatusLineData {
	data := busyLine(width)
	data.Limits.Scoped = nil
	data.Git.Branch = "main"
	data.MCP = servers(7, 1)
	return data
}

func TestMCPPillClosesLineOne(t *testing.T) {
	withMCPLine(t, false)
	for _, width := range []int{0, 200, 120, 80} {
		out := (&Powerline{}).Render(lightLine(width))
		line1, line2, _ := strings.Cut(out, "\n")
		if !strings.HasSuffix(stripSGR(line1), glyphs.MCP+" 7 ·1 "+RightRound) {
			t.Errorf("COLUMNS=%d: the pill closes line one, got %q", width, stripSGR(line1))
		}
		if strings.Contains(line2, glyphs.MCP) {
			t.Errorf("COLUMNS=%d: the pill is drawn once, got line two %q", width, stripSGR(line2))
		}
		if width > 0 && VisibleWidth(line1) > lineBudget(width) {
			t.Errorf("COLUMNS=%d: line one is %d wide", width, VisibleWidth(line1))
		}
	}
}

func TestMCPPillMovesToLineTwoWhenLineOneIsFull(t *testing.T) {
	withMCPLine(t, false)
	data := busyLine(80)
	data.MCP = servers(7, 1)
	out := (&Powerline{}).Render(data)
	line1, line2, _ := strings.Cut(out, "\n")
	if strings.Contains(line1, glyphs.MCP) {
		t.Errorf("a full line one gives the pill up, got %q", stripSGR(line1))
	}
	if VisibleWidth(line1) > lineBudget(80) {
		t.Errorf("line one still fits without the pill, got %d cells", VisibleWidth(line1))
	}
	if !strings.Contains(stripSGR(line2), glyphs.MCP+" 7 ·1") {
		t.Errorf("the pill moved to line two, got %q", stripSGR(line2))
	}
}

func TestMCPPillEnvKeepsItOnLineTwo(t *testing.T) {
	withMCPLine(t, true)
	out := (&Powerline{}).Render(lightLine(200))
	line1, line2, _ := strings.Cut(out, "\n")
	if strings.Contains(line1, glyphs.MCP) || !strings.Contains(line2, glyphs.MCP+" 7") {
		t.Errorf("STATUSLINE_MCP_LINE=2: the pill is on line two only, got %q / %q", stripSGR(line1), stripSGR(line2))
	}
}

func TestMCPPillAbsentWithoutServers(t *testing.T) {
	withMCPLine(t, false)
	data := lightLine(200)
	data.MCP = nil
	if out := (&Powerline{}).Render(data); strings.Contains(out, glyphs.MCP) {
		t.Errorf("no server, no pill, got %q", stripSGR(out))
	}
}
