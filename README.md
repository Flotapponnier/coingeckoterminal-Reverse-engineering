# CoinGecko Terminal - Reverse Engineering

Reverse engineering de CoinGecko Terminal (DEX data aggregator) pour accéder aux données temps réel.

## 🎯 Objectifs

1. **WebSocket Temps Réel** - Prix, trades, pools live
2. **API REST** - Pool info, OHLCV, trending coins
3. **Nouveau Token Detection** - Nouveaux pools/tokens créés
4. **Métadonnées** - Token info, socials, descriptions

## 📋 Plan d'Action

### Phase 1: Reconnaissance (Exploration)
- [ ] Analyser le site web CoinGecko Terminal
- [ ] Identifier les endpoints WebSocket
- [ ] Capturer les requêtes/réponses dans Network tab
- [ ] Documenter l'architecture

### Phase 2: Authentification
- [ ] Identifier le système d'auth (API key, session, JWT, etc.)
- [ ] Tester si l'auth est requise
- [ ] Documenter le flow d'authentification

### Phase 3: WebSocket Implementation
- [ ] Identifier le protocole (ws, graphql-ws, socket.io, etc.)
- [ ] Documenter les subscriptions disponibles:
  - Prix en temps réel
  - Trades
  - Pool events
  - Nouveau tokens
- [ ] Créer client Python (exploration)
- [ ] Créer client Go (production)

### Phase 4: REST API
- [ ] Documenter tous les endpoints REST
- [ ] Rate limits
- [ ] Response formats
- [ ] Créer client Go

### Phase 5: Integration
- [ ] Intégrer au benchmark (si pertinent)
- [ ] Comparer avec Mobula/Defined.fi
- [ ] Documentation finale

## 🔍 Endpoints Découverts

### WebSocket
```
URL: wss://cables.geckoterminal.com/cable
Protocol: AnyCable v1.6.8 (ActionCable compatible)
Auth: Non requis (juste Origin + User-Agent headers)
Server: Cloudflare + AnyCable
Status: ✅ Confirmé fonctionnel
```

**Headers requis:**
```
Origin: https://www.geckoterminal.com
User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36
```

### REST API
```
Base URL: https://api.geckoterminal.com/api/v2/
Auth: TBD
```

## 📊 Subscriptions Disponibles

| Subscription | Description | Status |
|-------------|-------------|---------|
| **SwapChannel** | Trades/swaps en temps réel | ✅ Fonctionnel |
| **PoolChannel** | Prix, volume, stats du pool | ✅ Fonctionnel |

### SwapChannel - Données reçues
```json
{
  "tx_hash": "0x...",
  "from_token_amount": "4756.419711",
  "to_token_amount": "2.00852662245715",
  "price_from_in_usd": "1.00197609252112",
  "price_to_in_usd": "2372.79346120288",
  "block_timestamp": 1770047795000,
  "tx_from_address": "0x...",
  "from_token_id": 1337529,
  "to_token_id": 1337526
}
```

### PoolChannel - Données reçues
```json
{
  "base_price_in_usd": "2372.79346120288",
  "quote_price_in_usd": "1.00197609252112",
  "reserve_in_usd": "66670960.6651",
  "from_volume_in_usd": "205688356.279988",
  "price_change_data": {
    "last_300_s": {...},    // 5 minutes
    "last_900_s": {...},    // 15 minutes
    "last_1800_s": {...},   // 30 minutes
    "last_3600_s": {...},   // 1 heure
    "last_7200_s": {...},   // 2 heures
    "last_21600_s": {...},  // 6 heures
    "last_43200_s": {...},  // 12 heures
    "last_86400_s": {...},  // 24 heures
    "last_172800_s": {...}" // 48 heures
  },
  "transaction_data": {
    "buys": 5983,
    "sells": 5413
  },
  "fdv_in_usd": "5341084451.250967",
  "market_cap_in_usd": "5339500976.10621"
}
```

## 🚀 Usage

### Go Client (Production Ready)
```bash
cd "/Users/user/mobula/reverse engineering/coingecko"
go run main.go
```

**Sortie :**
```
🦎 CoinGecko Terminal WebSocket Client
✅ Connecté au WebSocket
📨 Message de bienvenue reçu
✅ Subscription confirmée: PoolChannel
✅ Subscription confirmée: SwapChannel
[16:56:43] Swap reçu: 4756 USDC → 2.008 WETH ($4,765)
[16:56:43] Pool update: WETH = $2,372 | Volume 24h: $205M
```

### Python (Exploration)
```bash
pip install -r requirements.txt
python3 explore.py
```

## 🔧 Prochaines Étapes

### ✅ Phase 1: WebSocket (COMPLÉTÉ)
- [x] Connexion WebSocket fonctionnelle
- [x] SwapChannel subscription
- [x] PoolChannel subscription
- [x] Parsing des messages en temps réel

### 🚧 Phase 2: REST API (En cours)
- [ ] **Trouver pool_id depuis adresse de pool**
  - Endpoint probable: `/api/v2/networks/{network}/pools/{address}`
- [ ] Get pool info
- [ ] Get OHLCV data
- [ ] Get trending pools
- [ ] Get new pools

### 📋 Phase 3: Structures de données
- [ ] Créer types Go pour SwapData
- [ ] Créer types Go pour PoolData
- [ ] Parser les price_change_data
- [ ] Parser les transaction_data

## 📝 Notes

- CoinGecko Terminal: https://www.geckoterminal.com/
- Similaire à DEXTools, DEXScreener
- Focus sur DEX data (Uniswap, PancakeSwap, Raydium, etc.)
