<!-- updated: 2026-09-22T18:00:00Z -->
# Status Line

CLI Go pour afficher une status line Powerline personnalisée dans Claude Code.

## Architecture (Clean Architecture / Hexagonal)

```
cmd/statusline/              # Point d'entrée CLI (stdin JSON → stdout ANSI)
internal/
├── application/             # Service orchestration (StatusLineService)
├── domain/
│   ├── model/               # Entités (Input, Limit, LimitSet, Progress, Git, MCP...)
│   └── port/                # Interfaces (InputProvider, Renderer, GitRepository...)
├── adapter/                 # Adaptateurs externes
│   ├── git/                 # Git status + diff stats
│   ├── mcp/                 # Serveurs MCP (config, --mcp-config de l'hôte, plugins)
│   ├── mcpcalls/            # Serveurs MCP appelés en ce moment (transcripts)
│   ├── sessionstate/        # Session occupée + pid hôte (<config>/sessions/<pid>.json)
│   ├── system/              # Info système (OS, Docker)
│   ├── terminal/            # Largeur du terminal (COLUMNS)
│   ├── trace/               # STATUSLINE_TRACE : octets de chaque cadre
│   ├── transcript/          # Lecture de la fin d'un transcript JSONL
│   ├── updater/             # Auto-update binaire (GitHub releases)
│   └── usage/               # Usage API Anthropic (OAuth, burn-rate)
└── presentation/
    └── renderer/            # Rendu Powerline ANSI (segments, couleurs, icônes)
```

## Développement

```bash
make build          # Compile le binaire → bin/status-line
make test           # Lance les tests (go test ./...)
make lint           # Vérifie le code (ktn-linter)
make demo           # Démo avec données exemple
go run ./demo/widths [-light] [cols…]  # session chargée à 200…60 colonnes
```

Tout go test / go build / go vet passe par `~/.local/bin/lourd <cmd>` (slice
systemd plafonnée, partagée entre agents).

`demo/` = outil de dev (`demo/sample` = état réaliste) ; il utilise le vrai
renderer.

## Affichage

`STATUSLINE_LINE_GAP` = lignes vides entre les deux rangées, 0-3 (défaut `0`)
— **sans effet dans Claude Code**, qui supprime les lignes vides (voir
« Contrat avec l'hôte »)
`STATUS_LINE_NO_SELF_UPDATE` = `1` désactive l'auto-update (images managées)
`STATUSLINE_TRACE` = fichier où tracer chaque cadre (voir « Tracer les octets »)

L'auto-update vérifie le `.sha256` publié avec l'asset avant de remplacer le
binaire : une somme absente, malformée ou différente annule la mise à jour.
`STATUSLINE_GLYPHS` = `nerd` (défaut) | `text` (repli ASCII, sans Nerd Font)
`STATUSLINE_COLORS` = `truecolor` | `256` force la profondeur de couleur ;
absent, elle suit `COLORTERM` (`truecolor`/`24bit` → 24 bits, sinon 256)
`STATUSLINE_MCP_LINE` = `2` sort les serveurs MCP du segment OS vers une
pastille en ligne 2 (défaut : dans le segment OS de la ligne 1)
`STATUSLINE_WEIGHTS` = poids de condensation de la ligne 1 (voir plus bas)
`STATUSLINE_HIDE` = pastilles à masquer, séparées par des virgules :
`context`, `session`, `weekly`, `model`, `credits`, `health`

Les glyphes viennent tous des plages Nerd Font, comme le reste de la ligne : un
symbole Unicode générique retombe sur une autre police et devient illisible.

**Ligne 1 — tout ce qui concerne la session :**

| Segment | Description |
|---------|-------------|
| OS | Icône système + étincelles 󰙴 = état de Claude (vert/orange/rouge), puis `󰒍 N` serveurs MCP, puis `󰚩 N` sous-agents hors épic affiché |
| Model | Pill colorée (Haiku/Sonnet/Opus/Fable) + jauge d'effort + fast mode |
| Path | Répertoire où la session travaille réellement (voir ci-dessous) |
| Git | Branche + modifiés/non-trackés + nombre de worktrees liés (hors prunable) |
| Changes | Lignes ajoutées/supprimées |

Les quotas du compte (session 5h, hebdo, quota scopé au modèle courant) sont
rendus dans le segment du modèle, séparés par un `\ue0b1`. Un quota scopé à une
famille de modèles ne s'affiche que si ce modèle est en cours d'utilisation.

| Segment | Description |
|---------|-------------|

| Pastille | Description |
|----------|-------------|
| ctx | Fenêtre de contexte, en pourcentage |
| session | Quota 5h : repère de brûlure régulière, atterrissage, reset |
| weekly | Quota 7j global — absent sur les forfaits qui n'en ont pas |
| *modèle* | Quota 7j scopé par famille de modèle (`limits[]`) |
| coût / credits | Coût cumulé de la session, solde de crédits |

**Largeur (ligne 1 seulement)** : `adapter/terminal` lit `COLUMNS` (posé
par l'hôte pour la commande de status line, = `process.stdout.columns`, sans
rien en retrancher) ; absent, non entier, ≤ 0 ou > 10000 → 120. Pas de
`/dev/tty` ni d'exec. Budget = `COLUMNS − 4` (`lineMargin`). **Ce 4 n'est pas
une marge de prudence : c'est exactement la largeur de la boîte où l'hôte
pose la status line** — `colonnes − 2 × QK` avec `QK = 2`, mesuré dans
2.1.283 — et l'hôte tronque chaque ligne à cette largeur. Une ligne qui tient
dans le budget n'est donc **jamais** coupée ; une ligne qui déborde est
coupée **à chaque redraw**. Largeur visible
(`VisibleWidth`, `width.go`) : échappements CSI/OSC ignorés, glyphes Nerd
Font (PUA) = 1 cellule, blocs larges CJK/emoji = 2 (petite table), marques
combinantes = 0.

**Politique de condensation** (`condensePolicy`, `fit.go`, une ligne par
segment : paliers du plus riche au plus maigre, chacun lisible seul, et un
poids — plus il est bas, plus tôt le segment cède) :

| Segment | Poids | Paliers |
|---------|-------|---------|
| context | 10 | barre + libellé + % → icône + % |
| scoped | 20 | barre + libellé + % + rebours → libellé + % + rebours → libellé + % → initiale + % |
| weekly | 30 | idem scoped |
| session | 40 | barre + % + rebours → % + rebours → % |
| path | 50 | 30 → 20 → dernier élément → caché (dans un dépôt seulement) |
| branch | 60 | entière → 20 → 12 → 8 runes (`…`, compteurs `!3 ?1` toujours gardés) |
| changes | 240 | affichés → cachés |
| model | 255 | icône + nom → nom (la jauge d'effort reste) |
| OS | — | ne rétrécit jamais (icône, santé, MCP, sous-agents) |

Passes : l'étape n d'un segment de poids w a le rang (n−1)·100 + w ; la passe
n baisse d'un palier chaque segment qui en a encore un, poids croissant ; un
poids > 100 retient un segment pour une passe ultérieure (changes, model).
Ordre par défaut = celui de l'utilisateur : barres (context, scoped, weekly,
session), chemin et branche à 20, rebours (scoped, weekly, session), chemin
au dernier élément et branche à 12, initiales, changements, chemin caché,
icône du modèle, branche à 8. `STATUSLINE_WEIGHTS=context=10,weekly=30,…`
remplace des poids (noms ci-dessus, entiers 0-999 ; entrée inconnue ou
malformée ignorée). Les états successifs (`fitLevels`, 18 : la ligne pleine + 17 étapes) sont cumulatifs et
de largeur décroissante, donc le premier qui tient est trouvé par
bissection — même résultat qu'une marche pas à pas qui remesure après chaque
étape : 1 rendu si la ligne tient, 6 au plus sinon. Aucun état ne tient →
le dernier. Largeur 0 (tests) = pas de contrainte. `renderer.Condensed`
dit quels segments ont cédé (utilisé par `demo/widths`).

L'indicateur MCP du segment OS (≈ 4-7 cellules) est dessiné à chaque
état : jamais abandonné, jamais déplacé, compté dans le budget. Les deux
derniers paliers (icône du modèle, branche à 8) existent pour lui : la
session chargée de `demo/widths` tient alors en 76 cellules à 80 colonnes —
mais seulement sans `󰚩 N` : les sous-agents hors épic ajoutent 4 cellules et
la font passer à 80, donc au-delà du budget. Sous 80 colonnes le dernier
palier déborde toujours ; l'hôte coupe alors la queue de la ligne.

**Ligne 2 :** une pastille par épic ouvert mène la ligne, puis la pastille
MCP si `STATUSLINE_MCP_LINE=2`, puis la mise à jour. Nous ne tronquons rien
nous-mêmes ; c'est l'hôte qui coupe (voir ci-dessous), un titre long compris.

## Tracer les octets

Un fond qui déborde est soit mal écrit, soit coupé, et un screenshot ne
permet pas de trancher. `STATUSLINE_TRACE=<fichier>` (`adapter/trace`) ajoute
un enregistrement par cadre : en-tête d'une ligne puis **les octets du cadre,
verbatim**.

```bash
# dans .claude/settings.json, env de la commande de status line, ou :
STATUSLINE_TRACE=~/statusline.trace  # puis relancer la session
```

```
\x1e frame ts=<RFC3339Nano> bytes=1166 wrote=1166 lines=2 cols=213 budget=209 emitted=175 cut=no chip=lit render=14593us err=-
<les 1166 octets du cadre>
```

- Le séparateur est `\x1e` (RS) : la charge utile contient des `\n` et des
  échappements, donc un lecteur découpe là-dessus et se fie à `bytes=`.
- `wrote=` est ce que l'écriture sur stdout a réellement pris. `wrote < bytes`
  ou `err` ≠ `-` **est** la preuve d'un cadre coupé chez nous.
- `budget=` = ce que le condenseur visait, `emitted=` = la largeur visible que
  la ligne 1 fait réellement, `cut=yes` quand la seconde dépasse la première.
  Comme le budget **est** la largeur de la boîte de l'hôte (voir « Contrat
  avec l'hôte »), `cut=yes` veut dire : l'hôte a tronqué ce cadre. Les deux
  nombres répondent d'un coup d'œil à « est-ce l'hôte qui nous a coupés ».
- `chip=lit` marque les cadres où la puce MCP était allumée
  (`renderer.ChipLit`) : dans une trace d'une journée, ce sont les seuls à
  regarder.
- Plafond 192 Mio (≈ un jour à un cadre/seconde) puis la trace s'arrête ;
  toute erreur d'écriture est avalée — la barre ne doit jamais casser à cause
  de son propre journal.

```bash
# le cadre allumé qui a en plus été tronqué : le suspect
grep -a 'chip=lit' ~/statusline.trace | grep 'cut=yes'
# combien de cadres l'hôte tronque, et de combien
grep -a -o 'budget=[0-9]* emitted=[0-9]* cut=yes' ~/statusline.trace | sort | uniq -c
```

## Contrat avec l'hôte

Lu dans Claude Code 2.1.283. À revérifier si l'hôte change de version : tout
ce qui suit est du comportement observé, pas une API.

**Cadre entier ou rien.** L'hôte lit notre stdout jusqu'à EOF (`close` du
process **et** `end` des deux flux) puis ne dessine le cadre **que si notre
code de sortie est 0**. Un process tué par un signal vaut `status: 1`, un tick
annulé (`aborted`) n'atteint jamais l'écran et laisse le dernier bon cadre
affiché. Il n'existe aucun chemin de rendu incrémental : le handler `stdout`
ne renifle la première ligne que pour y chercher un marqueur de hook async.
**Conséquence : l'hôte ne peut pas afficher un cadre partiel — sauf si nous
sortons 0 après n'avoir écrit qu'une partie de la ligne.** C'est pour ça que
`emit` (`cmd/statusline/main.go`) traite une écriture courte comme une erreur
et que `main` sort 1 : mieux vaut un redraw vide qu'une barre bavée.

**Budget d'octets.** Tout le cadre part en un seul `fmt`/`Write`. `os.File.Write`
reboucle sur une écriture courte, donc `n < len` n'arrive qu'avec une erreur,
et un `EPIPE` sur le fd 1 lève `SIGPIPE` qui tue le process (donc `status != 0`,
donc cadre jeté). L'atomicité de `PIPE_BUF` (4096) n'est donc *pas* ce qui nous
protège — c'est le code de sortie. Pour mémoire : cadre typique ≈ 1,2-1,5 Ko ;
pire cas mesuré (40 serveurs MCP, 99 sous-agents, épic déplié, 24 bits,
`COLUMNS=400`) = 2052 octets à 1 épic, **+140 octets par épic ouvert, donc
4096 franchi à 16 épics ouverts**.

**Timeout.** 600 000 ms par défaut pour la commande de status line (`Fa`). Un
rendu prend 15-75 ms : le timeout n'est jamais la cause de quoi que ce soit.
En revanche chaque tick **annule** le précédent, donc sous forte charge la
barre gèle sur le dernier cadre au lieu de clignoter.

**Troncature.** L'hôte découpe la sortie sur `\n` et rend chaque ligne en
`wrap: "truncate"` dans une boîte de **`colonnes − 2 × QK`, `QK = 2`**, plus
son propre `paddingX = statusLine.padding ?? 0` : sans `padding` configuré,
la largeur de coupe est donc `COLUMNS − 4`, **le même nombre que notre
budget**. La coupe est une tranche **par cellule** (`Bun.sliceAnsi(ligne, 0,
largeur − 1)`) suivie d'un `…` : elle ne peut donc jamais tomber au milieu
d'un échappement. Une ligne déjà assez courte n'est pas touchée du tout
(`if (largeurVisible <= budget) return`). La largeur mesurée est
`Bun.stringWidth(ligne, {ambiguousIsNarrow: true})` — échappements ignorés,
glyphes Nerd Font (PUA) à 1 cellule, comme `VisibleWidth`.

Corollaire opérationnel : **il ne faut jamais tendre à l'hôte une ligne qu'il
doit couper.** Or c'est ce qui arrive dès que `fitLevels` est épuisé — avec
`󰚩 N` à 80 colonnes (80 cellules émises pour un budget de 76), et à toute
largeur sous 80. `STATUSLINE_TRACE` écrit `budget=`, `emitted=` et `cut=yes`
précisément pour rendre ça visible sans avoir à le déduire. Ce qui reste
**non mesuré** : si `Bun.sliceAnsi` referme ou non les attributs encore
ouverts à l'endroit de la coupe. Ça ne se joue que dans les 10 octets de la
puce allumée (voir « Règle des fonds »), et la coupe réelle tombe ~67
cellules plus loin, dans le segment des changements — donc hors de la puce.

**Nettoyage.** Avant l'affichage l'hôte fait
`stdout.trim().split("\n").map(trim).filter(non vide).join("\n")` : il
**supprime les lignes vides**, donc `STATUSLINE_LINE_GAP` n'a aucun effet
dans Claude Code, et il rogne l'espace de tête de la ligne 2.

**Report d'attributs.** L'hôte **reporte nos attributs d'une ligne sur la
suivante**, en préfixant chaque ligne de la concaténation de tous les
échappements SGR des lignes précédentes. Le `\033[0m` final de la ligne 1 est
donc la seule chose qui empêche la puce MCP allumée de peindre la ligne 2 —
c'est un invariant, pas un hasard : `bleed_internal_test.go` le vérifie.

## Règle des fonds : un seul fond par segment

**La puce MCP allumée est le seul endroit de toute la ligne qui change de
fond au milieu d'un segment.** Tous les autres accents (encre de quota, jauge
d'effort, compteurs git, cases d'épic) ne changent que l'encre, sur le fond du
segment. C'est donc le seul endroit où une coupure peut laisser un fond
« étranger » ouvert sur le reste de la rangée — et la seule raison pour
laquelle ce bug n'existe nulle part ailleurs.

Règle à suivre pour tout nouveau segment :

1. **Chaque écriture pose son propre fond et sa propre encre**, sans se fier à
   l'état ambiant laissé par un `Reset` précédent. C'est déjà le cas partout ;
   c'est ce qui rend 44 des 64 `Reset` du paquet redondants (retirer l'un
   d'eux ne change aucun octet visible).
2. **Fond et encre dans une seule séquence SGR** (`mergeSGR`, `mcppill.go`)
   dès qu'il s'agit d'un fond qui n'est pas celui du segment. Deux séquences
   laissent une fenêtre où le nouveau fond porte l'ancienne encre ; une seule
   séquence est appliquée en entier ou, coupée, pas du tout. C'est aussi plus
   court, donc plus loin du budget d'octets.
3. **Fermer par un `Reset` nu collé au dernier octet utile.** `\033[0m` fait
   4 octets, c'est la fermeture la plus courte possible ; la fondre avec
   l'ouverture suivante l'allongerait et agrandirait la fenêtre.

Résultat mesuré pour la puce : la plage d'octets où une coupure laisse la
sarcelle ouverte passe de **25 à 10 octets** (glyphes Nerd) — la charge utile
(`󰒍 7`, 6 octets) plus le reset (4). Ces 10 octets sont irréductibles : on ne
peut pas dessiner du blanc sur sarcelle sans que la sarcelle soit ouverte
au-dessus du glyphe. `TestACutNeverLeavesTealOpenBeyondTheChip` mesure cette
plage et échoue si elle grandit.

**Invariants de rendu** (`bleed_internal_test.go`) : un automate SGR rejoue
chaque ligne comme un terminal et exige que chaque ligne finisse sur un reset
complet, qu'aucun échappement ne soit tronqué, que chaque segment de la ligne 1
ouvre et ferme sur le même fond (un accent — la puce MCP allumée — est une
plage contiguë qui ne touche aucun bord), que les pastilles de la ligne 2
restent dans leurs capuchons (l'espace entre deux pastilles ne porte aucun
attribut), que `\033[9m` ne barre que des chiffres, et qu'allumer la puce ne
change pas la largeur de la ligne. La matrice couvre les deux jeux de glyphes,
les deux profondeurs de couleur, les deux lignes possibles pour MCP, tous les
états de santé, 16 formes de l'indicateur, les sous-agents, les formes de la
ligne 2 et les paliers de condensation.

À quoi s'ajoute `TestACutNeverLeavesTealOpenBeyondTheChip`, qui ne juge pas un
cadre entier mais **tous ses préfixes** : pour chaque décalage d'octet de la
ligne 1, il rejoue le préfixe et demande quel fond le terminal garderait en
main. La réponse doit être « la sarcelle seulement à l'intérieur des octets de
la puce », en une seule plage contiguë (voir « Règle des fonds »).

## Serveurs MCP

`adapter/mcp` lit, par ordre de priorité (le premier qui nomme un serveur
gagne, dédoublonnage par nom) :

1. **managed** — `/etc/claude-code/managed-mcp.json` (macOS : `/Library/Application Support/ClaudeCode`)
2. **ligne de commande** — `--mcp-config` de l'hôte : `sessionstate` retrouve
   le pid dans `<config>/sessions/<pid>.json` (un seul scan, partagé avec
   `Working`), puis `/proc/<pid>/cmdline`. Répétable, variadique jusqu'au
   flag suivant, `--mcp-config=v` accepté ; valeur = JSON si elle commence par
   `{`, sinon chemin (relatif au `cwd` de l'hôte). `--strict-mcp-config` = seuls
   managed + ligne de commande comptent. Hors Linux (pas de `/proc`) : ignoré.
3. **local** — `projects[<dir>].mcpServers` du fichier global
4. **projet** — `<dir>/.mcp.json`, repli `mcp.json`
5. **user** — `mcpServers` du fichier global
6. **plugins** — `<config>/plugins/installed_plugins.json` ; activé si
   `enabledPlugins["<plugin>@<marketplace>"]` vaut `true` (`<config>/settings.json`,
   puis `<dir>/.claude/settings.json`, puis `settings.local.json`) ;
   serveurs dans `<installPath>/.mcp.json` (`mcpServers` ou map nue), marqués
   `Plugin`.

Fichier global = `$CLAUDE_CONFIG_DIR/.claude.json`, sinon `~/.claude.json`
puis `~/.claude/.claude.json`. `<config>` = `$CLAUDE_CONFIG_DIR` ou `~/.claude`.
Désactivé = `"disabled": true`, ou listé dans `projects[<dir>].disabledMcpServers`
(nom nu ou `plugin:<plugin>:<serveur>`) ; `disabledMcpjsonServers` vise
`.mcp.json`. Tout fichier illisible ou malformé est ignoré ; pas d'exec.

**Indicateur (défaut)** : dans le segment OS de la ligne 1, sur son fond
blanc 255, après l'étincelle de santé et avant `󰚩 N` : glyphe MCP
`\U000F048D` (󰒍, présent dans MesloLGS NF — `fc-list ":charset=f048d"`)
en sarcelle 23 gras (6.46:1), puis le **total** des serveurs actifs en encre
OS 232 gras : `󰒍 7` (glyphes texte : `MCP 7`). Aucun détail par portée, pas
de capuchons. Des serveurs désactivés ajoutent `·N`, gris 241 (5.26:1 ; 242
ne tient que 4.53, 244 tombe à 3.4), N barré. Aucun serveur, rien.

**Pastille (ligne 2, `STATUSLINE_MCP_LINE=2`)** : même contenu, pastille à
capuchons arrondis 116, encre 23 gras sur 116 ; `·N` en 239 (240 ne tient
que 4.31:1 sur 116).

L'adaptateur garde la portée (`MCPServer.Source`, `WithSource`, posée sur
chaque source avant la fusion) : donnée utile, non affichée.

**Appels en cours** (`adapter/mcpcalls`, sans hook) : 128 Ko de fin de
`transcript_path` et des transcripts des sous-agents en cours
(`agents.json`, `<dir transcript>/<session>/subagents/agent-<id>.jsonl`).
Appel en vol = `tool_use` `mcp__<clé>__<outil>` sans `tool_result` plus loin ;
il reste allumé 2 s après le `timestamp` du résultat (la ligne se redessine
chaque seconde, un appel court ne se verrait jamais). Un appel sans résultat
depuis 30 min est tenu pour perdu. Clé → serveur (`model.WithBusy`) :
nom normalisé (`[^A-Za-z0-9_-]` → `_`), `plugin_<plugin>_<serveur>` → serveur ;
une clé inconnue est ajoutée, allumée, sans portée, comptée active. Un appel
en vol (ou dans les 2 s) allume l'indicateur : glyphe et nombre deviennent
une puce 255 gras sur 23 (7.5:1), sur exactement les mêmes cellules (la
ligne ne bouge pas) ; `·N` reste, gris sur blanc. Fond, encre et graisse
partent en **une seule séquence** (`mcpLitOpen`) — voir « Règle des fonds :
un seul fond par segment », c'est le seul fond non-segment de la ligne. En
pastille de ligne 2,
toute la pastille s'allume (capuchons 23) et `·N` passe en 116 sur 23
(4.54:1). Lecture de queue partagée avec `activity` : `adapter/transcript`.

## Effort

Claude Code n'envoie que le nom du niveau (`effort.level`), jamais l'échelle.
L'échelle `low → medium → high → xhigh → max` est figée dans
`model.effortScale` (testée) ; chaque niveau atteint est un disque dans
l'encre de la pastille, les autres un disque dans sa teinte pâle
(`GetModelTrack`). Un niveau hors échelle s'écrit en clair (`· ultra`), jamais
sur une jauge fausse.

## Tâches, épics et sous-agents

`adapter/tasks` lit deux sources, par session :

1. **MCP tasks de `kodflow-hooks`** (prioritaire dès que le fichier existe) :
   `<config>/kodflow/sessions/<session_id>/tasks.json`. v2 = liste `epics`
   (`id`, `agent`, `title` ≤ 20, `touched`) + `active` (agent → épic) ; un
   fichier v1 (`epics` en dict agent → épic courant) est converti à la
   lecture (`touched` = maintenant, `active[agent]` = son épic). Seules les
   tâches de `main` comptent ; `epic: 0` = hors épic ; une tâche d'un épic
   absent du fichier (épic fini, v1) est ignorée. Un fichier illisible
   n'affiche rien — il ne rebascule pas sur la source native.
2. **Outils natifs** (repli, fichier MCP absent) : `<config>/tasks/<liste>/<n>.json`,
   liste = `CLAUDE_CODE_TASK_LIST_ID` sinon `session-` + 8 premiers caractères,
   rendue comme la pastille `Tâches`.

**Épic ouvert** = au moins une tâche non terminée, ou l'épic actif encore
vide (il disparaît dès que le focus passe ailleurs). Un épic fini n'est plus
dessiné ; une tâche ajoutée à un épic fini le rouvre. `active.main` à 0,
absent ou pointant nulle part = aucun épic actif : la pastille `Tâches`
(tâches `epic: 0`) devient l'active. Ordre : l'actif, puis par `touched`
décroissant (`Tâches` : dernier `created`/`updated` de ses tâches).

**Pastilles (ligne 2)** : fond mauve 182, capuchons arrondis, encre prune 53.

- *Repliée* : `titre fait/total` (+ `󰚩 N`).
- *Dépliée* — seulement l'épic actif, seulement **pendant que la session
  travaille** : `titre fait/total cases titre-de-tâche` (+ `󰚩 N`). Cases
  triées : fait ■ encre 53, en cours ■ ambre `#82480b` qui **pulse** en
  `#9e5204` gras les secondes paires (horloge murale, `clockNow`), le reste
  (à faire, en attente) □ sur la piste pâle `#b48cb4`. Titre = la tâche en
  cours, sinon la première en attente de l'utilisateur, sinon la prochaine.
  Replis 256 : 94, 94 gras (le cube n'a pas d'ambre plus clair qui tienne
  3:1 sur le mauve — 130 tombe à 2.47:1), 139.
- Le pulse suppose `statusLine.refreshInterval: 1` — le minimum de l'hôte ;
  le clignotement ANSI (SGR 5) est filtré et ne sert à rien.

**Session occupée** : `adapter/sessionstate` lit `<config>/sessions/*.json`
(écrits par l'hôte) ; le premier dont `sessionId` = le `session_id` de stdin
décide : `status == "busy"`. Fichiers illisibles ignorés, pas d'exec.

**Sous-agents** : `agents.json` (hooks SubagentStart/Stop), `epic` = l'épic
actif de `main` au démarrage (absent = 0). En cours = pas de `stop` et
démarré il y a < 12 h. Un sous-agent dont l'épic est affiché va dans sa
pastille ; les autres (épic 0 ou épic non affiché) vont en **ligne 1**, dans
le segment OS après le glyphe de santé : `󰚩 N`.

## Répertoire actif

`workspace.current_dir` ne bouge pas quand l'agent travaille par `cd X && …`
ou par chemins absolus. `adapter/activity` lit les 256 derniers Ko de
`transcript_path` et prend le dernier emplacement nommé par un appel d'outil
(`file_path`/`path`, `cd X` en tête, `git -C X`), remonté à sa racine git.
`~/.claude/projects` (journaux, mémoire) est ignoré. Git s'exécute dans ce
répertoire ; MCP garde le répertoire de session.

## État de Claude

`adapter/health` lit `status.claude.com/api/v2/summary.json`, hors composant
« Government » et hors groupes. 1 composant dégradé/partiel = orange ; ≥ 2, ou
un `major_outage` = rouge ; la maintenance ne compte pas. Cache 2 min
rafraîchi par `--refresh-health` détaché ; au-delà de 15 min, rien n'est
dessiné. Le rendu ne touche jamais le réseau.

## Palette

- Fonds pâles (xterm-256), encre sombre de la même teinte : chaque couple
  encre/fond tient **≥ 4.5:1** (`contrast_internal_test.go`, qui lit aussi
  les échappements 24 bits `38;2;r;g;b`).
- Pastilles modèle : fonds poudrés (Haiku 224, Sonnet 189, Opus 223, Fable
  194, inconnu 252). Encres en **24 bits** vers ~5.4:1 — le cube 256 n'a rien
  dans ces teintes entre une encre profonde (7:1+) et une sous 4.5:1. Chaque
  encre a un repli 256 (`inkXxx256`) : la couleur du cube la plus proche en
  CIELAB qui tient 4.5:1 (Opus saute l'olive 58, jugé sale sur la pêche).
  Les deux jeux sont testés ; `pickInk` choisit au démarrage.
- Le curseur de rythme `●` et les barres prennent l'encre de leur pastille :
  pas de couleur d'accent fixe (l'ancien orange 166 tombait à 2:1 sur Sonnet).
- Contexte : fond 152, encre 24. Deux segments adjacents ne partagent jamais
  un fond proche (`TestAdjacentSegmentsDifferInGround`).

## Quotas : d'où viennent les chiffres

stdin (`rate_limits`, Claude Code >= 2.1.140) est prioritaire — gratuit et
synchrone. L'API OAuth n'apporte que ce que stdin ignore : quotas scopés par
modèle, crédits, et les buckets non envoyés. Côté API, `limits[]` est la source
agnostique au forfait ; `five_hour`/`seven_day` ne sont qu'un repli et valent
`null` sur certains forfaits.

**Un quota absent des deux sources reste absent** : l'afficher à 0 % mentirait
sur le compte. Réponse API mise en cache 60 s, rafraîchie par un processus
détaché — l'affichage n'attend jamais le réseau.

## Convention Go

- Tests: `*_test.go` dans le même package (`_internal_test` et `_external_test`)
- Structure: `cmd/` et `internal/` (pas de `/src`)
- Go 1.25.5 (go.mod), toolchain Go 1.26

## Documentation

| Fichier | Contenu |
|---------|---------|
| [`docs/vision.md`](docs/vision.md) | Vision projet, problème résolu, principes, non-goals |
| [`docs/architecture.md`](docs/architecture.md) | Diagramme C4, composants, data flow, contraintes |
| [`docs/workflows.md`](docs/workflows.md) | Setup, dev loop, tests, déploiement, CI/CD |
| [`AGENTS.md`](AGENTS.md) | Mapping stack → agents spécialistes (`/review`, `/lint`) |

## Reviews IA (label-triggered)

CodeRabbit et Qodo Merge ne tournent **pas** automatiquement sur chaque PR
(trop bruyant alongside l'un de l'autre). Ils sont déclenchés par label :

| Label | Outil | Quand l'utiliser |
|-------|-------|------------------|
| `coderabbit` | CodeRabbit | Review approfondie (89 règles ast-grep + path_instructions) |
| `qodo` | Qodo Merge | Second avis indépendant (security review + tests review) |

Les chemins `.devcontainer/**`, `.github/**`, `vendor/**` sont exclus des
deux outils — c'est du contenu template-managed, pas du code projet.

## Branch protection (main)

| Règle | Valeur |
|-------|--------|
| Required status checks | `test`, `build (linux\|darwin\|windows, amd64\|arm64)` (7 checks) |
| Required approving reviews | 0 (solo maintainer) |
| Dismiss stale reviews on push | ✓ (CodeRabbit CHANGES_REQUESTED s'auto-dismiss au push suivant) |
| Force push | ✗ |
| Branch deletion | ✗ |
