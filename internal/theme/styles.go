package theme

import "github.com/charmbracelet/lipgloss"

var (
	// ─── Color Palette (Minimal, High-Contrast) ─────────────────────
	// Only 4 semantic hues + neutral scale
	Accent = lipgloss.Color("#c084fc") // violet — brand/tabs/titles
	Blue   = lipgloss.Color("#7dd3fc") // user messages
	Green  = lipgloss.Color("#6ee7b7") // assistant messages
	Yellow = lipgloss.Color("#fcd34d") // keys/warnings
	Orange = lipgloss.Color("#fdba74") // tools
	Red    = lipgloss.Color("#fca5a5") // errors
	Cyan   = lipgloss.Color("#67e8f9") // numbers/stats
	Pink   = lipgloss.Color("#f0abfc") // thinking

	// Neutral scale (slate)
	White    = lipgloss.Color("#f8fafc")
	TextPri  = lipgloss.Color("#e2e8f0")
	TextSec  = lipgloss.Color("#94a3b8")
	TextDim  = lipgloss.Color("#475569")
	Border   = lipgloss.Color("#334155")
	BorderDm = lipgloss.Color("#1e293b")
	Bg       = lipgloss.Color("#0f172a") // main background
	BgCard   = lipgloss.Color("#1e293b") // card / surface
	BgHi     = lipgloss.Color("#334155") // selected highlight

	// ─── Tab Bar ────────────────────────────────────────────────────
	ActiveTabStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#0f172a")).
			Background(Accent).
			Padding(0, 1).
			Bold(true)

	InactiveTabStyle = lipgloss.NewStyle().
				Foreground(TextSec).
				Padding(0, 1)

	// ─── Status Bar ─────────────────────────────────────────────────
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(TextDim)

	StatusKeyStyle = lipgloss.NewStyle().
			Foreground(Yellow).
			Bold(true)

	// ─── Session List ───────────────────────────────────────────────
	SessionCursorStyle = lipgloss.NewStyle().
				Foreground(Accent).
				Bold(true)

	FilterInputStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(Accent).
				Padding(0, 1).
				MarginLeft(2)

	// ─── Search ─────────────────────────────────────────────────────
	SearchInputStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(Accent).
				Padding(0, 1).
				MarginLeft(2)

	// ─── Common ─────────────────────────────────────────────────────
	EmptyStateStyle = lipgloss.NewStyle().
			Foreground(TextDim).
			Italic(true).
			PaddingLeft(4).
			MarginTop(2)
)
