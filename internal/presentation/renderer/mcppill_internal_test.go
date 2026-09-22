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

// mcpItemsCase is one expectation on mcpItems, busy entries marked "*".
type mcpItemsCase struct {
	name    string
	servers model.MCPServers
	want    string
}

// mcpItemsCases lists the mcpItems expectations.
func mcpItemsCases() []mcpItemsCase {
	cli, user, plugin := model.MCPSourceCLI, model.MCPSourceUser, model.MCPSourcePlugin
	return []mcpItemsCase{
		{name: "nothing", servers: nil, want: ""},
		{
			name: "counted per scope, precedence order, empty scopes left out",
			servers: model.MCPServers{
				{Name: "p", Enabled: true, Source: plugin},
				{Name: "a", Enabled: true, Source: cli}, {Name: "b", Enabled: true, Source: cli},
				{Name: "u", Enabled: true, Source: user}, {Name: "m", Enabled: true, Source: model.MCPSourceManaged},
				{Name: "l", Enabled: true, Source: model.MCPSourceLocal}, {Name: "j", Enabled: true, Source: model.MCPSourceProject},
			},
			want: "managed 1|cli 2|local 1|project 1|user 1|plugin 1",
		},
		{
			name: "disabled counted once at the end, their scope emptied",
			servers: model.MCPServers{
				{Name: "a", Enabled: true, Source: cli}, {Name: "x", Source: user}, {Name: "y", Source: cli},
			},
			want: "cli 1|off 2",
		},
		{
			name:    "only disabled",
			servers: model.MCPServers{{Name: "x", Source: user}},
			want:    "off 1",
		},
		{
			name: "an unknown server is named after the scopes, before off",
			servers: model.MCPServers{
				{Name: "x", Source: user}, {Name: "a", Enabled: true, Source: cli},
				{Name: "ghost", Enabled: true, Busy: true},
			},
			want: "cli 1|+ghost*|off 1",
		},
		{
			name: "the scope owning a busy server lights, not the others",
			servers: model.MCPServers{
				{Name: "a", Enabled: true, Source: cli}, {Name: "b", Enabled: true, Source: cli, Busy: true},
				{Name: "u", Enabled: true, Source: user},
			},
			want: "cli 2*|user 1",
		},
		{
			name:    "a busy disabled server lights the off count",
			servers: model.MCPServers{{Name: "a", Enabled: true, Source: cli}, {Name: "x", Source: user, Busy: true}},
			want:    "cli 1|off 1*",
		},
	}
}

func TestMCPItems(t *testing.T) {
	for _, tt := range mcpItemsCases() {
		t.Run(tt.name, func(t *testing.T) {
			parts := make([]string, 0)
			for _, it := range mcpItems(tt.servers) {
				text := it.text
				if it.busy {
					text += "*"
				}
				parts = append(parts, text)
			}
			if got := strings.Join(parts, "|"); got != tt.want {
				t.Errorf("mcpItems() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPowerline_renderMCPPill(t *testing.T) {
	head := " " + LeftRound + " MCP " + glyphs.MCPArrow
	cli, user, plugin := model.MCPSourceCLI, model.MCPSourceUser, model.MCPSourcePlugin
	tests := []struct {
		name    string
		servers model.MCPServers
		want    string
	}{
		{name: "empty draws nothing", servers: model.MCPServers{}, want: ""},
		{
			name: "counts per scope",
			servers: model.MCPServers{
				{Name: "a", Enabled: true, Source: cli}, {Name: "b", Enabled: true, Source: cli},
				{Name: "c", Enabled: true, Source: cli}, {Name: "d", Enabled: true, Source: cli},
				{Name: "e", Enabled: true, Source: cli}, {Name: "u", Enabled: true, Source: user},
				{Name: "p", Enabled: true, Source: plugin},
			},
			want: head + " cli 5 · user 1 · plugin 1 " + RightRound,
		},
		{
			name:    "disabled after enabled",
			servers: model.MCPServers{{Name: "z", Source: cli}, {Name: "b", Enabled: true, Source: user}, {Name: "a", Source: plugin}},
			want:    head + " user 1 · off 2 " + RightRound,
		},
		{
			name:    "only disabled",
			servers: model.MCPServers{{Name: "off", Source: user}},
			want:    head + " off 1 " + RightRound,
		},
		{
			name:    "an unknown busy server",
			servers: model.MCPServers{{Name: "a", Enabled: true, Source: user}, {Name: "ghost", Enabled: true, Busy: true}},
			want:    head + " user 1 · +ghost " + RightRound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sb strings.Builder
			(&Powerline{}).renderMCPPill(&sb, tt.servers)
			if got := stripSGR(sb.String()); got != tt.want {
				t.Errorf("pill = %q, want %q", got, tt.want)
			}
			if tt.want != "" && strings.Count(sb.String(), LeftRound) != 1 {
				t.Errorf("want one pill, got %q", sb.String())
			}
		})
	}
}

func TestPowerline_renderMCPPillStyling(t *testing.T) {
	var sb strings.Builder
	(&Powerline{}).renderMCPPill(&sb, model.MCPServers{
		{Name: "on", Enabled: true, Source: model.MCPSourceUser},
		{Name: "down", Source: model.MCPSourceUser},
		{Name: "lit", Enabled: true, Source: model.MCPSourceCLI, Busy: true},
		{Name: "ghost", Enabled: true, Busy: true},
	})
	out := sb.String()
	label := " " + FgMCPEnabledText + LeftRound + Reset + BgMCPLabel + FgWhite + Bold + " MCP " + Reset
	if !strings.HasPrefix(out, label+BgMCPEnabled+FgMCPEnabledText+glyphs.MCPArrow+" ") {
		t.Errorf("the pill opens with the white label on dark teal, then the arrow, got %q", out)
	}
	if !strings.Contains(out, FgMCPEnabledText+"user 1"+Reset+BgMCPEnabled) {
		t.Errorf("an idle scope is ink 23 on the light teal, got %q", out)
	}
	if !strings.Contains(out, FgMCPMuted+mcpSeparator+FgMCPMuted+StrikeMCP+"off 1"+Reset+BgMCPEnabled) {
		t.Errorf("the disabled count is muted and crossed out on the teal, got %q", out)
	}
	if !strings.Contains(out, BgMCPLabel+FgWhite+Bold+"cli 1"+Reset+BgMCPEnabled) {
		t.Errorf("the scope owning a call is a bold white chip on dark teal, got %q", out)
	}
	if !strings.Contains(out, BgMCPLabel+FgWhite+Bold+"+ghost"+Reset+BgMCPEnabled) {
		t.Errorf("an unknown server being called is a chip of its own, got %q", out)
	}
	if strings.Contains(out, StrikeMCP+"user") || strings.Contains(out, BgMCPLabel+FgWhite+Bold+"user") {
		t.Errorf("an idle scope is neither crossed out nor lit, got %q", out)
	}
	if !strings.HasSuffix(out, " "+Reset+FgMCPEnabled+RightRound+Reset) {
		t.Errorf("the pill closes with a light teal cap, got %q", out)
	}
	if got := stripSGR(out); !strings.Contains(got, "cli 1 · user 1 · +ghost · off 1") {
		t.Errorf("order: scopes, unknown, off; got %q", got)
	}
}

func TestPowerline_renderMCPPillTextGlyphs(t *testing.T) {
	saved := glyphs
	t.Cleanup(func() { glyphs = saved })
	glyphs = textGlyphs
	var sb strings.Builder
	(&Powerline{}).renderMCPPill(&sb, model.MCPServers{{Name: "x", Enabled: true, Source: model.MCPSourceCLI}})
	if !strings.Contains(sb.String(), " MCP "+Reset+BgMCPEnabled+FgMCPEnabledText+"> "+FgMCPEnabledText+"cli 1") {
		t.Errorf("text glyphs: want the MCP label then '>', got %q", sb.String())
	}
}
