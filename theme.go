package main

import "github.com/gdamore/tcell/v2"

// ThemeName identifies a color theme.
type ThemeName string

const (
	ThemeDefault  ThemeName = "default"
	ThemeDark     ThemeName = "dark"
	ThemeLight    ThemeName = "light"
	ThemeSolarize ThemeName = "solarized"
	ThemeMonokai  ThemeName = "monokai"
	ThemeNord     ThemeName = "nord"
)

// ThemeColors holds all colors for a theme.
type ThemeColors struct {
	Name                string
	PanelBorderActive   tcell.Color
	PanelBorderInactive tcell.Color
	MenuBarBg           tcell.Color
	MenuBarFg           tcell.Color
	MenuBarHotkeyFg     tcell.Color
	DropdownBg          tcell.Color
	DropdownFg          tcell.Color
	DropdownSelected    tcell.Color
}

var themes = map[ThemeName]ThemeColors{
	ThemeDefault: {
		Name:                "Default",
		PanelBorderActive:   tcell.ColorAqua,
		PanelBorderInactive: tcell.ColorDefault,
		MenuBarBg:           tcell.ColorNavy,
		MenuBarFg:           tcell.ColorWhite,
		MenuBarHotkeyFg:     tcell.ColorYellow,
		DropdownBg:          tcell.ColorNavy,
		DropdownFg:          tcell.ColorWhite,
		DropdownSelected:    tcell.ColorTeal,
	},
	ThemeDark: {
		Name:                "Dark",
		PanelBorderActive:   tcell.ColorGreen,
		PanelBorderInactive: tcell.ColorDarkGray,
		MenuBarBg:           tcell.ColorBlack,
		MenuBarFg:           tcell.ColorSilver,
		MenuBarHotkeyFg:     tcell.ColorGreen,
		DropdownBg:          tcell.ColorBlack,
		DropdownFg:          tcell.ColorSilver,
		DropdownSelected:    tcell.ColorDarkGreen,
	},
	ThemeLight: {
		Name:                "Light",
		PanelBorderActive:   tcell.ColorBlue,
		PanelBorderInactive: tcell.ColorGray,
		MenuBarBg:           tcell.ColorWhite,
		MenuBarFg:           tcell.ColorBlack,
		MenuBarHotkeyFg:     tcell.ColorBlue,
		DropdownBg:          tcell.ColorWhite,
		DropdownFg:          tcell.ColorBlack,
		DropdownSelected:    tcell.ColorLightBlue,
	},
	ThemeSolarize: {
		Name:                "Solarized",
		PanelBorderActive:   tcell.NewRGBColor(38, 139, 210),  // blue
		PanelBorderInactive: tcell.NewRGBColor(88, 110, 117),  // base01
		MenuBarBg:           tcell.NewRGBColor(0, 43, 54),     // base03
		MenuBarFg:           tcell.NewRGBColor(131, 148, 150), // base0
		MenuBarHotkeyFg:     tcell.NewRGBColor(181, 137, 0),   // yellow
		DropdownBg:          tcell.NewRGBColor(7, 54, 66),     // base02
		DropdownFg:          tcell.NewRGBColor(131, 148, 150), // base0
		DropdownSelected:    tcell.NewRGBColor(38, 139, 210),  // blue
	},
	ThemeMonokai: {
		Name:                "Monokai",
		PanelBorderActive:   tcell.NewRGBColor(166, 226, 46),  // green
		PanelBorderInactive: tcell.NewRGBColor(117, 113, 94),  // comment
		MenuBarBg:           tcell.NewRGBColor(39, 40, 34),    // bg
		MenuBarFg:           tcell.NewRGBColor(248, 248, 242), // fg
		MenuBarHotkeyFg:     tcell.NewRGBColor(249, 38, 114),  // pink
		DropdownBg:          tcell.NewRGBColor(39, 40, 34),    // bg
		DropdownFg:          tcell.NewRGBColor(248, 248, 242), // fg
		DropdownSelected:    tcell.NewRGBColor(73, 72, 62),    // line highlight
	},
	ThemeNord: {
		Name:                "Nord",
		PanelBorderActive:   tcell.NewRGBColor(136, 192, 208), // frost
		PanelBorderInactive: tcell.NewRGBColor(76, 86, 106),   // polar night 3
		MenuBarBg:           tcell.NewRGBColor(46, 52, 64),    // polar night 0
		MenuBarFg:           tcell.NewRGBColor(216, 222, 233), // snow storm
		MenuBarHotkeyFg:     tcell.NewRGBColor(235, 203, 139), // aurora yellow
		DropdownBg:          tcell.NewRGBColor(59, 66, 82),    // polar night 1
		DropdownFg:          tcell.NewRGBColor(216, 222, 233), // snow storm
		DropdownSelected:    tcell.NewRGBColor(136, 192, 208), // frost
	},
}

// GetTheme returns the ThemeColors for the given name, defaulting to ThemeDefault.
func GetTheme(name ThemeName) ThemeColors {
	if tc, ok := themes[name]; ok {
		return tc
	}
	return themes[ThemeDefault]
}

// AllThemes returns all available theme names in display order.
func AllThemes() []ThemeName {
	return []ThemeName{
		ThemeDefault,
		ThemeDark,
		ThemeLight,
		ThemeSolarize,
		ThemeMonokai,
		ThemeNord,
	}
}
