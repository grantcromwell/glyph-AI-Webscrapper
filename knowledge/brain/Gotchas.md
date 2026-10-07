# Gotchas

- The edit spell's `out` variable was scoped inside `if` blocks — fixed by using named variables
- bridge spell has pre-existing compile errors (undefined json, strings) — not our problem
- DDG returns empty results for narrow technical queries — use curl + pdftotext for datasheets
- The pronounce spell's `extractShort()` cannot detect language quality — engram hygiene matters
- CAN sockets available in Python but candump/cansend tools not installed — use python-can
- sudo not available from this session — install tools via pacman manually
