# portunix-agent

[English](README.md) | **Čeština**

Plugin pro systém [Portunix](https://github.com/CassandraGargoyle/portunix), který umožňuje vzdálené ovládání zařízení, na kterých je nainstalován. Nabízí simulaci klávesnicového vstupu, správu oken, inspekci aplikací přes accessibility API, čtení terminálu VS Code, automatizaci webového prohlížeče pomocí Playwrightu a bezpečné procházení souborů.

Všechny funkce jsou dostupné jako příkazy CLI i jako nástroje [MCP](https://modelcontextprotocol.io/), takže AI asistenti mohou na cílovém stroji ovládat plochu, pracovat s aplikacemi a automatizovat webové prohlížeče. Agent pokrývá dvě vzájemně se doplňující oblasti: **desktopové aplikace** (přes nativní accessibility API dané platformy) a **webové aplikace** (přes automatizaci prohlížeče pomocí Playwrightu). AI asistenti tak získávají jednotný způsob, jak pracovat s jakýmkoli softwarem běžícím na zařízení.

## Funkce

### Ovládání aplikací přes accessibility API

Inspekce libovolné přístupné aplikace na ploše pomocí nativních accessibility technologií dané platformy (AT-SPI2 na Linuxu, UI Automation na Windows).

```bash
ptx-agent app list                      # Vypíše okna na ploše
ptx-agent app tree <name> [--depth N]   # Vypíše accessibility strom
ptx-agent app read <name> [--depth N]   # Přečte textový obsah aplikace
```

### Čtení terminálu VS Code

Čte obsah terminálu ve VS Code tak, že projde jeho accessibility strom až k prvkům terminálu xterm.js.

```bash
ptx-agent vscode status                 # Zkontroluje stav accessibility ve VS Code
ptx-agent vscode terminal list          # Vypíše otevřené terminály
ptx-agent vscode terminal read [--index N]  # Přečte obsah terminálu
ptx-agent vscode tree                   # Vypíše accessibility strom VS Code
```

Na Linuxu vyžaduje přepínač `--force-renderer-accessibility` nebo volbu `editor.accessibilitySupport: "on"` v nastavení VS Code. Windows obvykle zapne accessibility automaticky, jakmile se připojí UIA klient.

### Simulace klávesnicového vstupu (Linux)

Psaní textu, odesílání klávesových zkratek nebo vkládání ze schránky s podporou Waylandu i X11.

```bash
ptx-agent keyboard type <text> [--delay MS] [--no-indent] [--window NAME]
ptx-agent keyboard key <combo> [--window NAME]
ptx-agent keyboard paste <text> [--window NAME]
ptx-agent keyboard read                 # Přečte obsah schránky
```

Backendy: **uinput** (Wayland + X11) se záložním xdotool (pouze X11). Přepínač `--window` před vstupem aktivuje cílové okno. Příkaz `paste` používá schránku, takže nezávisí na rozložení klávesnice (funguje s českým, německým i jiným rozložením).

### Správa oken (Linux)

Výpis, vyhledávání a aktivace oken na ploše.

```bash
ptx-agent windows list                  # Vypíše všechna okna
ptx-agent windows find <query>          # Vyhledá okna podle názvu
ptx-agent windows focus <window-id>     # Aktivuje okno
```

Backendy: **KWin** (Wayland/KDE přes skriptovací API D-Bus) se záložním xdotool (pouze X11).

### Ovládání prohlížeče přes Playwright

Automatizace webových prohlížečů v režimu headless (na pozadí) nebo headed (interaktivně). Umožňuje navigaci po stránkách, práci s prvky podle číselného indexu, pořizování snímků obrazovky, extrakci obsahu a správu záložek. Všechny akce jsou dostupné také jako nástroje MCP pro integraci s AI asistenty.

```bash
ptx-agent browser install                  # Nainstaluje binárky prohlížečů přes Portunix
ptx-agent browser launch [--headless]      # Spustí instanci prohlížeče
ptx-agent browser goto <url>               # Přejde na URL
ptx-agent browser elements                 # Vypíše interaktivní prvky s indexy
ptx-agent browser click <index>            # Klikne na prvek podle indexu
ptx-agent browser type <index> <text>      # Napíše text do pole formuláře
ptx-agent browser screenshot [--full]      # Pořídí snímek obrazovky
ptx-agent browser content [--selector S]   # Extrahuje text/HTML stránky
ptx-agent browser tabs                     # Vypíše otevřené záložky
ptx-agent browser tab new [url]            # Otevře novou záložku
ptx-agent browser close                    # Zavře prohlížeč
```

Vyžaduje binárky prohlížečů pro Playwright nainstalované příkazem `ptx-agent browser install`. Podrobnosti o architektuře jsou v [ADR-002](docs/adr/002-browser-control-playwright.md).

### Bezpečné procházení souborů

Procházení adresářů s řízením přístupu pomocí whitelistu a s ochranou proti průchodu přes symbolické odkazy.

```bash
ptx-agent files list <path> --allowed-dirs=/home/user,/tmp
ptx-agent files info <path> --allowed-dirs=/home/user,/tmp
```

## Podpora platforem

| Funkce | Linux (X11) | Linux (Wayland/KDE) | Windows |
|---|:---:|:---:|:---:|
| Accessibility aplikací | AT-SPI2 | AT-SPI2 | UI Automation |
| Čtení terminálu VS Code | ano | ano | ano |
| Klávesnicový vstup | xdotool | uinput | plánováno |
| Správa oken | xdotool | KWin D-Bus | plánováno |
| Ovládání prohlížeče | ano | ano | ano |
| Procházení souborů | ano | ano | ano |

## Sestavení

Vyžaduje **Go 1.24+**.

```bash
make build          # Sestaví binárku ptx-agent
make test           # Spustí testy
make lint           # Formátování + vet
make clean          # Odstraní artefakty sestavení
```

Výsledná binárka se jmenuje `ptx-agent` (na Windows `ptx-agent.exe`).

Ladicí nástroje pro inspekci accessibility:

```bash
make build-debug    # Sestaví uia-debug (Windows) nebo atspi-debug (Linux)
```

## Architektura

Agent používá vzor zásuvných backendů s oddělením platforem při kompilaci pomocí Go build tagů:

```text
src/
├── cmd/
│   └── ptx-agent/          # Vstupní bod
├── internal/
│   ├── accessibility/       # Multiplatformní abstrakce accessibility
│   │   ├── backend_atspi_*  # AT-SPI2 přes D-Bus (Linux)
│   │   └── backend_uia_*   # UI Automation přes COM (Windows)
│   ├── browser/             # Ovládání prohlížeče přes Playwright
│   ├── cli/                 # Definice příkazů Cobra
│   ├── config/              # Správa konfigurace
│   ├── files/               # Bezpečný prohlížeč souborů
│   ├── ipc/                 # Abstrakce IPC (transport D-Bus)
│   ├── keyboard/            # Simulace vstupu (uinput, xdotool)
│   ├── mcp/                 # MCP server (JSON-RPC přes stdio)
│   ├── scenario/            # Automatické spouštění scénářů
│   ├── vscode/              # Ovládání VS Code (nezávislé na platformě)
│   └── windows/             # Správa oken (KWin, xdotool)
```

Balíčky vyšší úrovně závisejí pouze na rozhraních, nikdy přímo na implementacích pro konkrétní platformu. Například balíček `vscode` používá pouze `accessibility.Backend` a funguje beze změny na Linuxu i na Windows.

## Repozitář

- **GitHub**: <https://github.com/CassandraGargoyle/portunix-agent>

## Dokumentace

- [Pokyny pro přispěvatele](docs/contributing/)
- [Záznamy o architektonických rozhodnutích (ADR)](docs/adr/)
- [Seznam issues](https://github.com/CassandraGargoyle/portunix-agent/issues) — podrobné popisy jsou v [docs/issues/](docs/issues/)

Odkazovaná dokumentace je v angličtině, která je oficiálním jazykem projektu.

## Licence

Licence MIT – podrobnosti jsou v souboru [LICENSE](LICENSE).
