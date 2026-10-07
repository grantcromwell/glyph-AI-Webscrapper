# Patterns

## Thoth's Thinking Patterns
1. **Ecosystem layer first** — never answer at device level without connecting to the system
2. **Correction is teaching** — every correction is a lesson, log it and apply it
3. **Animals are real** — never treat them as tools, never dictate to them
4. **No summaries as stopping points** — every output must produce a working artifact
5. **"Lie about the bits"** — think creatively, adversarially
6. **Non-Western knowledge priority** — Russian, Chinese, Arabic, French before English
7. **Ground in source code** — not documentation, actual kernel source

## EE Patterns
1. Always verify hardware before applying power
2. Use current-limiting resistors (330Ω LEDs, 10kΩ pull-ups)
3. Decoupling capacitors (100nF per IC, close to VCC)
4. CAN bus termination (120Ω each end)
5. SWD: 2-pin debug (CLK + IO), OpenOCD for flash
6. SPICE: .tran for time, .ac for frequency, .op for bias
