# GUI Design Communication Template

## Purpose
This document provides a structured way to communicate GUI designs between human and AI without requiring visual screenshots.

## Current GUI Structure (v0.1.0)

```
╔═══════════════════════════════════════════════════════════════════════╗
║ MTGA Companion                                          [_] [□] [X]   ║
╠═══════════════════════════════════════════════════════════════════════╣
║ ┌─Statistics──┬─Match History─┬─Charts────────────────────────────┐  ║
║ │                                                                   │  ║
║ │  Overall Statistics                                               │  ║
║ │  ==================                                               │  ║
║ │                                                                   │  ║
║ │  Matches: 150 (90-60)                                            │  ║
║ │  Win Rate: 60.0%                                                 │  ║
║ │                                                                   │  ║
║ │  Games: 420 (252-168)                                            │  ║
║ │  Game Win Rate: 60.0%                                            │  ║
║ │                                                                   │  ║
║ │                                                                   │  ║
║ │                                                                   │  ║
║ │                                                                   │  ║
║ │                                                                   │  ║
║ │                                                                   │  ║
║ │                                                                   │  ║
║ │                                                                   │  ║
║ │                                                                   │  ║
║ │  ┌────────────┐                                                  │  ║
║ │  │  Refresh   │                                                  │  ║
║ │  └────────────┘                                                  │  ║
║ └───────────────────────────────────────────────────────────────────┘  ║
║                                                    800x600             ║
╚═══════════════════════════════════════════════════════════════════════╝
```

## Design Communication Methods

### Method 1: ASCII Wireframes (Best for Layout)

Use box-drawing characters to show component hierarchy and layout:

```
┌─Window Title──────────────────────────────────────────────┐
│ [Menu Bar]                                                 │
├────────────────────────────────────────────────────────────┤
│ ┌─Left Panel (200px)─┬─Main Content (600px)─────────────┐ │
│ │                     │                                   │ │
│ │ • Navigation Item 1 │  ╔═ Card Title ════════════════╗ │ │
│ │ • Navigation Item 2 │  ║                              ║ │ │
│ │ • Navigation Item 3 │  ║  Content here                ║ │ │
│ │                     │  ║                              ║ │ │
│ │                     │  ╚══════════════════════════════╝ │ │
│ │                     │                                   │ │
│ └─────────────────────┴───────────────────────────────────┘ │
│ [Status Bar]                                               │
└────────────────────────────────────────────────────────────┘
```

### Method 2: Component Tree (Best for Structure)

```
Window: "MTGA Companion" (800x600)
├─ AppTabs (container)
│  ├─ Tab: "Statistics"
│  │  └─ Border Container
│  │     ├─ Top: null
│  │     ├─ Bottom: Button "Refresh"
│  │     ├─ Left: null
│  │     ├─ Right: null
│  │     └─ Center: Scroll(Label with stats text)
│  │
│  ├─ Tab: "Match History"
│  │  └─ Border Container
│  │     ├─ Top: Filter Controls
│  │     │  ├─ Entry: search box
│  │     │  ├─ Select: format dropdown
│  │     │  ├─ Select: result dropdown
│  │     │  └─ HBox: date filters
│  │     ├─ Center: List widget (matches)
│  │     └─ Bottom: Pagination controls
│  │
│  └─ Tab: "Charts"
│     └─ AppTabs (nested)
│        ├─ Tab: "Win Rate Trend"
│        ├─ Tab: "Result Breakdown"
│        └─ Tab: "Rank Progression"
```

### Method 3: Fyne Code Pseudocode (Best for Implementation)

```go
window := NewWindow("Title", 800, 600)

content := container.NewBorder(
    // TOP
    widget.NewLabel("Header"),
    
    // BOTTOM
    widget.NewButton("Action", callback),
    
    // LEFT
    nil,
    
    // RIGHT
    nil,
    
    // CENTER
    container.NewVBox(
        widget.NewLabel("Item 1"),
        widget.NewLabel("Item 2"),
    ),
)
```

### Method 4: Textual Description with Zones

```
ZONE LAYOUT:
┌─────────────────────────────────┐
│ A: Header (full width, 60px)   │
├─────────┬───────────────────────┤
│ B: Side │ C: Main Content       │
│ (200px) │ (600px)               │
│         │                       │
├─────────┴───────────────────────┤
│ D: Footer (full width, 40px)   │
└─────────────────────────────────┘

ZONE A (Header):
- Title: "Dashboard"
- Settings icon (top-right)
- User profile (top-right)

ZONE B (Sidebar):
- Navigation menu (vertical list)
  * Dashboard (active)
  * Statistics
  * Charts
  * Settings

ZONE C (Main):
- Card grid (2 columns)
  * Card 1: Win Rate (shows 60.5%)
  * Card 2: Total Matches (shows 150)
  * Card 3: Recent Performance chart
  * Card 4: Format breakdown

ZONE D (Footer):
- Status text (left)
- Action buttons (right)
```

### Method 5: Interaction Flow

```
USER ACTION               UI RESPONSE                 STATE CHANGE
─────────────────────────────────────────────────────────────────
Click "Statistics" tab -> Switch view                -> activeTab = 0
                       -> Show stats content         -> statsData loaded
                       -> Highlight tab              -> UI updates

Click "Refresh" button -> Show loading spinner       -> isLoading = true
                       -> Fetch new data from DB     -> query database
                       -> Update display             -> statsData = new
                       -> Hide spinner               -> isLoading = false

Type in search box     -> Filter match list          -> filterText = input
(debounced 300ms)      -> Show filtered results      -> matchList updates
                       -> Update result count        -> countLabel = N
```

## Design Change Request Template

When requesting a GUI change, use this format:

```markdown
### Change Request: [Brief Title]

**Location:** Tab/Section/Component path
**Current Behavior:** [What it looks like/does now]
**Desired Behavior:** [What you want it to look like/do]
**Priority:** High/Medium/Low

**ASCII Mockup:**
[Draw the desired layout]

**Component Tree:**
[Show the structure]

**Rationale:**
[Why this change improves UX]

**User Story:**
As a [user type], I want to [action] so that [benefit].

**Acceptance Criteria:**
- [ ] Criterion 1
- [ ] Criterion 2
```

## Example Change Request

```markdown
### Change Request: Add Settings Panel to Main Window

**Location:** Main Window → New Tab

**Current Behavior:**
- No settings accessible from GUI
- User must use CLI flags
- 3 tabs: Statistics, Match History, Charts

**Desired Behavior:**
- 4th tab: "Settings"
- UI controls for all CLI configuration
- Save/Apply buttons

**ASCII Mockup:**
┌─Statistics─┬─Match History─┬─Charts─┬─Settings─────────┐
│            │                │        │                  │
│            │                │        │  ┌─Log Config──┐ │
│            │                │        │  │              │ │
│            │                │        │  │ Path: [...] │ │
│            │                │        │  │ Poll: [2s]  │ │
│            │                │        │  └─────────────┘ │
│            │                │        │                  │
│            │                │        │  ┌─Cache Config┐│
│            │                │        │  │              ││
│            │                │        │  │ Enable: [✓] ││
│            │                │        │  │ TTL: [24h]  ││
│            │                │        │  └─────────────┘│
│            │                │        │                  │
│            │                │        │  [Save] [Apply]  │
└────────────┴────────────────┴────────┴──────────────────┘

**Component Tree:**
Tab: "Settings"
└─ VBox
   ├─ Accordion "Log Configuration"
   │  ├─ Entry: log-file-path
   │  ├─ Entry: log-poll-interval
   │  └─ Check: log-use-fsnotify
   ├─ Accordion "Cache Configuration"
   │  ├─ Check: cache-enabled
   │  ├─ Entry: cache-ttl
   │  └─ Entry: cache-max-size
   └─ HBox
      ├─ Button: "Save"
      └─ Button: "Apply"

**Rationale:**
Users shouldn't need CLI knowledge to configure the app. This improves
accessibility and matches user expectations for GUI applications.

**User Story:**
As a casual user, I want to adjust settings through the GUI so that
I don't have to learn CLI flags or restart the application.

**Acceptance Criteria:**
- [ ] Settings tab accessible from main window
- [ ] All CLI flags have UI equivalents
- [ ] Changes save to config file
- [ ] Apply button updates without restart (where possible)
- [ ] Input validation with helpful error messages
```

## Quick Reference: Box Drawing Characters

```
Single Line:
┌─┬─┐  ╔═╦═╗  ╒═╤═╕  ╓─╥─╖
│ │ │  ║ ║ ║  │ │ │  ║ ║ ║
├─┼─┤  ╠═╬═╣  ╞═╪═╡  ╟─╫─╢
│ │ │  ║ ║ ║  │ │ │  ║ ║ ║
└─┴─┘  ╚═╩═╝  ╘═╧═╛  ╙─╨─╜

Special:
▲ ▼ ◀ ▶  arrows
● ○ ◉ ◎  bullets/radio
✓ ✗ ☐ ☑ ☒  checkboxes
⚙ ⚡ ⚠ ⛔  icons
```

## Tips for Effective GUI Communication

1. **Be Specific About Dimensions:** "200px wide" not "small"
2. **Use Relative Positioning:** "Below the search bar" not "at y=150"
3. **Describe Interaction:** "On click, show..." not just static state
4. **Mention Colors/Styling:** "Primary button (blue)" vs "Danger button (red)"
5. **Reference Existing Components:** "Similar to the Statistics tab layout"
6. **Provide Context:** Explain WHY a change is needed
7. **Show Before/After:** Draw both current and desired state

## Additional Resources

- **Fyne Widget Catalog:** https://fyne.io/widgets/
- **Fyne Layouts:** https://developer.fyne.io/container/
- **Material Design Patterns:** For general GUI principles
