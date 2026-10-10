---
name: deploy-plugin
description: Lokální build a nasazení ptx-agent jako helper-mode pluginu do běžícího systému Portunix
---

# Nasazení ptx-agent pluginu do Portunix

Lokální build a nasazení ptx-agent jako helper-mode pluginu do běžícího systému Portunix.

**Argument:** $ARGUMENTS

## Instrukce

### KROK 1: Kontrola předpokladů

1. Ověř, že Portunix je dostupný v PATH:

   ```bash
   portunix --version
   ```

2. Pokud není dostupný, zobraz chybu a **STOP**

### KROK 2: Build pluginu

1. Spusť build distribučního balíčku:

   ```bash
   make dist
   ```

2. Ověř, že vznikl adresář `dist/ptx-agent/` obsahující:
   - `ptx-agent` (nebo `ptx-agent.exe` na Windows) — binárka
   - `plugin.json` — manifest pluginu (type: helper, runtime: native)

3. Pokud build selže, zobraz chybu a **STOP**

### KROK 3: Odinstalace předchozí verze (pokud existuje)

Zkontroluj, zda je plugin již nainstalován:

```bash
portunix plugin list | grep '^agent '
```

**Pokud existuje**, odinstaluj:

```bash
echo 'y' | portunix plugin uninstall agent
```

**Pro Windows** (v Git Bash):

```powershell
powershell.exe -Command "echo 'y' | portunix plugin uninstall agent"
```

### KROK 4: Instalace pluginu

```bash
portunix plugin install dist/ptx-agent
```

**Pro Windows** (v Git Bash):

```powershell
powershell.exe -Command "portunix plugin install dist/ptx-agent"
```

### KROK 5: Povolení pluginu

```bash
portunix plugin enable agent
```

### KROK 6: Ověření

1. Ověř instalaci:

   ```bash
   portunix plugin list
   ```

2. Ověř funkčnost přes Portunix dispatch:

   ```bash
   portunix agent files list /tmp --allowed-dirs /tmp
   ```

3. Zobraz výsledek uživateli

### KROK 7: Shrnutí

Zobraz shrnutí:

```text
Plugin 'agent' (binárka ptx-agent) nasazen (helper mode, native runtime).

Build: make dist
Zdroj: dist/ptx-agent/plugin.json
Instalace: portunix plugin install dist/ptx-agent
Registry: ~/.portunix/plugins/registry.json

Pro ověření funkčnosti:
  portunix plugin list
  portunix agent files list /tmp --allowed-dirs /tmp
  portunix agent mcp serve --help
```

## Příklady použití

```bash
# Standardní nasazení
/deploy-plugin
```

## Poznámky

- Tento příkaz je specifický pro projekt portunix-agent
- Vyžaduje funkční Portunix v PATH
- Plugin typ: `helper` (stdio/argv subprocess dispatch)
- Runtime: `native` (Go compiled binary)
- `make dist` musí vytvořit adresář `dist/ptx-agent/` s binárkou a `plugin.json`
- Plugin je v Portunix registrován pod jménem `agent` (pole `name` v `plugin.json`), binárka se jmenuje `ptx-agent`
- `portunix plugin install` očekává adresář s manifestem, ne cestu k `plugin.json`
- `files list` vyžaduje příznak `--allowed-dirs`, bez něj skončí chybou
- Plugin se registruje do `~/.portunix/plugins/registry.json`
- Po instalaci je plugin dostupný přes `portunix agent <command>`
