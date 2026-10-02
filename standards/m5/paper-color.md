# M5Paper Color Display Standards

How to drive the M5Paper Color panel, a Spectra E6 full-colour e-paper
display, correctly and efficiently.

## Hardware facts that shape the rules

- MCU: ESP32-S3. OPI PSRAM is required; the colour framebuffer lives there.
- Panel: 4" Spectra E6 full-colour e-paper (for example ED2208-DOA).
- Palette: six colours only, black, white, red, yellow, green, and blue. Any
  RGB565 value drawn is snapped to the nearest of these. There are no
  gradients and no photographs.
- Bistable and slow: the image persists with no power, and a refresh takes on
  the order of seconds while visibly flashing through intermediate states.

The last point drives everything below: refresh as rarely as possible, and
refresh in one operation.

## Compose off-screen, push once

Do not draw incrementally to the live panel. Every `drawString` or `fillRect`
issued directly against `M5.Display` can trigger a panel refresh, so many
small draws produce slow, flickering updates. Build the whole frame in an
off-screen `M5Canvas` sprite, a buffer in PSRAM, and push it in a single
operation.

```cpp
#include <M5Unified.h>

M5Canvas canvas(&M5.Display);   // off-screen buffer bound to the panel

void setup() {
  auto cfg = M5.config();
  cfg.clear_display = false;                       // keep the retained image
  M5.begin(cfg);
  M5.Display.setEpdMode(epd_mode_t::epd_fastest);  // see "Refresh modes"
  M5.Display.setRotation(0);
  canvas.createSprite(M5.Display.width(), M5.Display.height());  // once
}

static void pushFrame() {
  canvas.pushSprite(0, 0);     // single panel update
  M5.Display.waitDisplay();    // block until the refresh finishes
}

void drawFrame() {
  canvas.fillSprite(WHITE);                          // clear the buffer
  canvas.setFont(&fonts::FreeSansBold24pt7b);
  canvas.setTextDatum(middle_center);
  canvas.setTextColor(RED);
  canvas.drawString("123", M5.Display.width() / 2, 200);
  pushFrame();
}
```

- Create the sprite once, in `setup`, and reuse it for every frame. Never
  create and delete it per frame.
- Per frame: `fillSprite` to clear, draw, then one `pushSprite(0, 0)`.
- Call `waitDisplay()` after the push, and always before sleeping or powering
  off. Cutting power mid-refresh corrupts the image.

## Refresh modes

Set with `M5.Display.setEpdMode`:

| Mode | Behaviour | Use for |
| --- | --- | --- |
| `epd_fastest` | Quickest, lowest fidelity. | Interactive or frequently updated UI. The default. |
| `epd_fast` | Fast, with some ghosting. | |
| `epd_text` | Tuned for text. | |
| `epd_quality` | Best colour and contrast, slowest, most flashing. | A clean image that will stay up for a long time. |

For a battery-powered device that updates occasionally, prefer `epd_fastest`
so each update is short. Use `epd_quality` only when colour fidelity is poor.

## Colours

- Use the M5GFX constants `BLACK`, `WHITE`, `RED`, `YELLOW`, `GREEN`, and
  `BLUE`, which map exactly to the E6 palette. The `TFT_*` aliases also exist.
- Arbitrary RGB565 values are snapped to the nearest palette entry. This is
  fine for solid fills and unreliable for anything subtle.
- Use `canvas.setTextColor(c)` for text; pass the colour to `fillRect`,
  `drawRect`, and the other shape calls.

## Centring and per-character colour

After setting the font, `canvas.textWidth(str)` returns the pixel width of a
string. Use it to centre text, or to colour characters individually by summing
per-character widths and advancing `x` yourself.

## Memory and PSRAM

- A full-screen 16-bit sprite is `width * height * 2` bytes, about 0.5 MB,
  and must be in PSRAM. `createSprite` falls back to PSRAM automatically for
  large sprites, but OPI PSRAM must be enabled in the board options. Otherwise
  `createSprite` returns null and the first draw crashes.
- To cut memory use by roughly four times, call `canvas.setColorDepth(4)`
  before `createSprite` and set a palette. Sixteen colours is more than the E6
  can show. This is optional; the default 16-bit depth is fine given the PSRAM.

## Power and retention

- The panel keeps its image with no power. Refresh only when something has
  actually changed; every refresh costs seconds and battery.
- `cfg.clear_display = false` keeps the last image instead of wiping it on
  boot.
- Always call `waitDisplay()` before `M5.Power.timerSleep()` or `powerOff()`.

## Checklist

1. `cfg.clear_display = false` in `M5.config()`.
2. `setEpdMode(epd_fastest)`, or `epd_quality` for a final image.
3. `createSprite(w, h)` once, in `setup`.
4. Per frame: `fillSprite`, draw, `pushSprite(0, 0)`, `waitDisplay`.
5. Redraw only on a real change.
