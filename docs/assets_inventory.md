# Asset Inventory

Lists every non-code asset the game needs, its current status, and the
minimum specification required before it can be committed to the repo.
All assets must have a CREDITS.md entry before first commit (see CLAUDE.md).

---

## Status key

| Tag | Meaning |
|---|---|
| PLACEHOLDER | Can be faked in Go code for MVP (coloured rectangle, text label, etc.) |
| NEEDED-MVP | Required before MVP can run correctly; no viable code substitute |
| NEEDED-PRODUCTION | Required for a shippable game; not blocking MVP |

---

## 1. Map geometry  (NEEDED-MVP)

The Map tab is the core of the game. Without geometry data the tab
degrades to a region list, which is acceptable for MVP only if explicitly
marked as placeholder. A real Tiled map is required before the game is
content-complete.

| Asset | Path | Specification |
|---|---|---|
| Taitan region tilemap | `app/assets/map/taitan_regions.tmx` | Tiled `.tmx` file. 12 named region objects (matching `config.RegionDef.ID`). Each region contains ~2-4 named tile objects (matching `config.TileDef.ID`). Coordinate system: top-left origin, 1px = 1km scale recommended. Must use MIT/CC0/Apache-licensed Tiled tileset or a custom one. |
| Tileset image | `app/assets/map/taitan_tileset.png` | Referenced by the `.tmx`. Can be a simple coloured grid for MVP. Requires permissive licence. |
| Region boundary overlay | `app/assets/map/region_borders.png` | Optional but recommended. Vector-style borders drawn over the tileset as a separate layer. |

**MVP substitute**: A 12-row table (one row per region) with name, installer
capacity, avg fuel poverty, and skills network. Tile drill-down shown as a
secondary list. This gives all the data, just not the spatial view.

---

## 2. Fonts  (NEEDED-MVP)

ebitenui requires a loaded `font.Face` for every text element. Go's
`golang.org/x/image/font/basicfont` (public domain) provides a fixed
7x13px bitmap face and requires no file asset -- suitable for MVP.
Production needs a proper TrueType font for readability at different
window sizes.

| Asset | Path | Specification |
|---|---|---|
| UI regular | `app/assets/fonts/ui_regular.ttf` | TrueType. Any weight 400. Recommended: Roboto Regular (Apache 2.0, Google Fonts). Min sizes used: 12px body, 14px labels, 18px headings. |
| UI bold | `app/assets/fonts/ui_bold.ttf` | Weight 700. Recommended: Roboto Bold (Apache 2.0). Used for headings and badge labels. |
| Monospace | `app/assets/fonts/ui_mono.ttf` | For numerical readouts (prices, carbon values). Recommended: Roboto Mono (Apache 2.0). |

**MVP substitute**: `golang.org/x/image/font/basicfont.Face7x13` --
already available via an indirect dependency, no download needed.

---

## 3. Party colour palette  (PLACEHOLDER)

No file asset. Defined as Go constants in the ui package. Values below
are the design-intent reference so all developers use the same palette.

| Party | Colour | Hex |
|---|---|---|
| The Common Wealth (Left) | Warm red | `#C0392B` |
| The Union Party (Right) | Navy blue | `#1A3A5C` |
| Renewal (FarLeft) | Vivid green | `#27AE60` |
| Taitan Restoration (FarRight) | Deep purple | `#6C3483` |
| Neutral / no party | Mid grey | `#7F8C8D` |

---

## 4. Climate level colour palette  (PLACEHOLDER)

No file asset. Used for the climate badge in the HUD.

| Level | Colour | Hex |
|---|---|---|
| LOW | Soft green | `#2ECC71` |
| MEDIUM | Amber | `#F39C12` |
| HIGH | Orange | `#E67E22` |
| CRITICAL | Red | `#E74C3C` |
| EMERGENCY | Dark red | `#922B21` |

---

## 5. Technology icons  (NEEDED-PRODUCTION / PLACEHOLDER for MVP)

One icon per technology type. MVP uses a coloured text badge with the
technology abbreviation.

| Technology | Abbreviation | Path |
|---|---|---|
| OFFSHORE_WIND | OWind | `app/assets/icons/tech/offshore_wind.png` |
| ONSHORE_WIND | Wind | `app/assets/icons/tech/onshore_wind.png` |
| SOLAR_PV | Solar | `app/assets/icons/tech/solar_pv.png` |
| NUCLEAR | Nuke | `app/assets/icons/tech/nuclear.png` |
| HEAT_PUMPS | HeatP | `app/assets/icons/tech/heat_pumps.png` |
| EVS | EVs | `app/assets/icons/tech/evs.png` |
| HYDROGEN | H2 | `app/assets/icons/tech/hydrogen.png` |
| INDUSTRIAL_CCS | CCS | `app/assets/icons/tech/industrial_ccs.png` |

Specification: 24x24px PNG, transparent background, permissive licence.

---

## 6. Event type icons  (NEEDED-PRODUCTION / PLACEHOLDER for MVP)

One icon per event type. MVP uses a coloured dot.

| Event type | Colour | Path |
|---|---|---|
| WEATHER | Steel blue | `app/assets/icons/events/weather.png` |
| ENERGY_SHOCK | Orange | `app/assets/icons/events/energy_shock.png` |
| INTERNATIONAL | Gold | `app/assets/icons/events/international.png` |
| ECONOMIC | Green | `app/assets/icons/events/economic.png` |
| SOCIAL | Pink | `app/assets/icons/events/social.png` |
| TECHNOLOGICAL | Cyan | `app/assets/icons/events/technological.png` |

Specification: 16x16px PNG, transparent background, permissive licence.

---

## 7. Minister portrait placeholders  (NEEDED-PRODUCTION / PLACEHOLDER for MVP)

One portrait per named cast member (~30+ figures). MVP uses a coloured
circle with the minister's initials rendered in white text.

| Path | Specification |
|---|---|
| `app/assets/portraits/<stakeholder_id>.png` | 64x64px PNG. Stylised illustration, not photographic. Permissive licence or original work. |

The stakeholder IDs are defined in `app/internal/config/stakeholders.go`.

---

## 8. Department icons  (NEEDED-PRODUCTION / PLACEHOLDER for MVP)

Five small icons for the Budget and Politics tabs.

| Department | Path |
|---|---|
| Power | `app/assets/icons/depts/power.png` |
| Transport | `app/assets/icons/depts/transport.png` |
| Buildings | `app/assets/icons/depts/buildings.png` |
| Industry | `app/assets/icons/depts/industry.png` |
| Cross-cutting | `app/assets/icons/depts/cross.png` |

Specification: 20x20px PNG, transparent background, permissive licence.

---

## 9. Org type badges  (PLACEHOLDER)

No file asset. Three badge styles rendered as coloured rounded rectangles
in code. Reference:

| Org type | Colour | Label |
|---|---|---|
| CONSULTANCY | `#8E44AD` (purple) | `CONSULT` |
| THINK_TANK | `#2980B9` (blue) | `THINK` |
| ACADEMIC | `#16A085` (teal) | `ACADEMIC` |

---

## 10. Audio  (OUT OF SCOPE for MVP)

Ambient music, event stings, and UI click sounds. Deferred to
post-content-complete milestone. All audio assets must be CC0 or CC-BY.

| Asset | Path | Notes |
|---|---|---|
| Main theme | `app/assets/audio/main_theme.ogg` | Looping ambient. 60-120 BPM. Climate/civic feel. |
| Event sting | `app/assets/audio/event_sting.ogg` | Short (1-2s). Plays on new event. |
| Week advance click | `app/assets/audio/week_advance.ogg` | Short UI click (0.1-0.2s). |

---

## Summary: what must exist before MVP can be called complete

1. **Fonts** -- either embed Roboto (needs download + CREDITS entry) or accept the
   basicfont bitmap fallback for the MVP milestone.
2. **Map geometry** -- either build a Tiled `.tmx` for Taitan or formally accept
   the region-list placeholder as the MVP map tab.
3. All other tabs are fully serviceable with Go-coded placeholders (colour patches,
   text labels, initials circles).
