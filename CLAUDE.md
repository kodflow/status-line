<!-- updated: 2026-11-13T18:00:00Z -->
# Status Line

CLI Go pour afficher une status line Powerline personnalisée dans Claude Code.

## Architecture (Clean Architecture / Hexagonal)

```
cmd/statusline/              # Point d'entrée CLI (stdin JSON → stdout ANSI)
internal/
├── application/             # Service orchestration (StatusLineService)
├── domain/
│   ├── model/               # Entités (Input, Limit, LimitSet, Progress, Git...)
│   └── port/                # Interfaces (InputProvider, Renderer, GitRepository...)
├── adapter/                 # Adaptateurs externes
│   ├── activity/            # Répertoire où la session travaille (transcript)
│   ├── detach/              # Processus détachés (--refresh-*, --check-update)
│   ├── diskcache/           # Caches disque privés (usage, santé)
│   ├── git/                 # Git status + diff stats
│   ├── health/              # État des services Claude (status page, cache)
│   ├── system/              # Info système (OS, Docker)
│   ├── terminal/            # Largeur du terminal (COLUMNS)
│   ├── trace/               # STATUSLINE_TRACE : octets de chaque cadre
│   ├── transcript/          # Lecture de la fin d'un transcript JSONL
│   ├── updater/             # Auto-update détaché (état caché, enfants sous flock)
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

`STATUS_LINE_NO_SELF_UPDATE` = toute valeur non vide désactive l'auto-update
(images managées) : aucun enfant lancé, pastille masquée
`STATUSLINE_TRACE` = fichier où tracer les cadres qui prouvent quelque chose
(écriture courte, erreur, `cut=yes`) — voir « Tracer les octets »

L'auto-update vérifie le `.sha256` publié avec l'asset avant de remplacer le
binaire : une somme absente, malformée ou différente annule la mise à jour.

**La trame ne touche jamais le réseau.** L'hôte annule la commande en cours dès
qu'une nouvelle mise à jour de la barre arrive : l'ancien téléchargement fait
dans la trame était tué à mi-course (9 `status-line-update-*` partiels mesurés
sur le poste, v0.20.7/v0.20.8 jamais installées). Désormais :

- `cmd/statusline` : `run(args, stdin, stdout, *deps)` ; les drapeaux
  (`--version`, `--refresh-usage`, `--refresh-health`, `--check-update`,
  `--apply-update=<v>`) passent avant stdin et avant le `recover`. La trame lit
  `<dossier réel du binaire>/.status-line-update.json` (sous son propre
  `recover` : une panique y vaut « contrôle dû ») et lance
  `detach.Spawn("--check-update", STATUS_LINE_UPDATING)` **avant** le rendu,
  seulement si `next_attempt_at` est atteint, jamais sous l'opt-out, pour un
  build dev, ni si le dossier n'est pas inscriptible.
- `adapter/updater` : `Check` et `Apply` tiennent un `flock LOCK_EX|LOCK_NB`
  sur `.status-line-update.lock` (`lock_unix.go`, `linux || darwin` ; stub
  ailleurs : rien n'est remplacé sous Windows ; un fichier de verrou
  impossible à ouvrir repousse `next_attempt_at` d'1 h). Sous le verrou, `Check`
  supprime les temporaires d'état orphelins, relit l'état et sort s'il n'est
  plus dû (sauf `--check-update=force`), écrit `next_attempt_at` avant l'appel
  (succès +1 h ; échec 5 min doublés jusqu'à 1 h ; 403/429 : `Retry-After` puis
  `X-RateLimit-Reset` ne font qu'allonger ce délai, plafond 24 h), valide
  `^v\d+\.\d+\.\d+$`, et applique un tag plus récent dans le même processus.
  `Due` tient pour périmé un `next_attempt_at` à plus de 24 h (horloge dans le
  futur). Avant le `rename`, le binaire téléchargé est lancé avec `--version`
  (5 s) et doit annoncer le tag attendu (poste sans AVX2, `-X main.version`
  erroné).
  L'état s'écrit en `.status-line-update.json.tmp-<pid>` puis `rename` : le point
  initial le tient hors du nettoyage des `status-line-update-*` de plus d'1 h.
  `apiURL` et `downloadURLFor` sont des variables pour les tests.
- Champs de l'état : `checked_at`, `tag`, `next_attempt_at`, `check_error`,
  `apply_error`, plus `check_failing_since` (début de la série d'échecs) et
  `failures` (compteur du backoff).
- Pastille (en fin de ligne, après les changements) : allumée tant que `tag` > version
  compilée ; variante échec (`\uf071` sur fond rouge) si `apply_error` sur un tag
  encore en attente, ou si les contrôles échouent depuis plus d'1 h.
- Code de sortie suivi par un écrivain compteur : 0 octet sorti → trame de
  secours en texte brut sans SGR, exit 0 ; trame entière → 0 ; partielle → 1.
  `gather` (application) lance chaque fournisseur par `sync.WaitGroup.Go`
  (Go 1.25) ; chaque fermeture porte son propre `recover`, pile sur stderr :
  un fournisseur qui panique laisse sa valeur à zéro, les autres arrivent.
`STATUSLINE_GLYPHS` = `nerd` (défaut) | `text` (repli ASCII, sans Nerd Font)
`STATUSLINE_COLORS` = `truecolor` | `256` force la profondeur de couleur ;
absent, elle suit `COLORTERM` (`truecolor`/`24bit` → 24 bits, sinon 256)
`STATUSLINE_WEIGHTS` = poids de condensation de la ligne (voir plus bas)
`STATUSLINE_HIDE` = pastilles à masquer, séparées par des virgules :
`context`, `session`, `weekly`, `model`, `credits`, `health`

Les glyphes viennent tous des plages Nerd Font, comme le reste de la ligne : un
symbole Unicode générique retombe sur une autre police et devient illisible.

**Une seule ligne — tout ce qui concerne la session :**

| Segment | Description |
|---------|-------------|
| OS | Icône système + étincelles 󰙴 = état de Claude (vert/orange/rouge) — la santé est le dernier pas de toute la condensation |
| Model | Pill colorée (Haiku/Sonnet/Opus/Fable) + jauge d'effort + fast mode |
| Path | Répertoire où la session travaille réellement (voir ci-dessous) |
| Git | Branche + modifiés/non-trackés + nombre de worktrees liés (hors prunable) |
| Changes | Lignes ajoutées/supprimées |
| Update | Pastille de mise à jour, en fin de ligne (voir « Affichage ») |

Épics, serveurs MCP et compteur de sous-agents ne sont plus dessinés : la
status line n'en montre rien (le HUD garde l'arbre des épics). Les
adaptateurs `mcp`, `mcpcalls`, `sessionstate` et `tasks`, le port `tasks`,
les modèles `mcp`/`tasks`, la puce MCP (`mcppill.go`, `ChipLit`,
`mergeSGR`), `STATUSLINE_MCP_LINE` et `STATUSLINE_LINE_GAP` ont été retirés
avec eux.

Les quotas du compte (session 5h, hebdo, quota scopé au modèle courant) sont
rendus dans le segment du modèle, séparés par un ``. Un quota scopé à une
famille de modèles ne s'affiche que si ce modèle est en cours d'utilisation.

| Pastille | Description |
|----------|-------------|
| ctx | Fenêtre de contexte, en pourcentage |
| session | Quota 5h : repère de brûlure régulière, atterrissage, reset |
| weekly | Quota 7j global — absent sur les forfaits qui n'en ont pas |
| *modèle* | Quota 7j scopé par famille de modèle (`limits[]`) |
| coût / credits | Coût cumulé de la session, solde de crédits |

**Largeur** : `adapter/terminal` lit `COLUMNS` (posé
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
| update | 45 (échec : 345) | pastille → cachée |
| path | 50 | 30 → 20 → dernier élément → caché (dans un dépôt seulement) |
| branch | 60 | entière → 20 → 12 → 8 runes (`…`, compteurs `!3 ?1` toujours gardés) |
| changes | 240 | affichés → cachés |
| model | 255 | icône + nom → nom (la jauge d'effort reste) |
| OS | 370 | santé → icône OS seule |

Passes : l'étape n d'un segment de poids w a le rang (n−1)·100 + w ; la passe
n baisse d'un palier chaque segment qui en a encore un, poids croissant ; un
poids > 100 retient un segment pour une passe ultérieure (changes, model, OS).
Ordre par défaut : barres (context, scoped, weekly, session), **pastille
update**, chemin et branche à 20, rebours (scoped, weekly, session), chemin au
dernier élément et branche à 12, initiales, changements, chemin caché, icône
du modèle, branche à 8, **santé**. `STATUSLINE_WEIGHTS=context=10,weekly=30,…`
remplace des poids (noms ci-dessus, entiers 0-999 ; entrée inconnue ou
malformée ignorée). Les états successifs (`fitLevels`, 20 : la ligne pleine +
19 étapes) sont cumulatifs et de largeur décroissante, donc le premier qui
tient est trouvé par bissection — même résultat qu'une marche pas à pas qui
remesure après chaque étape : 1 rendu si la ligne tient, 6 au plus sinon.
Aucun état ne tient → le dernier. Largeur 0 (tests) = pas de contrainte.
`renderer.Condensed` dit quels segments ont cédé (utilisé par `demo/widths`).

**La santé cède en dernier, sur les deux échelles.** Poids 370 : au-dessus
de tout autre pas (le plus haut est la branche à 8 runes, rang 260). Le seul
moment où la lumière compte est le seul moment où il ne faut pas l'avoir
perdue pour gagner deux cellules. L'icône OS ne cède jamais : c'est elle qui
dit que la barre est la barre. `TestHealthIsTheLastStep` verrouille l'ordre
et la marge de poids.

**La pastille update** (`segUpdate`) est dessinée dans `renderLine1Fit`
après les changements. Allumée (une version plus récente est connue), elle
cède tôt, au rang 45 : c'est une information, pas une alerte. Sa **variante
échec** (application échouée, ou contrôles en échec depuis plus d'1 h) suit
une seconde échelle (`failLevels`, choisie par `levelsFor`) où la pastille
pèse `updateFailedWeight` = 345 : au-dessus du premier passage, donc barres,
chemin et branche cèdent avant elle, et sous la santé. Un poids
`STATUSLINE_WEIGHTS=update=…` plus haut que 345 vaut aussi pour la variante
échec (`failedWeights` prend le max).
`TestFailedUpdateHoldsOutPastTheFirstPass` le vérifie.

**Mesuré** (`busyLine` : Opus 5 xhigh, trois quotas avec fenêtres, chemin
profond, branche longue, `+12 -3`) : la ligne la plus serrée fait
**67 cellules**, le plancher est donc **71 colonnes**, inchangé — le segment
OS au palier le plus serré était déjà l'icône seule, et la pastille update y
est cachée. En dessous, les 34 cellules du segment modèle (nom + jauge
d'effort + deux quotas à l'initiale) et les 18 de git sont incompressibles
avec les paliers actuels. `TestTheFloorIsMeasured` épingle ce 71 ;
`TestLineOneNeverOverflowsTheHostsBox` vérifie les deux moitiés, sans
pastille, pastille allumée et variante échec : ça tient partout où le budget
le permet, et là où c'est impossible la ligne est exactement le palier le
plus serré — jamais un cran de trop.

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
\x1e frame ts=<RFC3339Nano> bytes=1166 wrote=1166 lines=1 cols=213 budget=209 emitted=175 cut=no render=14593us err=-
<les 1166 octets du cadre>
```

- Le séparateur est `\x1e` (RS) : la charge utile contient des `\n` et des
  échappements, donc un lecteur découpe là-dessus et se fie à `bytes=`.
- `wrote=` est ce que l'écriture sur stdout a réellement pris. `wrote < bytes`
  ou `err` ≠ `-` **est** la preuve d'un cadre coupé chez nous.
- `budget=` = ce que le condenseur visait, `emitted=` = la largeur visible que
  la ligne fait réellement, `cut=yes` quand la seconde dépasse la première.
  Comme le budget **est** la largeur de la boîte de l'hôte (voir « Contrat
  avec l'hôte »), `cut=yes` veut dire : l'hôte a tronqué ce cadre.
- Le champ `chip=` a disparu avec la puce MCP ; les autres champs gardent leur
  ordre, un lecteur d'une ancienne trace lit encore les nouvelles.
- Toute erreur d'écriture est avalée — la barre ne doit jamais casser à cause
  de son propre journal. Plafond 192 Mio en garde-fou.

**La trace ne garde que les cadres qui prouvent quelque chose**
(`Frame.worthKeeping`). La barre se redessine chaque seconde et a presque
toujours raison ; garder ces cadres-là, c'est remplir un disque avec la preuve
que tout va bien. Trois cas seuls sont retenus :

| Gardé quand | Ce que ça prouve |
|-------------|------------------|
| `wrote < bytes` | l'écriture est partie en morceaux : la coupure est chez nous |
| `err` ≠ `-` | idem, avec la raison |
| `cut=yes` | la ligne dépassait la boîte de l'hôte : c'est l'hôte qui a coupé |

**Mesuré** (400 cadres au repos à 344 colonnes, la largeur réelle du poste) :
**0 octet** — le fichier n'est même pas créé. Sans filtre, les mêmes 400
cadres font 523 600 octets (1 309 octets par enregistrement). Sur les 15 h
réellement enregistrées sur le poste — 152 367 cadres, 192 Mio, plafond
atteint — aucun cadre n'avait `cut=yes`.

Conséquence pratique : un fichier absent ou vide **est** le résultat normal,
pas une panne de câblage. Pour vérifier que la trace est bien armée, forcer un
cadre intéressant :

```bash
# COLUMNS minuscule => cut=yes => un enregistrement apparaît
echo '{}' | COLUMNS=20 STATUSLINE_TRACE=~/statusline.trace status-line >/dev/null
grep -a -c 'frame ts=' ~/statusline.trace
```

```bash
# combien de cadres l'hôte tronque, et de combien
grep -a -o 'budget=[0-9]* emitted=[0-9]* cut=yes' ~/statusline.trace | sort | uniq -c
```

## Enquête close : fond vert du segment OS

Épic #6. Symptôme rapporté : le fond sarcelle de la puce MCP (`BgMCPLabel`,
lu « vert » à l'écran) s'étendait au-delà de la puce et remplissait la fin du
segment blanc, texte en blanc. Intermittent, uniquement sur ce fragment.

**Close : la cause est retirée.** La puce MCP allumée était le seul endroit
de toute la ligne qui changeait de fond au milieu d'un segment, donc le seul
fond qu'un cadre coupé ou mal rendu pouvait laisser ouvert sur le reste de la
rangée. La décision D1 a retiré l'indicateur MCP de la status line : il n'y a
plus de fond étranger à un segment, et l'invariant de rendu l'exige
désormais (voir « Règle des fonds »). La seconde rangée a disparu avec les
épics, donc le report d'attributs d'une ligne sur la suivante n'a plus de
ligne suivante sur laquelle peindre.

Ce que l'enquête avait établi, gardé pour mémoire : aucun `Reset` manquant
(balayage par mutation des 64 `Reset` du paquet), aucun palier du condenseur
qui coupe, aucun cadre partiel affiché par l'hôte, aucune coupe de l'hôte au
milieu d'un échappement (tranche par cellule), jamais `cut=yes` en 15 h à 344
colonnes, et les 9 cadres allumés capturés étaient parfaits au niveau octet.
Le suspect restant était le terminal ; il n'a plus rien à mal rendre.

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
et que `run` sort 1 : mieux vaut un redraw vide qu'une barre bavée.

**Budget d'octets.** Tout le cadre part en un seul `fmt`/`Write`. `os.File.Write`
reboucle sur une écriture courte, donc `n < len` n'arrive qu'avec une erreur,
et un `EPIPE` sur le fd 1 lève `SIGPIPE` qui tue le process (donc `status != 0`,
donc cadre jeté). L'atomicité de `PIPE_BUF` (4096) n'est donc *pas* ce qui nous
protège — c'est le code de sortie. Pour mémoire : avec une seule ligne et
plus aucune pastille par épic, le cadre ne grandit plus avec la session ;
mesuré sur `demo/sample` à `COLUMNS=400` (pastille update comprise) :
**1559 octets** en 24 bits, 1425 en 256 couleurs.

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
doit couper.** C'est ce qui arrive dès que `fitLevels` est épuisé, sous le
plancher de 71 colonnes. `STATUSLINE_TRACE` écrit `budget=`, `emitted=` et
`cut=yes` précisément pour rendre ça visible sans avoir à le déduire.

**Nettoyage.** Avant l'affichage l'hôte fait
`stdout.trim().split("\n").map(trim).filter(non vide).join("\n")` : il
**supprime les lignes vides** — c'est pourquoi l'ancien `STATUSLINE_LINE_GAP`
n'a jamais eu d'effet, et il a été retiré.

**Report d'attributs.** L'hôte **reporte nos attributs d'une ligne sur la
suivante**, en préfixant chaque ligne de la concaténation de tous les
échappements SGR des lignes précédentes. Avec une seule ligne il n'y a plus
de suivante ; le `\033[0m` final reste exigé par l'invariant
(`bleed_internal_test.go`) au cas où l'hôte ajouterait ses propres lignes.

## Règle des fonds : un seul fond par segment

**Aucun segment ne change de fond en son milieu.** Tous les accents (encre de
quota, jauge d'effort, compteurs git, lumière de santé) ne changent que
l'encre, sur le fond du segment. La puce MCP allumée était l'unique
exception ; elle est partie, et avec elle `mergeSGR` et toute la mesure de la
plage d'octets « sarcelle ouverte ».

Règle à suivre pour tout nouveau segment :

1. **Chaque écriture pose son propre fond et sa propre encre**, sans se fier à
   l'état ambiant laissé par un `Reset` précédent.
2. **Un seul fond par segment.** Un second fond au milieu d'un segment est un
   défaut, quel qu'il soit : l'invariant le refuse.
3. **Fermer par un `Reset` nu collé au dernier octet utile.**

La pastille update n'est pas un segment de la chaîne : elle vient après la
dernière flèche, précédée d'un espace sans attribut, entre deux capuchons
arrondis, et son fond reste entre ses capuchons.

**Invariants de rendu** (`bleed_internal_test.go`, automate dans
`sgrreplay_internal_test.go`) : un automate SGR rejoue la ligne comme un
terminal et exige qu'elle finisse sur un reset complet, qu'aucun échappement
ne soit tronqué, que la sortie soit **une seule ligne** terminée par un `\n`,
que chaque segment tienne sur **un seul fond** d'un bout à l'autre, que la
pastille update reste dans ses capuchons (l'espace qui la précède ne porte
aucun attribut), qu'aucune cellule ne soit barrée, et qu'aucun séparateur
n'hérite de la graisse du bloc précédent. La matrice couvre les deux jeux de
glyphes, les deux profondeurs de couleur, tous les états de santé, les quatre
états de la pastille update, les formes de la ligne (segments absents) et
chaque palier des deux échelles de condensation.

## Effort

Claude Code n'envoie que le nom du niveau (`effort.level`), jamais l'échelle.
L'échelle `low → medium → high → xhigh → max` est figée dans
`model.effortScale` (testée) ; chaque niveau atteint est un disque dans
l'encre de la pastille, les autres un disque dans sa teinte pâle
(`GetModelTrack`). Un niveau hors échelle s'écrit en clair (`· ultra`), jamais
sur une jauge fausse.

## Répertoire actif

`workspace.current_dir` ne bouge pas quand l'agent travaille par `cd X && …`
ou par chemins absolus. `adapter/activity` lit les 256 derniers Ko de
`transcript_path` et prend le dernier emplacement nommé par un appel d'outil
(`file_path`/`path`, `cd X` en tête, `git -C X`), remonté à sa racine git.
`~/.claude/projects` (journaux, mémoire) est ignoré. Git s'exécute dans ce
répertoire.

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

CodeRabbit ne tourne **pas** automatiquement sur chaque PR (trop bruyant).
Il est déclenché par label :

| Label | Outil | Quand l'utiliser |
|-------|-------|------------------|
| `coderabbit` | CodeRabbit | Review approfondie (89 règles ast-grep + path_instructions) |

Les chemins `.devcontainer/**`, `.github/**`, `vendor/**` sont exclus de
la review — c'est du contenu template-managed, pas du code projet.

## Branch protection (main)

| Règle | Valeur |
|-------|--------|
| Required status checks | `test`, `build (linux\|darwin\|windows, amd64\|arm64)` (7 checks) |
| Required approving reviews | 0 (solo maintainer) |
| Dismiss stale reviews on push | ✓ (CodeRabbit CHANGES_REQUESTED s'auto-dismiss au push suivant) |
| Force push | ✗ |
| Branch deletion | ✗ |
