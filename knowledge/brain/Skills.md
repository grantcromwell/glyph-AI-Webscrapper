# Skills

## EE Spells (grimoire/wild/)
- `pcb-design` — KiCad projects, schematics, Gerber, BOM
- `firmware-flash` — OpenOCD flash/read/erase/verify
- `logic-analyze` — Sigrok capture + protocol decode
- `spice-sim` — Ngspice netlists + simulation
- `gpio-control` — GPIO, SPI, I2C, PWM, ADC
- `gpu-compute` — CUDA/Vulkan/PyTorch acceleration
- `can-bus` — SocketCAN sniff/send/filter
- `serial-debug` — pyserial list/open/read/write/monitor
- `edit` — Read/write/patch files
- `autonomy-loop` — Assess → plan → act → report

## Casting Spells
```bash
echo '{"args":{"action":"assess"}}' | go run ./grimoire/wild/autonomy-loop
echo '{"args":{"action":"flash","target":"stm32f4","file":"firmware.bin"}}' | go run ./grimoire/wild/firmware-flash
```
