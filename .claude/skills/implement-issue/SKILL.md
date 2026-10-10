---
name: implement-issue
description: Průvodce implementací issue - od přečtení zadání v docs/issues/ přes vytvoření branch, implementaci kódu, commit a merge do main až po archivaci do docs/issues/done/ a uzavření issue na GitHubu
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

Zeptej se mě na název nebo číslo issue (číslo z GitHubu). Počkej na mou odpověď.

## KROK 2: Potvrzení porozumění

Poté, co ti poskytnu issue:

1. Přečti soubor issue z `docs/issues/` (aktivní issue; `NNN` = číslo z GitHubu).
   Pokud tam není, zkus `docs/issues/done/` — může jít o už uzavřené issue.
   Stav na GitHubu: `gh issue view <N> --repo cassandragargoyle/portunix-agent`
2. Přeformuluj issue vlastními slovy
3. Vysvětli potřebné změny v kódu
4. Požádej mě o potvrzení

**STOP** - Počkej na mé výslovné potvrzení než budeš pokračovat.

## KROK 3: Vytvoř branch

Vytvoř git branch: `feature/issue-<N>-<kratky-popis>` nebo `fix/issue-<N>-<kratky-popis>`

- Použij malá písmena a pomlčky pro popis
- Příklad: `feature/issue-42-add-export-api`
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

## KROK 7: Aktualizace a archivace issue

1. Uprav soubor issue v `docs/issues/`:
   - Změň `**Status**: Open` na `**Status**: Implemented`
   - Přidej datum uzavření

2. Archivuj soubor do `done/` (zachová git historii):

   ```bash
   git mv docs/issues/<NNN>-<nazev>.md docs/issues/done/<NNN>-<nazev>.md
   ```

3. Zavři GitHub issue (GitHub je seznam issue, žádné README tabulky nejsou):

   ```bash
   gh issue close <N> --repo cassandragargoyle/portunix-agent --comment "Implemented in <commit/verze>."
   ```

4. Commitni změnu stavu issue společně s přesunem:

   ```text
   docs(<issue-id>): archive - implementation complete
   ```
