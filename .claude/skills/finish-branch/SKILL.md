---
name: finish-branch
description: Dokončení práce na feature branch - merge do hlavní větve, úklid branch a volitelná aktualizace stavu issue
---

Dokončení práce na feature branch - merge do hlavní větve, úklid branch a volitelná aktualizace stavu issue.

@.claude\prompts\assistant-programming.md

**DŮLEŽITÉ: NIKDY nepřidávej "Co-Authored-By: Claude <noreply@anthropic.com>" do kódových souborů ani git commitů.**

## KROK 1: Zjisti kontext

Zeptej se mě na:

1. Název feature branch (nebo použij aktuální branch)
2. Cílová větev pro merge (např. main, develop)

Ověř, že branch existuje a obsahuje commity.

## KROK 2: Merge a úklid

> **Co je PR (Pull Request)?** Žádost o začlenění změn z feature branch do hlavní větve. Používá se pro code review v týmech. Pro solo vývoj lze mergovat přímo.

Zeptej se uživatele na preferovaný postup:

### Varianta A: Přímý merge (výchozí pro solo vývoj)

Zjisti zdrojovou větev (ze které feature branch vznikla) a proveď merge:

```bash
git checkout <zdrojová-větev>    # např. main, develop, release/v1.0
git merge feature/<issue-id>-<popis>
git branch -d feature/<issue-id>-<popis>
```

**STOP** - Počkej na potvrzení před provedením merge.

### Varianta B: Pull Request (pro týmovou spolupráci)

Připrav PR popis:

- **Summary**: Co tento PR dělá
- **Motivation**: Proč (odkaz na issue)
- **Changes**: Seznam změn
- **Testing**: Kroky pro ověření

Vytvoř PR pomocí `gh pr create`.

## KROK 3: Aktualizace issue (volitelné)

Zeptej se uživatele: "Je tímto issue kompletně dokončeno?"

### Pokud ANO

Uprav soubor issue v `docs/issues/internal/`:

- Změň `**Status:** 📋 Open` na `**Status:** ✅ Implemented`
- Přidej datum uzavření

Aktualizuj také `docs/issues/README.md` - změň status v tabulce.

Commitni změnu stavu issue:

```text
docs(<issue-id>): Close issue - implementation complete
```

### Pokud NE

Přeskoč aktualizaci stavu. Issue zůstává otevřené pro další práci.

## KROK 4: Shrnutí

Zobraz shrnutí:

- Merge provedený do větve: `<název>`
- Feature branch smazána: ano/ne
- Issue status aktualizován: ano/ne/přeskočeno
- Další kroky (pokud jsou)
