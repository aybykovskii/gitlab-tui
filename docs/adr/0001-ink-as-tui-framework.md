# ADR 0001: Ink как TUI-фреймворк

**Status:** Accepted

## Context

Нужен TUI-фреймворк для Node.js/TypeScript. Альтернативы: `blessed`/`neo-blessed` (низкоуровневый box-based API, слабая поддержка), `terminal-kit` (мало распространён), нативный `readline` (слишком низкий уровень для сложного UI).

## Decision

Используем **Ink** (React для терминала).

## Rationale

- Компонентная модель знакома разработчикам React — снижает порог входа для контрибьюторов
- Активно поддерживается, богатая экосистема готовых компонентов (`ink-select-input`, `ink-text-input`, `ink-table`)
- Flexbox-лейаут (`<Box flexDirection="row">`) позволяет реализовать side-by-side diff без велосипедов
- TypeScript-типы из коробки

## Trade-offs

`blessed` даёт больше низкоуровневого контроля над терминалом, но его API устарел и поддержка слабая. Go-based инструменты (bubbletea) производительнее, но требуют смены языка — проект намеренно остаётся в Node.js экосистеме.
