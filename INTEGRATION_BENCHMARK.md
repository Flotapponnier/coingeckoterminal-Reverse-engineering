# Intégration GeckoTerminal au Benchmark Aggregator-Latency

## 📊 Ce qu'on a découvert

### ✅ Fonctionnel
- **WebSocket:** `wss://cables.geckoterminal.com/cable`
- **Protocol:** AnyCable v1.6.8 (ActionCable compatible)
- **Auth:** Aucune (juste Origin + User-Agent headers)
- **Channels:**
  - `SwapChannel` → Trades en temps réel avec `block_timestamp`
  - `PoolChannel` → Prix, volume, stats complètes

### 📈 Head Lag observé
- **GeckoTerminal:** ~11 secondes de lag
- **Comparaison approximative:**
  - Mobula: ~400ms - 1s
  - Codex: ~1-3s
  - **GeckoTerminal: ~11s** ⚠️

**Note:** Le lag élevé de GeckoTerminal pourrait être dû à :
- Batch processing côté serveur
- Délai de propagation des données
- Moins de priorité sur la latence (focus sur données agrégées/analytics)

### ⚠️ Limitation majeure : pool_id

**Problème:** Le WebSocket utilise un `pool_id` numérique interne (ex: `147971598`)

**pool_id trouvés:**
- ✅ ETH/USDC Uniswap V3 (ethereum): `147971598`
- ❌ SOL/USDC Raydium (solana): TBD
- ❌ WETH/USDC Base: TBD
- ❌ WBNB/BUSD BSC: TBD
- ❌ WETH/USDC Arbitrum: TBD

**Solution:** Trouver manuellement via DevTools (voir `FINDING_POOL_IDS.md`)

---

## 🔧 Fichiers créés

```
/Users/user/mobula/reverse engineering/coingecko/
├── main.go                      # Client WebSocket de base
├── head_lag_monitor.go         # Monitor standalone avec stats
├── find_pool_id.py             # Script pour chercher pool_id (scraping)
├── FINDING_POOL_IDS.md         # Guide pour trouver pool_id manuellement
├── INTEGRATION_BENCHMARK.md    # Ce fichier
└── README.md                    # Documentation générale
```

---

## 🚀 Option 1: Intégration Complète (Recommandée)

Ajouter GeckoTerminal comme 4ème provider dans le benchmark.

### Fichiers à modifier

#### 1. `/cmd/script/config.go`
```go
type Config struct {
	CoinGeckoAPIKey       string
	MobulaAPIKey          string
	DefinedSessionCookie  string
	// Pas besoin d'API key pour GeckoTerminal (WebSocket public)
}
```

#### 2. `/cmd/script/head_lag_monitor.go`
Ajouter après la fonction `runCodexHeadLagMonitor`:

```go
// ============================================================================
// GeckoTerminal WebSocket Monitor
// ============================================================================

type GeckoSwapEvent struct {
	Data struct {
		BlockTimestamp int64  `json:"block_timestamp"` // ms
		TxHash         string `json:"tx_hash"`
		// ... autres champs
	} `json:"data"`
	Type string `json:"type"` // "newSwap"
}

var geckoTerminalPools = []struct {
	Name    string
	PoolID  string
	Chain   string
}{
	{
		Name:   "ETH/USDC Uniswap V3",
		PoolID: "147971598",
		Chain:  "ethereum",
	},
	// TODO: Ajouter les autres pools une fois pool_id trouvés
}

func runGeckoTerminalHeadLagMonitor(config *Config, stopChan <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	fmt.Println("[HEAD-LAG][GECKO] Starting WebSocket monitor...")

	wsURL := "wss://cables.geckoterminal.com/cable"
	headers := http.Header{
		"Origin":     []string{"https://www.geckoterminal.com"},
		"User-Agent": []string{"Mozilla/5.0 (...)"},
	}

	// ... (copier depuis head_lag_monitor.go)
}
```

#### 3. `/cmd/script/head_lag_monitor.go` - Appel du monitor
Dans `runHeadLagMonitor`, ajouter:

```go
func runHeadLagMonitor(config *Config, stopChan <-chan struct{}) {
	// ... existing code ...

	var wg sync.WaitGroup

	// Start Mobula monitor
	wg.Add(1)
	go runMobulaHeadLagMonitor(config, stopChan, &wg)

	// Start Codex monitor
	wg.Add(1)
	go runCodexHeadLagMonitor(config, stopChan, &wg)

	// Start GeckoTerminal monitor
	wg.Add(1)
	go runGeckoTerminalHeadLagMonitor(config, stopChan, &wg)

	wg.Wait()
}
```

#### 4. `/grafana/dashboards/head_lag.json`
Ajouter une série pour GeckoTerminal avec couleur distincte (ex: violet)

```json
{
  "matcher": {
    "id": "byRegexp",
    "options": ".*geckoterminal.*"
  },
  "properties": [{
    "id": "color",
    "value": {
      "fixedColor": "purple",
      "mode": "fixed"
    }
  }]
}
```

### Métriques Prometheus

Utiliser les métriques existantes:
```
head_lag_milliseconds{aggregator="geckoterminal", chain="ethereum"}
head_lag_seconds{aggregator="geckoterminal", chain="ethereum"}
```

---

## 🎯 Option 2: Monitor Standalone (Plus Simple)

Garder GeckoTerminal dans son propre dossier comme outil de comparaison séparé.

**Avantages:**
- Pas de modifications au benchmark existant
- Facile à activer/désactiver
- Utile pour tester rapidement

**Utilisation:**
```bash
cd "/Users/user/mobula/reverse engineering/coingecko"
go run head_lag_monitor.go
```

---

## 📋 TODO pour intégration complète

### Phase 1: Trouver tous les pool_id (PRIORITÉ)
- [ ] Ouvrir chaque pool dans le navigateur
- [ ] DevTools → Network → WS → Copier pool_id
- [ ] Mettre à jour `geckoTerminalPools` dans le code

### Phase 2: Code
- [ ] Copier la logique de `head_lag_monitor.go` dans le benchmark
- [ ] Ajouter au `runHeadLagMonitor`
- [ ] Tester localement

### Phase 3: Grafana
- [ ] Ajouter série GeckoTerminal au dashboard
- [ ] Choisir couleur (violet/rose)
- [ ] Tester visualisation

### Phase 4: Déploiement
- [ ] Push sur GitHub
- [ ] Vérifier Railway redeploy
- [ ] Valider métriques dans Grafana

---

## 🤔 Recommandation

**Option 1 (Intégration complète)** SI:
- ✅ Tu veux une comparaison exhaustive de tous les providers
- ✅ Tu as le temps de trouver les pool_id pour les 5 chaînes
- ✅ Tu veux GeckoTerminal dans Grafana

**Option 2 (Standalone)** SI:
- ✅ Tu veux tester rapidement
- ✅ Tu ne veux pas toucher au benchmark stable
- ✅ Le lag de 11s te semble trop élevé pour le benchmark principal

**Mon avis:** Commencer par **Option 2** pour valider, puis **Option 1** si les résultats sont intéressants.

---

## 📊 Comparaison attendue dans Grafana

```
Head Lag (Seconds):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Mobula       ▁▂▁▂  (~0.5s)  🟠
Codex        ▂▃▂▃  (~2s)    🟢
GeckoTerminal ████ (~11s)   🟣
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

GeckoTerminal sera probablement le plus lent, mais c'est une donnée utile pour montrer:
- Les différences de priorité des providers
- Que certains priorisent l'analytics sur la latence
- La qualité du service de Mobula/Codex en comparaison

---

## 🎓 Leçons apprises

1. **Reverse engineering:** Utiliser DevTools Network/WS pour capturer les vraies requêtes
2. **AnyCable/ActionCable:** Format de message bien documenté
3. **Pool ID mapping:** Pas d'API publique, nécessite scraping ou inspection manuelle
4. **Head lag:** GeckoTerminal sacrifie latence pour données agrégées/analytics

---

## 🔗 Ressources

- GeckoTerminal: https://www.geckoterminal.com/
- AnyCable Protocol: https://docs.anycable.io/
- Repo: https://github.com/Flotapponnier/coingeckoterminal-Reverse-engineering
