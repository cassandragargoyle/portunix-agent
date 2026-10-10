---
name: implement-issue
description: Průvodce implementací issue - od přečtení zadání přes vytvoření branch, implementaci kódu, commit až po merge do main a uzavření issue
---

Průvodce implementací issue - od přečtení zadání přes vytvoření branch, implementaci kódu, commit až po merge do main a uzavření issue.

@.claude\prompts\assistant-programming.md

<!-- Výše uvedený soubor definuje základní kontext (jazyk, role). Tento soubor přidává workflow. -->

## KROK 0: Kontrola role

Zkontroluj aktuální roli v `CLAUDE.local.md`. Pokud role není **Developer**:

1. Informuj uživatele, že implementace vyžaduje roli Developer
2. Nabídni přepnutí role příkazem `/role developer`
3. **STOP** - Nepokračuj dokud není role Developer aktivní

## KROK 1: Získej issue

Zeptej se mě na název nebo ID interního issue. Počkej na mou odpověď.

## KROK 2: Potvrzení porozumění

Poté, co ti poskytnu issue:

1. Přečti soubor issue z `docs/issues/internal/`
2. Přeformuluj issue vlastními slovy
3. Vysvětli potřebné změny v kódu
4. Požádej mě o potvrzení

**STOP** - Počkej na mé výslovné potvrzení než budeš pokračovat.

## KROK 3: Vytvoř branch

Vytvoř git branch: `feature/<issue-id>-<kratky-popis>` nebo `fix/<issue-id>-<kratky-popis>`

- Použij malá písmena a pomlčky pro popis
- Příklad: `feature/TA042-add-export-api`
Zobraz git příkaz před jeho spuštěním.

## KROK 4: Implementace

- Proveď všechny změny kódu v jedné souvislé fázi
- NEPTEJ se na otázky, pokud to není skutečně nejednoznačné
- Modifikuj pouze relevantní soubory
- Dodržuj coding style projektu

Na konci poskytni shrnutí:

- Modifikované soubory
- Přidané metody/třídy
- Odstraněný kód
- Klíčová rozhodnutí

Zeptej se mě, co chci:

- a) Přijmout tak jak je
- b) Upravit konkrétní části
- c) Předělat kompletně

**STOP** - Počkej na mé schválení než budeš pokračovat.

## KROK 5: Commit

Připrav commit message:

```text
<type>(<issue-id>): <krátké shrnutí>

- klíčová změna 1
- klíčová změna 2
```

Povolené typy: `feat`, `fix`, `refactor`, `docs`, `test`, `chore`

**STOP** - Zeptej se mě na potvrzení commit message před commitnutím.

## KROK 6: Merge a úklid

> **Co je PR (Pull Request)?** Žádost o začlenění změn z feature branch do hlavní větve. Používá se pro code review v týmech. Pro solo vývoj lze mergovat přímo.

Zeptej se uživatele na preferovaný postup:

### Varianta A: Přímý merge (výchozí pro solo vývoj)

Zjisti zdrojovou větev (ze které feature branch vznikla) a proveď merge:

```bash
git checkout <zdrojová-větev>    # např. main, develop, release/v1.0
git merge feature/<issue-id>-<popis>
git branch -d feature/<issue-id>-<popis>
```

### Varianta B: Pull Request (pro týmovou spolupráci)

Připrav PR popis:

- **Summary**: Co tento PR dělá
- **Motivation**: Proč (odkaz na issue)
- **Changes**: Seznam změn
- **Testing**: Kroky pro ověření

## KROK 7: Aktualizace issue

Uprav soubor issue v `docs/issues/internal/`:

- Změň `**Status:** 📋 Open` na `**Status:** ✅ Implemented`
- Aktualizuj status v `docs/issues/README.md`

Commitni změnu stavu issue.
