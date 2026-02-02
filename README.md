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
URL: TBD
Protocol: TBD
Auth: TBD
```

### REST API
```
Base URL: TBD
Auth: TBD
```

## 📊 Subscriptions Disponibles

| Subscription | Description | Status |
|-------------|-------------|---------|
| TBD | TBD | ⏳ |

## 🚀 Usage

### Python (Exploration)
```python
# TBD
```

### Go (Production)
```go
// TBD
```

## 📝 Notes

- CoinGecko Terminal: https://www.geckoterminal.com/
- Similaire à DEXTools, DEXScreener
- Focus sur DEX data (Uniswap, PancakeSwap, Raydium, etc.)
