---
name: use-case
description: Průvodce evidencí use-casů - od popisu přes kontrolu duplicit, vytvoření UC souboru, vytvoření issue až po commitnutí
---

Průvodce evidencí use-casů — od popisu přes kontrolu duplicit, vytvoření UC souboru, vytvoření issue až po commitnutí.

## KROK 0: Kontrola role

Zkontroluj aktuální roli v `CLAUDE.local.md`. Pokud role není **Developer** nebo **Architect** nebo **Product Owner**:

1. Informuj uživatele, že evidence use-casů vyžaduje jednu z rolí: Developer, Architect, Product Owner
2. Nabídni přepnutí role příkazem `/role`
3. **STOP** - Nepokračuj dokud není vhodná role aktivní

## KROK 1: Získej popis use-casu

Pokud uživatel poskytl popis jako argument příkazu, použij ho. Jinak se zeptej:

1. Název use-casu (krátký, výstižný)
2. Popis — kdo je actor, co chce provést a proč

**STOP** - Počkej na odpověď uživatele (pokud nebyl popis zadán jako argument).

## KROK 2: Kontrola duplicit

1. Přečti `docs/use-cases/README.md` — zkontroluj index existujících UC
2. Přečti existující UC soubory v `docs/use-cases/` (pokud existují)
3. Porovnej nový UC s existujícími — hledej překryvy v účelu, actorovi a hlavním flow

Pokud existuje potenciální duplicita:

- Informuj uživatele: "Nalezen podobný UC: UC-{NNN} — {Title}. Popis podobnosti."
- Zeptej se zda pokračovat nebo upravit
- **STOP** - Počkej na rozhodnutí uživatele

Pokud duplicita neexistuje, pokračuj automaticky.

## KROK 3: Eviduj use-case

### 3a: Urči číslo UC

Zjisti nejvyšší existující číslo UC z `docs/use-cases/README.md` indexu. Nový UC dostane číslo o 1 vyšší. Pokud žádné UC neexistují, začni od 001.

### 3b: Vytvoř UC soubor

Vytvoř soubor `docs/use-cases/UC-{NNN}-{slug}.md` s následující šablonou:

```markdown
# UC-{NNN} — {Title}

**Status**: 📋 New
**Related Issue**: —
**Created**: {YYYY-MM-DD}

## Actor
{kdo provádí akci}

## Preconditions
- {co musí platit předem}

## Main Flow
1. {krok 1}
2. {krok 2}

## Alternative Flows
- {alternativní scénáře, chybové stavy}

## Postconditions
- {co platí po úspěšném provedení}

## Acceptance Criteria
1. [ ] {kritérium 1}
2. [ ] {kritérium 2}

## Notes
{dodatečné poznámky, nebo prázdné}
```

Vyplň šablonu na základě popisu od uživatele. Kde informace chybí, použij rozumné výchozí hodnoty nebo označ jako `TBD`.

### 3c: Aktualizuj UC index

Přidej řádek do tabulky v `docs/use-cases/README.md`:

```text
| [UC-{NNN}](UC-{NNN}-{slug}.md) | {Title} | 📋 New | — |
```

Zobraz uživateli vytvořený UC soubor a zeptej se na potvrzení:

- a) Přijmout tak jak je
- b) Upravit (uživatel specifikuje co změnit)
- c) Zrušit

**STOP** - Počkej na schválení.

## KROK 4: Realizace — vytvoř issue

Na základě UC vytvoř issue pro implementaci:

### 4a: Vytvoř GitHub issue (číslo přidělí GitHub)

Postupuj podle skillu `/create-issue` (kontrola duplicit, kontrola citlivého obsahu, `gh issue create`).
Číslo `N` vezmi z adresy, kterou `gh` vypíše; `NNN` je `N` doplněné nulami na tři číslice.
Číslo nikdy neodvozuj z existujících souborů.

### 4b: Vytvoř issue soubor

Vytvoř soubor `docs/issues/{NNN}-{slug}.md`:

```markdown
# Issue #{N}: {Title from UC}

**Status**: Open
**Priority**: Medium
**Type**: Enhancement
**Created**: {YYYY-MM-DD}
**GitHub**: #{N}
**Related UC**: [UC-{UC-NNN}](../use-cases/UC-{UC-NNN}-{slug}.md)

## Feature Description

Implementation of use case UC-{UC-NNN} — {UC Title}.

{popis odvozený z UC main flow a acceptance criteria}

## Acceptance Criteria

{zkopíruj acceptance criteria z UC}

## Complexity Estimate

{odhad na základě rozsahu UC: Low / Medium / High}
```

### 4c: Odkaž detailní soubor z GitHubu

Doplň do těla GitHub issue odkaz na detailní soubor (viz KROK 6 skillu `/create-issue`).
Žádný index issue se neudržuje.

### 4d: Aktualizuj UC soubor

V UC souboru změň:

- `**Status**: 📋 New` na `**Status**: ✅ Covered by Issue`
- `**Related Issue**: —` na `**Related Issue**: [#{N}](../issues/{NNN}-{slug}.md)`

Aktualizuj i status v UC indexu (`docs/use-cases/README.md`).

### 4e: Commitni

Commitni všechny nové a upravené soubory:

```text
docs(UC-{UC-NNN}): add use case and implementation issue #{N}
```

**STOP** - Zeptej se na potvrzení commit message před commitnutím.

## KROK 5: Shrnutí

Zobraz shrnutí:

- Vytvořený UC: `docs/use-cases/UC-{NNN}-{slug}.md`
- Vytvořený issue: GitHub `#{N}` a `docs/issues/{NNN}-{slug}.md`
- Aktualizovaný index: `docs/use-cases/README.md`
- Další kroky: implementace issue pomocí `/implement-issue`
