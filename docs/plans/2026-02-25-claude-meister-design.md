# claude-meister - Projekt-Aufraeum-Tool

## Zweck

CLI/TUI-Tool das vergessene Claude Code Projekte findet (via squirrel) und beim Aufraeumen hilft: Git bereinigen, Docker stoppen, Projekte archivieren oder loeschen.

## Architektur

Eigenstaendiges Go-Binary. Nutzt `squirrel status --json --deep` als Datenquelle.

```
squirrel status --json --deep
        |
        v
+---------------------+
|   claude-meister    |
|                     |
|  +---------------+  |
|  |  squirrel     |  |  <- Datenquelle: parst squirrel JSON
|  |  adapter      |  |
|  +-------+-------+  |
|          |          |
|  +-------v-------+  |
|  |  project      |  |  <- Kernlogik: Git, Docker, FS-Operationen
|  |  manager      |  |
|  +-------+-------+  |
|          |          |
|  +-------v-------+  |
|  |  CLI / TUI    |  |  <- Cobra Commands + Bubble Tea TUI
|  +---------------+  |
+---------------------+
```

### Packages

- `cmd/` -- Cobra-basierte CLI-Commands
- `internal/squirrel/` -- Adapter: squirrel aufrufen, JSON parsen
- `internal/project/` -- Projekt-Operationen (Git, Docker, FS)
- `internal/tui/` -- Bubble Tea TUI-Komponenten
- `internal/config/` -- Konfiguration (Archiv-Pfad etc.)

## Commands

```
claude-meister
  scan              # TUI-Uebersicht aller Projekte (Hauptfeature)
  clean <project>   # Git aufraeumen: commit/discard/branch cleanup
  archive <project> # Projekt ins Archiv verschieben
  delete <project>  # Projekt komplett loeschen (mit Sicherheitsabfrage)
  docker-stop       # Docker/DDEV aller vergessenen Projekte stoppen
  config            # Einstellungen (Archiv-Pfad, Tage-Schwelle etc.)
```

### scan (TUI-Modus)

- Tabellenansicht: Name, Status, Branch, Dirty, Uncommitted, Days Idle, Score
- Farbcodierung: Rot = dirty/lange idle, Gelb = feature branch, Gruen = clean
- Enter -> Detail-View mit Aktions-Menue
- Aktionen direkt aus der TUI ausloesbar
- Filter: `--category openWork|sleeping|all`, `--dirty-only`

### clean <project>

- Zeigt Git-Status (uncommitted files, branch, stash)
- Optionen: commit, stash, discard
- Feature-Branch Cleanup: merge/delete verwaister Branches
- Immer mit Bestaetigung

### archive <project>

- Verschiebt Projektordner nach `~/Archive/projects/YYYY-MM/`
- Optional: letzten Git-Status als Notiz speichern
- `squirrel ack` aufrufen um es aus der Liste zu entfernen

### delete <project>

- Doppelte Bestaetigung (Projektname eintippen)
- Optional: vorher archivieren anbieten

### docker-stop

- Sucht in allen Squirrel-Projekten nach `docker-compose.yml` / `.ddev/`
- Stoppt laufende Container (`docker compose down` / `ddev stop`)
- Zeigt freigegebene Ressourcen an

## TUI-Layout

### Hauptansicht

```
+- claude-meister -------------------------------------------------+
| 25 Open Work | 22 Active | 63 Sleeping | 1 Acknowledged         |
+------------------------------------------------------------------+
| # | Name                    | Branch       | Dirty | Days | Score|
|---|-------------------------|--------------|-------|------|------|
| > | homematic-poc           | main         |  * 2  |   0  |  112 |
|   | nanoclaw                | main         |  * 2  |   0  |  104 |
|   | quasi                   | feat/shell-  |       |   0  |  101 |
|---|-------------------------|--------------|-------|------|------|
| [c]lean  [a]rchive  [d]elete  [D]ocker-stop  [q]uit  [?]help   |
+------------------------------------------------------------------+
```

### Detail-View

```
+- homematic-poc --------------------------------------------------+
| Path:     /Users/.../homematic-poc                               |
| Branch:   main                                                   |
| Status:   2 uncommitted files                                    |
| Last:     "mach mal eine zusammenfassung..."  (heute)            |
| Prompts:  79                                                     |
| Docker:   docker-compose.yml gefunden, Container laeuft          |
|                                                                  |
| Uncommitted:                                                     |
|   M  src/config.py                                               |
|   ?  notes.txt                                                   |
|                                                                  |
| +----------------------------+                                   |
| | > Commit all changes       |                                   |
| |   Discard all changes      |                                   |
| |   Stash changes            |                                   |
| |   Stop Docker              |                                   |
| |   Archive project          |                                   |
| |   Delete project           |                                   |
| |   Back                     |                                   |
| +----------------------------+                                   |
+------------------------------------------------------------------+
```

## Sicherheit

- Jede destruktive Aktion -> Bestaetigungsprompt
- Delete erfordert Eintippen des Projektnamens
- Archiv statt Delete wird immer als erste Option angeboten
- `--dry-run` Flag fuer alle Commands
- Kein automatisches `git push` -- nur lokale Operationen

## Konfiguration

Datei: `~/.config/claude-meister/config.yaml`

```yaml
archive_path: ~/Archive/projects
squirrel_days: 30
squirrel_depth: deep
ignore_paths:
  - ~/Versioncontrol/important-project
```

## Dependencies

- **cobra** -- CLI Framework
- **bubbletea** + **lipgloss** + **bubbles** -- TUI Framework
- **go-git** -- Git-Operationen
- **squirrel** -- Externe CLI-Dependency (muss installiert sein)

## Build-Phasen

1. Grundgeruest + `scan` Command (squirrel parsen, Tabelle anzeigen)
2. Detail-View + `clean` Command (Git-Operationen)
3. `archive` + `delete` Commands
4. `docker-stop` Command
5. Konfiguration + Polish
