# CoinGecko Terminal - Étapes de Reverse Engineering

## 🎯 Objectif Final
Créer un client Go complet pour accéder aux données temps réel de CoinGecko Terminal (prix, trades, pools, nouveaux tokens).

---

## 📝 Étape 1: Reconnaissance - Identifier les WebSocket

### Actions:
1. Ouvrir https://www.geckoterminal.com/
2. Ouvrir DevTools → Network → WS (WebSocket)
3. Naviguer sur une page de pool (ex: ETH/USDC)
4. Observer les connexions WebSocket

### À documenter:
- [ ] URL du WebSocket
- [ ] Protocole utilisé (ActionCable, Socket.io, GraphQL-WS, custom)
- [ ] Messages de connexion initiale
- [ ] Format des messages (JSON, binary, etc.)

### Script de test:
```bash
python3 explore.py
```

---

## 📝 Étape 2: Analyser l'Authentification

### Actions:
1. Dans Network tab, regarder les Headers de la connexion WebSocket
2. Vérifier s'il y a:
   - Cookies
   - API Key
   - Origin/Referer requirements
   - Tokens JWT

### À documenter:
- [ ] Type d'authentification
- [ ] Headers requis
- [ ] Est-ce que le WebSocket marche sans auth?

### Test:
```bash
# Tester connexion sans auth
go run main.go
```

---

## 📝 Étape 3: Identifier les Channels/Subscriptions

GeckoTerminal utilise probablement ActionCable (Ruby on Rails WebSocket).

### Channels à documenter:

#### 1. PoolChannel - Prix en temps réel
```json
{
  "command": "subscribe",
  "identifier": "{\"channel\":\"PoolChannel\",\"pool_address\":\"0x...\"}"
}
```

**Données attendues:**
- Prix actuel
- Volume 24h
- Transactions récentes
- Price chart updates

#### 2. TradeChannel - Trades en temps réel
```json
{
  "command": "subscribe",
  "identifier": "{\"channel\":\"TradeChannel\",\"pool_address\":\"0x...\"}"
}
```

**Données attendues:**
- Timestamp
- Prix
- Montant
- Type (buy/sell)

#### 3. NewPoolsChannel - Nouveaux pools créés
```json
{
  "command": "subscribe",
  "identifier": "{\"channel\":\"NewPoolsChannel\",\"network\":\"eth\"}"
}
```

**Données attendues:**
- Pool address
- Token0/Token1
- Liquidité initiale
- Timestamp de création

#### 4. TrendingChannel - Tokens trending
```json
{
  "command": "subscribe",
  "identifier": "{\"channel\":\"TrendingChannel\"}"
}
```

### Actions:
1. Pour chaque channel, capturer un message exemple
2. Documenter la structure des données
3. Identifier les paramètres requis

---

## 📝 Étape 4: REST API Endpoints

### Endpoints à documenter:

#### 1. Get Pool Info
```
GET /api/v2/networks/{network}/pools/{address}
```

#### 2. Get OHLCV Data
```
GET /api/v2/networks/{network}/pools/{address}/ohlcv
```

#### 3. Get Recent Trades
```
GET /api/v2/networks/{network}/pools/{address}/trades
```

#### 4. Search Tokens
```
GET /api/v2/search/pools?query={search}
```

#### 5. Trending Pools
```
GET /api/v2/networks/trending_pools
```

#### 6. New Pools
```
GET /api/v2/networks/new_pools
```

### À documenter pour chaque endpoint:
- [ ] URL complète
- [ ] Query parameters
- [ ] Response format
- [ ] Rate limits
- [ ] Auth requirements

### Test:
```bash
curl -H "Accept: application/json" \
  "https://api.geckoterminal.com/api/v2/networks/eth/pools/0x88e6a0c2ddd26feeb64f039a2c41296fcb3f5640"
```

---

## 📝 Étape 5: Implémenter le Client Go

### Fichiers à créer:

#### `client.go` - Client WebSocket principal
```go
type GeckoTerminalClient struct {
    conn *websocket.Conn
    subscriptions map[string]bool
}
```

#### `types.go` - Structures de données
```go
type Pool struct {
    Address string
    Network string
    Token0  Token
    Token1  Token
    Price   float64
    Volume24h float64
}

type Trade struct {
    Timestamp time.Time
    Price     float64
    Amount    float64
    Type      string
}
```

#### `subscriptions.go` - Gestion des subscriptions
```go
func (c *Client) SubscribeToPool(address string) error
func (c *Client) SubscribeToTrades(address string) error
func (c *Client) SubscribeToNewPools(network string) error
```

#### `rest.go` - Client REST API
```go
func GetPoolInfo(network, address string) (*Pool, error)
func GetOHLCV(network, address string, timeframe string) ([]OHLCV, error)
func GetTrendingPools() ([]Pool, error)
```

---

## 📝 Étape 6: Tests et Validation

### Tests à effectuer:

#### WebSocket
- [ ] Connexion/Déconnexion
- [ ] Subscribe/Unsubscribe
- [ ] Recevoir prix temps réel
- [ ] Recevoir trades
- [ ] Recevoir nouveaux pools
- [ ] Gestion des erreurs
- [ ] Reconnexion automatique

#### REST API
- [ ] Get pool info
- [ ] Get OHLCV
- [ ] Get trades
- [ ] Search tokens
- [ ] Rate limiting handling

---

## 📝 Étape 7: Documentation Finale

### À créer:
1. **README.md complet** avec:
   - Installation
   - Usage examples
   - API reference
   - Rate limits

2. **Examples/**:
   - `live_price.go` - Prix en temps réel
   - `new_tokens.go` - Détection nouveaux tokens
   - `pool_monitor.go` - Monitoring d'un pool
   - `trending.go` - Tokens trending

3. **FINDINGS.md**:
   - Résumé des découvertes
   - Limites identifiées
   - Comparaison avec Mobula/Defined.fi

---

## 📝 Étape 8: Intégration (Optionnel)

### Si pertinent pour le benchmark:
1. Ajouter GeckoTerminal comme provider
2. Comparer latence vs Mobula/Codex
3. Comparer couverture de données
4. Dashboard Grafana

---

## 🎯 Checklist Complète

### Phase 1: Exploration
- [ ] URL WebSocket identifiée
- [ ] Protocole identifié (ActionCable?)
- [ ] Format des messages documenté
- [ ] Test de connexion réussi

### Phase 2: WebSocket
- [ ] PoolChannel implémenté
- [ ] TradeChannel implémenté
- [ ] NewPoolsChannel implémenté
- [ ] TrendingChannel implémenté
- [ ] Reconnexion automatique

### Phase 3: REST API
- [ ] Pool info endpoint
- [ ] OHLCV endpoint
- [ ] Trades endpoint
- [ ] Search endpoint
- [ ] Trending endpoint

### Phase 4: Production Ready
- [ ] Error handling complet
- [ ] Rate limiting
- [ ] Tests unitaires
- [ ] Documentation
- [ ] Examples

### Phase 5: Déploiement
- [ ] Code propre et commenté
- [ ] README complet
- [ ] Push sur GitHub
- [ ] (Optionnel) Intégration au benchmark

---

## 🚀 Quick Start

### 1. Exploration initiale
```bash
# Python
pip install -r requirements.txt
python3 explore.py
```

### 2. Test Go
```bash
go mod download
go run main.go
```

### 3. Développement
```bash
# Créer une branche pour chaque feature
git checkout -b feat/pool-channel
# ... développer ...
git commit -m "Add PoolChannel subscription"
git push origin feat/pool-channel
```

---

## 📚 Ressources

- GeckoTerminal: https://www.geckoterminal.com/
- ActionCable Protocol: https://github.com/anycable/actioncable-client-node
- Gorilla WebSocket: https://github.com/gorilla/websocket
