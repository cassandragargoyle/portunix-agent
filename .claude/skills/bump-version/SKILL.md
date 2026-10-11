---
name: bump-version
description: Zvýšení verze ptx-agent podle SemVer - určení typu (major/minor/patch z argumentu nebo z commitů), úprava plugin.json, převod sekce [Unreleased] v CHANGELOG.md na datovanou verzi, commit a tag v pevném pořadí (commit → tag). Spouštěče - "zvyš verzi", "bumpni verzi", "nová verze", "bump version", "bump patch/minor/major".
---

# Bump Version

Zvýšení verze `ptx-agent` podle [SemVer 2.0.0](https://semver.org/) s úpravou `plugin.json`
a `CHANGELOG.md`. Pravidla pro changelog: `docs/contributing/CHANGELOG-GUIDELINES.md`.

**DŮLEŽITÉ: NIKDY nepřidávej "Co-Authored-By: Claude <noreply@anthropic.com>" do kódových souborů ani git commitů.**

Komunikuj česky. Obsah `CHANGELOG.md`, commit message a technické termíny zůstávají anglicky.

## PARAMETRY

Argument `$ARGUMENTS` určuje, které číslo ve verzi `MAJOR.MINOR.PATCH` se zvýší:

| Hodnota | Pozice | Význam |
| ------- | ------ | ------ |
| `major` \| `1` \| `první` | 1. číslo | Nekompatibilní změny (CLI, MCP tools, veřejné Go API v `src/pkg/`) |
| `minor` \| `2` \| `druhé` | 2. číslo | Nové funkce, zpětně kompatibilní |
| `patch` \| `3` \| `třetí` | 3. číslo | Opravy chyb |

Pokud je `$ARGUMENTS` **prázdný**, navrhni typ podle commitů (KROK 2b).

## KROK 1: Zjisti aktuální verzi

Zdrojem verze je `plugin.json` (`make` a `scripts/make-release.py` ji předávají do binárky):

```bash
grep '"version"' plugin.json
git tag -l "v*" --sort=-v:refname | head -1
```

Pokud se poslední tag a `plugin.json` liší, nebo žádný tag neexistuje, upozorni na to uživatele.

Ověř také, že pracovní strom je čistý (`git status --short`) a že jsi na větvi `main`.
Verze se zvyšuje na `main`, ne na feature branch.

## KROK 2: Urči typ bumpu

### 2a) `$ARGUMENTS` je zadán

Převeď parametr na typ podle tabulky výše.

### 2b) `$ARGUMENTS` není zadán

```bash
LAST_TAG=$(git tag -l "v*" --sort=-v:refname | head -1)
git log ${LAST_TAG:+$LAST_TAG..}HEAD --format="%h %s"
```

- Obsahuje `BREAKING CHANGE` nebo `!:` → **major**
- Obsahuje `feat(` nebo `feat:` → **minor**
- Jen `fix`, `chore`, `docs`, `refactor`, `perf`, `test`, `build` → **patch**

Dokud je verze `0.y.z`, nekompatibilní změna zvyšuje **minor**, ne major.

**STOP** - Zobraz aktuální verzi, seznam commitů od posledního tagu, navržený typ se zdůvodněním
a novou verzi. Počkej na potvrzení nebo opravu.

## KROK 3: Uprav plugin.json

Nástrojem `Edit` změň pouze řádek `"version": "X.Y.Z"` v `plugin.json`. Zachovej tvar
`"version": "X.Y.Z"` - `Makefile` ho čte textově.

Ověř build a verzi v binárce:

```bash
make build
./ptx-agent --version
```

Výstup musí být `ptx-agent version X.Y.Z`.

## KROK 4: Uprav CHANGELOG.md

### 4.1 Porovnej [Unreleased] s commity

Záznamy vznikají průběžně na feature branch (viz `CHANGELOG-GUIDELINES.md`). Tady se jen
kontroluje úplnost:

```bash
git log ${LAST_TAG:+$LAST_TAG..}HEAD --format="%h %s"
```

Pro každý `feat` / `fix` commit a každou změnu viditelnou pro uživatele ověř, že má záznam pod
`## [Unreleased]`. Chybějící záznamy doplň podle tabulky:

| Conventional commit | Sekce |
| ------------------- | ----- |
| `feat` | `### Added in X.Y.Z` |
| `fix` | `### Fixed in X.Y.Z` |
| `refactor`, `perf`, `chore`, `build`, `docs` (viditelné pro uživatele) | `### Changed in X.Y.Z` |
| `BREAKING CHANGE` / `!:` | `### Changed in X.Y.Z` s poznámkou **BREAKING** |
| odstranění funkce | `### Removed in X.Y.Z` |
| bezpečnostní oprava | `### Security in X.Y.Z` |

Commity `test`, archivace issue a interní refaktoring záznam nemají. Prázdné sekce nepiš.

### 4.2 Převeď [Unreleased] na verzi

- `## [Unreleased]` → `## [X.Y.Z] - YYYY-MM-DD` (datum z `date -I`)
- `### Added in Unreleased` → `### Added in X.Y.Z` (stejně `Fixed`, `Changed`, ...)
- Nad novou sekci vlož prázdný nadpis `## [Unreleased]`

Pravidla zápisu: anglicky, na začátku **tučně** název funkce nebo oblasti, pomlčka, vysvětlení
dopadu pro uživatele, číslo issue `(Issue #N)` a ADR `(ADR-NNN)`, pokud existují.

**STOP** - Před zápisem zobraz uživateli navrženou sekci a počkej na připomínky.

Po schválení zapiš změnu a zkontroluj formát:

```bash
npx markdownlint-cli2 CHANGELOG.md
```

## KROK 5: Commit

```bash
git add plugin.json CHANGELOG.md
git commit -m "chore: bump version to vX.Y.Z"
```

**STOP** - Požádej o potvrzení commitu.

## KROK 6: Tag (až po commitu)

Tag musí ukazovat na commit, který obsahuje nový `plugin.json` a `CHANGELOG.md`:

```bash
git status --short
git log -1 --oneline
git tag -a vX.Y.Z -m "Release vX.Y.Z"
```

**STOP** - Push commitu a tagu (`git push origin main vX.Y.Z`) proveď jen po výslovném potvrzení.

## KROK 7: Shrnutí

- Původní a nová verze
- Sekce `CHANGELOG.md`, která vznikla
- Commit a tag: vytvořeno / pushnuto
- Další krok: release balíček (`python scripts/make-release.py`) nebo nasazení (`/deploy-plugin`)

## POZNÁMKY

- Pořadí je pevné: `plugin.json` → `CHANGELOG.md` → commit → tag. Nikdy netaguj před commitem.
- Tagy mají prefix `v` (`v0.2.0`), nadpisy v `CHANGELOG.md` ne (`[0.2.0]`).
- Datum vždy ve tvaru `YYYY-MM-DD`.
