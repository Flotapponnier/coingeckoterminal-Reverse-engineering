# Comment trouver les pool_id pour GeckoTerminal

Le `pool_id` est un identifiant numérique interne utilisé par le WebSocket de GeckoTerminal.

## ❌ Ce qui NE marche PAS

### REST API
```bash
curl "https://api.geckoterminal.com/api/v2/networks/eth/pools/0x88e6a0c2ddd26feeb64f039a2c41296fcb3f5640"
```
Renvoie `"id": "eth_0x88e6a0c2ddd26feeb64f039a2c41296fcb3f5640"` (pas le pool_id numérique)

### Scraping HTML
Le pool_id n'est pas dans le HTML initial (chargé en JavaScript côté client)

## ✅ Ce qui MARCHE

### Méthode 1: DevTools Network Tab (RECOMMANDÉ)

1. Ouvrir la page du pool sur GeckoTerminal :
   ```
   https://www.geckoterminal.com/{network}/pools/{address}
   ```

2. Ouvrir DevTools → **Network** → **WS** (WebSocket)

3. Cliquer sur la connexion `cables.geckoterminal.com`

4. Aller dans l'onglet **Messages**

5. Chercher les messages **envoyés** (flèche verte ⬆️) :
   ```json
   {"command":"subscribe","identifier":"{\"channel\":\"PoolChannel\",\"pool_id\":\"147971598\"}"}
   ```

6. Copier le `pool_id`

### Méthode 2: Browser Console (Alternative)

1. Ouvrir la page du pool

2. Ouvrir DevTools → **Console**

3. Coller ce code :
   ```javascript
   // Chercher dans le state React/Vue
   window.__NEXT_DATA__ ||
   Object.keys(window).find(k => k.includes('pool'))
   ```

4. Ou inspecter l'élément et chercher `data-pool-id` ou attributs similaires

### Méthode 3: Intercepter les requêtes (Pour bulk)

Utiliser un proxy comme mitmproxy ou Burp Suite pour capturer toutes les subscriptions WebSocket et extraire les pool_id automatiquement.

## 📋 Pool IDs trouvés manuellement

### Pour le benchmark aggregator-latency:

| Pool | Network | Address | pool_id | Status |
|------|---------|---------|---------|--------|
| **ETH/USDC Uniswap V3** | eth | `0x88e6a0c2ddd26feeb64f039a2c41296fcb3f5640` | `147971598` | ✅ Confirmé |
| SOL/USDC Raydium | solana | `7qbRF6YsyGuLUVs6Y1q64bdVrfe4ZcUUz1JRdoVNUJnm` | `???` | ⏳ À trouver |
| WETH/USDC Base | base | `0x4c36388be6f416a29c8d8eee81c771ce6be14b18` | `???` | ⏳ À trouver |
| WBNB/BUSD BSC | bsc | `0x58f876857a02d6762e0101bb5c46a8c1ed44dc16` | `???` | ⏳ À trouver |
| WETH/USDC Arbitrum | arbitrum | `0xc6962004f452be9203591991d15f6b388e09e8d0` | `???` | ⏳ À trouver |

## 🔧 TODO: Trouver les pool_id manquants

Pour chaque pool ci-dessus :
1. Ouvrir `https://www.geckoterminal.com/{network}/pools/{address}`
2. DevTools → Network → WS → Messages
3. Copier le `pool_id` depuis `{"command":"subscribe"...}`
4. Mettre à jour ce document
5. Mettre à jour `head_lag_monitor_gecko.go`

## 💡 Alternative: Utiliser l'ID de l'API REST

Si on ne peut pas obtenir les pool_id facilement, on peut:
- Utiliser uniquement ETH/USDC (pool_id connu)
- Ou ajouter GeckoTerminal comme provider "bonus" avec moins de pools
- Ou créer un mapping manuel pool_id → address dans le code
