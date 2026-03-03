package views

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ccss/internal/model"
	"github.com/ccss/internal/theme"
)

// StatsPeriod controls the aggregation level.
type StatsPeriod int

const (
	PeriodDaily StatsPeriod = iota
	PeriodWeekly
	PeriodMonthly
)

// StatsView displays cost statistics and usage dashboard.
type StatsView struct {
	viewport viewport.Model
	stats    *model.StatsCache
	period   StatsPeriod
	ready    bool

	width, height int
}

func NewStatsView() *StatsView {
	return &StatsView{
		period: PeriodDaily,
	}
}

func (v *StatsView) SetSize(w, h int) {
	v.width = w
	v.height = h
	if v.ready {
		v.viewport.Width = w
		v.viewport.Height = h - 2
	}
}

func (v *StatsView) SetData(stats *model.StatsCache) {
	v.stats = stats
	if v.width == 0 {
		v.width = 80
	}
	if v.height == 0 {
		v.height = 24
	}
	v.viewport = viewport.New(v.width, v.height-2)
	v.viewport.SetContent(v.renderContent())
	v.ready = true
}

func (v *StatsView) Update(msg tea.Msg) tea.Cmd {
	if !v.ready {
		return nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "d":
			v.period = PeriodDaily
			v.viewport.SetContent(v.renderContent())
			v.viewport.GotoTop()
			return nil
		case "w":
			v.period = PeriodWeekly
			v.viewport.SetContent(v.renderContent())
			v.viewport.GotoTop()
			return nil
		case "m":
			v.period = PeriodMonthly
			v.viewport.SetContent(v.renderContent())
			v.viewport.GotoTop()
			return nil
		}
	}

	var cmd tea.Cmd
	v.viewport, cmd = v.viewport.Update(msg)
	return cmd
}

func (v *StatsView) renderContent() string {
	if v.stats == nil {
		return theme.EmptyStateStyle.Render("No statistics available.")
	}
	var b strings.Builder

	b.WriteString("\n")

	// ─── Summary Cards ────────────────────────────────────────────
	totalCost := v.calculateTotalCost()

	cardWidth := 22
	if v.width > 80 {
		cardWidth = (v.width - 12) / 3
		if cardWidth > 30 {
			cardWidth = 30
		}
	}

	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Border).
		Width(cardWidth).
		Padding(0, 1).
		Align(lipgloss.Center)

	card1 := cardStyle.Render(
		lipgloss.NewStyle().Foreground(theme.TextDim).Render("Sessions") + "\n" +
			lipgloss.NewStyle().Foreground(theme.Cyan).Bold(true).Render(fmt.Sprintf("%d", v.stats.TotalSessions)))

	card2 := cardStyle.Render(
		lipgloss.NewStyle().Foreground(theme.TextDim).Render("Messages") + "\n" +
			lipgloss.NewStyle().Foreground(theme.Cyan).Bold(true).Render(fmt.Sprintf("%d", v.stats.TotalMessages)))

	card3 := cardStyle.Render(
		lipgloss.NewStyle().Foreground(theme.TextDim).Render("Est. Cost") + "\n" +
			lipgloss.NewStyle().Foreground(theme.Green).Bold(true).Render(fmt.Sprintf("$%.2f", totalCost)))

	cards := lipgloss.JoinHorizontal(lipgloss.Top, "  ", card1, " ", card2, " ", card3)
	b.WriteString(cards + "\n\n")

	// ─── Activity Chart ──────────────────────────────────────────
	b.WriteString("  " + lipgloss.NewStyle().Foreground(theme.Accent).Bold(true).Render("Activity") + "\n")

	// Period selector
	periods := []struct {
		label  string
		key    string
		active bool
	}{
		{"Daily", "d", v.period == PeriodDaily},
		{"Weekly", "w", v.period == PeriodWeekly},
		{"Monthly", "m", v.period == PeriodMonthly},
	}

	var periodParts []string
	for _, p := range periods {
		label := fmt.Sprintf("[%s] %s", p.key, p.label)
		if p.active {
			periodParts = append(periodParts, lipgloss.NewStyle().Foreground(theme.Accent).Bold(true).Underline(true).Render(label))
		} else {
			periodParts = append(periodParts, lipgloss.NewStyle().Foreground(theme.TextDim).Render(label))
		}
	}
	b.WriteString("  " + strings.Join(periodParts, "  ") + "\n\n")

	v.renderActivityChart(&b)

	// ─── Separator ────────────────────────────────────────────────
	b.WriteString("\n  " + lipgloss.NewStyle().Foreground(theme.Border).Render(safeRepeat("─", v.width-4)) + "\n\n")

	// ─── Model Usage Table ────────────────────────────────────────
	b.WriteString("  " + lipgloss.NewStyle().Foreground(theme.Accent).Bold(true).Render("Model Usage") + "\n\n")
	v.renderModelTable(&b)

	// ─── Separator ────────────────────────────────────────────────
	b.WriteString("\n  " + lipgloss.NewStyle().Foreground(theme.Border).Render(safeRepeat("─", v.width-4)) + "\n\n")

	// ─── Peak Hours ───────────────────────────────────────────────
	b.WriteString("  " + lipgloss.NewStyle().Foreground(theme.Accent).Bold(true).Render("Peak Hours") + "\n\n")
	v.renderPeakHours(&b)

	return b.String()
}

func (v *StatsView) renderActivityChart(b *strings.Builder) {
	activity := v.aggregateActivity()
	if len(activity) == 0 {
		b.WriteString(theme.EmptyStateStyle.Render("No activity data.") + "\n")
		return
	}

	// Show last 14 entries
	start := 0
	if len(activity) > 14 {
		start = len(activity) - 14
	}
	entries := activity[start:]

	// Find max for scaling
	maxCount := 0
	for _, e := range entries {
		if e.count > maxCount {
			maxCount = e.count
		}
	}

	barWidth := v.width - 28
	if barWidth < 10 {
		barWidth = 10
	}

	for _, e := range entries {
		label := lipgloss.NewStyle().Foreground(theme.TextSec).Width(14).Align(lipgloss.Right).Render(e.label)
		bar := ""
		if maxCount > 0 {
			barLen := e.count * barWidth / maxCount
			if barLen == 0 && e.count > 0 {
				barLen = 1
			}
			bar = lipgloss.NewStyle().Foreground(theme.Accent).Render(safeRepeat("█", barLen))
			remaining := barWidth - barLen
			if remaining > 0 {
				bar += lipgloss.NewStyle().Foreground(theme.BorderDm).Render(safeRepeat("░", remaining))
			}
		}
		count := lipgloss.NewStyle().Foreground(theme.Cyan).Bold(true).Render(fmt.Sprintf(" %d", e.count))
		b.WriteString("  " + label + " " + bar + count + "\n")
	}
}

type activityEntry struct {
	label string
	count int
}

func (v *StatsView) aggregateActivity() []activityEntry {
	if len(v.stats.DailyActivity) == 0 {
		return nil
	}

	switch v.period {
	case PeriodDaily:
		var entries []activityEntry
		for _, d := range v.stats.DailyActivity {
			entries = append(entries, activityEntry{
				label: d.Date[5:], // MM-DD
				count: d.MessageCount,
			})
		}
		return entries

	case PeriodWeekly:
		weekly := make(map[string]int)
		var keys []string
		for _, d := range v.stats.DailyActivity {
			weekKey := d.Date[:8] + "W"
			if _, exists := weekly[weekKey]; !exists {
				keys = append(keys, weekKey)
			}
			weekly[weekKey] += d.MessageCount
		}
		var entries []activityEntry
		for _, k := range keys {
			entries = append(entries, activityEntry{
				label: k[5:],
				count: weekly[k],
			})
		}
		return entries

	case PeriodMonthly:
		monthly := make(map[string]int)
		var keys []string
		for _, d := range v.stats.DailyActivity {
			monthKey := d.Date[:7] // YYYY-MM
			if _, exists := monthly[monthKey]; !exists {
				keys = append(keys, monthKey)
			}
			monthly[monthKey] += d.MessageCount
		}
		sort.Strings(keys)
		var entries []activityEntry
		for _, k := range keys {
			entries = append(entries, activityEntry{
				label: k,
				count: monthly[k],
			})
		}
		return entries
	}
	return nil
}

func (v *StatsView) renderModelTable(b *strings.Builder) {
	// Table header
	headerFmt := fmt.Sprintf("  %-26s %12s %12s %14s %14s %10s",
		"Model", "Input", "Output", "Cache Read", "Cache Write", "Cost")
	b.WriteString(lipgloss.NewStyle().Foreground(theme.White).Bold(true).Render(headerFmt) + "\n")
	b.WriteString("  " + lipgloss.NewStyle().Foreground(theme.Border).Render(safeRepeat("─", 90)) + "\n")

	for modelID, usage := range v.stats.ModelUsage {
		pricing := model.LookupPricing(modelID)
		cost := model.CalculateCost(pricing,
			usage.InputTokens,
			usage.OutputTokens,
			usage.CacheReadInputTokens,
			usage.CacheCreationInputTokens,
		)

		name := shortModel(modelID)
		row := fmt.Sprintf("  %-26s %12s %12s %14s %14s",
			lipgloss.NewStyle().Foreground(theme.TextPri).Render(name),
			formatTokenCount(usage.InputTokens),
			formatTokenCount(usage.OutputTokens),
			formatTokenCount(usage.CacheReadInputTokens),
			formatTokenCount(usage.CacheCreationInputTokens),
		)
		costStr := lipgloss.NewStyle().Foreground(theme.Green).Bold(true).Render(fmt.Sprintf("$%.2f", cost))
		b.WriteString(row + " " + costStr + "\n")
	}
}

func (v *StatsView) renderPeakHours(b *strings.Builder) {
	if len(v.stats.HourCounts) == 0 {
		b.WriteString(theme.EmptyStateStyle.Render("No hour data.") + "\n")
		return
	}

	type hourEntry struct {
		hour  string
		count int
	}
	var entries []hourEntry
	maxCount := 0
	for h, c := range v.stats.HourCounts {
		entries = append(entries, hourEntry{hour: h, count: c})
		if c > maxCount {
			maxCount = c
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].count > entries[j].count
	})

	limit := 10
	if len(entries) < limit {
		limit = len(entries)
	}

	barWidth := 30
	for i := 0; i < limit; i++ {
		e := entries[i]
		label := lipgloss.NewStyle().Foreground(theme.TextSec).Width(8).Align(lipgloss.Right).Render(e.hour + "h")
		barLen := 0
		if maxCount > 0 {
			barLen = e.count * barWidth / maxCount
			if barLen == 0 && e.count > 0 {
				barLen = 1
			}
		}
		bar := lipgloss.NewStyle().Foreground(theme.Accent).Render(safeRepeat("█", barLen))
		remaining := barWidth - barLen
		if remaining > 0 {
			bar += lipgloss.NewStyle().Foreground(theme.BorderDm).Render(safeRepeat("░", remaining))
		}
		count := lipgloss.NewStyle().Foreground(theme.Cyan).Bold(true).Render(fmt.Sprintf(" %d", e.count))
		b.WriteString("  " + label + " " + bar + count + "\n")
	}
}

func (v *StatsView) calculateTotalCost() float64 {
	total := 0.0
	for modelID, usage := range v.stats.ModelUsage {
		pricing := model.LookupPricing(modelID)
		total += model.CalculateCost(pricing,
			usage.InputTokens,
			usage.OutputTokens,
			usage.CacheReadInputTokens,
			usage.CacheCreationInputTokens,
		)
	}
	return total
}

func formatTokenCount(n int) string {
	if n >= 1_000_000_000 {
		return fmt.Sprintf("%.1fB", float64(n)/1_000_000_000)
	}
	if n >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	}
	if n >= 1_000 {
		return fmt.Sprintf("%.1fK", float64(n)/1_000)
	}
	return fmt.Sprintf("%d", n)
}

func (v *StatsView) View() string {
	if !v.ready {
		return "  Loading statistics..."
	}
	return v.viewport.View()
}

func (v *StatsView) StatusKeys() string {
	return "↑↓:scroll  d:daily  w:weekly  m:monthly  esc:back"
}

func safeRepeat(s string, count int) string {
	if count <= 0 {
		return ""
	}
	return strings.Repeat(s, count)
}
