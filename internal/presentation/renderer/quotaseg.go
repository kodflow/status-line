// Package renderer provides status line rendering.
package renderer

import (
	"strconv"
	"strings"

	"github.com/florent/status-line/internal/domain/model"
)

// Quota segment constants.
const (
	// segBarWidth is the bar width of a quota segment on the shared line.
	segBarWidth int = 10
)

// quotaSegment is one quota rendered as a chained powerline segment rather
// than as an isolated pill, so the whole line reads as a single ribbon the way
// the original status line did.
type quotaSegment struct {
	limit  model.Limit
	bg     string
	cap    string
	ink    string
	tokens int
	size   int
}

// quotaSegments builds the chain of quota segments for the shared line.
// Colours are fixed per bucket rather than derived from state: the ribbon is an
// identity, and a segment that changes hue as consumption moves would make the
// line flicker between redraws.
//
// Params:
//   - data: everything the status line renders
//
// Returns:
//   - []quotaSegment: quotas to chain, most local first
func quotaSegments(data model.StatusLineData) []quotaSegment {
	limits := data.Limits
	segments := make([]quotaSegment, 0, len(limits.Scoped)+3)

	// Lead with the context window: it is the quota that bites first, and the
	// only one that is always present
	if limits.Context.IsValid() && !isKindHidden(model.KindContext) {
		segments = append(segments, quotaSegment{
			limit:  limits.Context,
			bg:     BgHueContext,
			cap:    FgHueContext,
			ink:    FgHueContextInk,
			tokens: data.ContextTokens,
			size:   data.ContextSize,
		})
	}

	return segments
}

// modelQuotas returns the account quotas that belong beside the model name.
//
// The session and weekly windows, and any quota scoped to the model in use,
// are all limits on the same thing: what this account may spend on this model.
// They read as one group, so they are drawn as one.
//
// Params:
//   - data: everything the status line renders
//
// Returns:
//   - []model.Limit: quotas to draw inside the model segment
func modelQuotas(data model.StatusLineData) []model.Limit {
	limits := data.Limits
	quotas := make([]model.Limit, 0, 3)

	// The rolling five-hour window
	if limits.Session.IsValid() && !isKindHidden(model.KindSession) {
		quotas = append(quotas, limits.Session)
	}
	// The plan-wide seven-day window
	if limits.Weekly.IsValid() && !isKindHidden(model.KindWeekly) {
		quotas = append(quotas, limits.Weekly)
	}
	// Only the scoped quotas that restrict the model actually in use
	if !isKindHidden(model.KindScoped) {
		quotas = append(quotas, limits.ScopedFor(data.Model.FullName())...)
	}

	return quotas
}

// segmentColors returns the fixed colours of a quota segment.
//
// Params:
//   - limit: quota to colour
//
// Returns:
//   - bg: segment background
//   - cap: segment background as a foreground, for the separator
//   - ink: segment text colour
func segmentColors(limit model.Limit) (bg, cap, ink string) {
	// Map each bucket onto the palette the project already uses
	switch limit.Kind {
	// The rolling five-hour quota
	case model.KindSession:
		return BgHueSession, FgHueSession, FgHueSessionInk
	// The plan-wide seven-day quota keeps the grey the original line used
	case model.KindWeekly:
		return BgWeekly, FgWeekly, FgWeeklyText
	// A model-scoped quota takes the hue of the model it restricts
	case model.KindScoped:
		return GetModelColors(limit.Label)
	// Anything else falls back to the neutral ground
	default:
		return BgQuota, FgQuota, FgQuotaText
	}
}

// renderQuotaSegment writes one quota as a chained powerline segment.
//
// Params:
//   - sb: string builder to write to
//   - seg: quota segment to render
//   - nextBg: background of the segment that follows
func renderQuotaSegment(sb *strings.Builder, seg quotaSegment, nextBg string) {
	limit := seg.limit
	cursor := noCursor
	// Place the even-burn cursor only when the window makes it meaningful
	if limit.HasWindow() {
		cursor = limit.CursorPosition()
	}
	bar := RenderProgressBarWidth(limit.Progress(), cursor, segBarWidth, FgCursorOrange, seg.bg+seg.ink+Bold)

	// Write the label, the bar and the consumed percentage
	sb.WriteString(seg.bg + seg.ink + Bold + " " + QuotaLabel(limit) + " " + Reset)
	sb.WriteString(seg.bg + seg.ink + Bold + bar + " " + strconv.Itoa(limit.Percent) + "%" + Reset)
	// Append the countdown to the refill, which is what the bar cannot say
	if limit.HasWindow() {
		sb.WriteString(seg.bg + seg.ink + " " + glyphs.Reset + FormatDuration(limit.Remaining()) + Reset)
	}
	// The context window has no refill, so it carries its token figures instead
	if seg.size > 0 {
		sb.WriteString(seg.bg + seg.ink + " " + FormatTokens(seg.tokens) + "/" + FormatTokens(seg.size) + Reset)
	}
	sb.WriteString(seg.bg + " " + Reset)

	// Write the separator into whatever follows
	sb.WriteString(nextBg + seg.cap + SepRight + Reset)
}
