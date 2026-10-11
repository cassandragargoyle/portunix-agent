---
name: finish-branch
description: Dokončení práce na feature branch - merge do hlavní větve, úklid branch a volitelná archivace issue do docs/issues/done/ s uzavřením na GitHubu
---

Dokončení práce na feature branch - merge do hlavní větve, úklid branch a volitelná aktualizace stavu issue.

@.claude\prompts\assistant-programming.md

**DŮLEŽITÉ: NIKDY nepřidávej "Co-Authored-By: Claude <noreply@anthropic.com>" do kódových souborů ani git commitů.**

## KROK 1: Zjisti kontext

Zeptej se mě na:

1. Název feature branch (nebo použij aktuální branch)
2. Cílová větev pro merge (např. main, develop)

Ověř, že branch existuje a obsahuje commity.

## KROK 1.5: Kontrola CHANGELOG.md

Ověř, že branch přidává záznam do `CHANGELOG.md` pod `## [Unreleased]`:

```bash
git diff <cílová-větev>...HEAD -- CHANGELOG.md
```

- Záznam existuje → pokračuj.
- Záznam chybí a změna je viditelná pro uživatele → navrhni text záznamu podle
  `docs/contributing/CHANGELOG-GUIDELINES.md`, po schválení ho zapiš a commitni
  (`docs(<issue-id>): update changelog`) ještě před merge.
- Změna není viditelná pro uživatele (testy, interní refaktoring, archivace issue) → záznam není potřeba.

**STOP** - Pokud záznam chybí, počkej na rozhodnutí uživatele.

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

### Pokud NE

Přeskoč aktualizaci stavu. Issue zůstává otevřené na GitHubu i v `docs/issues/` pro další práci.

## KROK 4: Shrnutí

Zobraz shrnutí:

- Merge provedený do větve: `<název>`
- Feature branch smazána: ano/ne
- Záznam v `CHANGELOG.md`: ano/není potřeba
- Issue archivováno a zavřeno na GitHubu: ano/ne/přeskočeno
- Další kroky (pokud jsou)
