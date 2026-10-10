---
name: create-issue
description: Založení nového issue GitHub-first - nejdřív vznikne issue na GitHubu (číslo přiděluje GitHub), potom detailní popis v docs/issues/NNN-*.md a odkaz na něj z GitHubu
argument-hint: [krátký popis issue]
---

Založení nového issue GitHub-first — číslo přiděluje GitHub, detailní popis je v `docs/issues/NNN-*.md`.

**Argument:** $ARGUMENTS

## Princip

**GitHub je jediný zdroj pravdy pro seznam issue.** Číslo přiděluje GitHub. GitHub issue obsahuje jen
základ (název, krátký popis, odkaz na detailní soubor). Detaily jsou v `docs/issues/NNN-*.md`, kde
`NNN` je číslo z GitHubu doplněné nulami na tři číslice.

**Žádné přehledové tabulky, žádný README index, žádný mapovací soubor.**

Úplná metodika: `docs/contributing/ISSUE-MANAGEMENT.md`.

## KROK 1: Získej vstupy

Z argumentu nebo kontextu (chybějící údaje zjisti dotazem):

- **Název** — krátký, výstižný, anglicky, bez čísla
- **Slug** — kebab-case pro název souboru (odvozený z názvu)
- **Typ (label)**: `bug` | `enhancement` | `documentation`
- **Priorita**: Critical | High | Medium | Low
- **Krátký popis** (1–3 věty) — půjde do těla GitHub issue
- **Detailní obsah** — problém, akceptační kritéria, testování — půjde do `.md`

Nejdřív zkontroluj duplicity:

```bash
gh issue list --repo cassandragargoyle/portunix-agent --state all --search "<klíčová slova>"
```

## KROK 2: Kontrola citlivého obsahu (POVINNÉ)

`docs/issues/` i GitHub issue jsou **veřejné**. Než cokoli zapíšeš, ověř, že tělo GitHub issue ani
detailní soubor neobsahují:

- jména zákazníků, obchodní vztahy, zakázky, priority dané zákazníkem
- názvy projektů třetích stran nebo zaměstnavatele a jejich interní identifikátory
- lokální cesty (`/home/<user>/…`, `/media/…`) — použij cesty relativní ke kořeni repozitáře
- interní názvy strojů, IP adresy, přihlašovací údaje, tokeny

Jakýkoli nález oprav před pokračováním, nebo se zeptej uživatele.

## KROK 3: Kontrola `gh`

```bash
gh --version && gh auth status
```

Pokud není přihlášen, zastav a požádej uživatele o `! gh auth login`.

## KROK 4: Vytvoř GitHub issue (číslo přidělí GitHub)

```bash
REPO="cassandragargoyle/portunix-agent"
gh issue create --repo "$REPO" \
  --title "<NÁZEV>" \
  --body "<KRÁTKÝ POPIS>

Full details: docs/issues/<NNN>-<slug>.md" \
  --label "<typ>"
```

Číslo **`N`** vezmi z vypsané adresy (`.../issues/31` → `N=31`, `NNN=031`).
Číslo nikdy nevymýšlej — issue a pull requesty sdílejí jedno počítadlo.

## KROK 5: Napiš detailní soubor

Vytvoř `docs/issues/<NNN>-<slug>.md`:

```markdown
# Issue #<N>: <Title>

**Status**: Open
**Priority**: <priority>
**Type**: <Bug | Enhancement | Documentation>
**Created**: <YYYY-MM-DD>
**GitHub**: #<N>
**Related**: #<...>

## Summary

## Problem Description

## Acceptance Criteria

1. ...

## Testing
```

Bez emoji v nadpisech a v poli Status (viz `ISSUE-MANAGEMENT.md`). Obsah souboru je anglicky.

## KROK 6: Odkaž detailní soubor z GitHubu

```bash
URL="https://github.com/$REPO/blob/main/docs/issues/<NNN>-<slug>.md"
gh issue edit <N> --repo "$REPO" --body "<KRÁTKÝ POPIS>

[Full details](${URL})"
```

Odkaz začne fungovat, až se soubor dostane do `main` na GitHubu.

## KROK 7: Shrnutí

```text
Issue #<N> vytvořeno.

GitHub:  https://github.com/cassandragargoyle/portunix-agent/issues/<N>
Detail:  docs/issues/<NNN>-<slug>.md
Label:   <typ>

Další kroky:
  git checkout -b feature/issue-<N>-<slug>
  commity: feat(<N>): …   (uzavření: closes #<N>)
```

## Poznámky

- Archivace hotového issue: `git mv docs/issues/<NNN>-*.md docs/issues/done/` a `gh issue close <N>`
  (dělá `/implement-issue` nebo `/finish-branch`).
- Detailní soubor necommituj bez pokynu uživatele.
