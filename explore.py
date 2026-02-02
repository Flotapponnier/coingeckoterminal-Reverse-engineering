#!/usr/bin/env python3
"""
CoinGecko Terminal WebSocket Explorer
Explore et documente les WebSocket de GeckoTerminal
"""

import asyncio
import json
import websockets
from datetime import datetime

# URLs possibles à tester
POSSIBLE_WS_URLS = [
    "wss://www.geckoterminal.com/cable",
    "wss://api.geckoterminal.com/cable",
    "wss://cable.geckoterminal.com",
    "wss://www.geckoterminal.com/api/cable",
]

def log(msg):
    timestamp = datetime.now().strftime("%H:%M:%S.%f")[:-3]
    print(f"[{timestamp}] {msg}")

async def explore_websocket(url):
    """Explore un endpoint WebSocket"""
    log(f"Tentative de connexion à: {url}")

    try:
        async with websockets.connect(url) as ws:
            log(f"✅ Connecté à {url}")

            # Test 1: Attendre un message de bienvenue
            try:
                welcome = await asyncio.wait_for(ws.recv(), timeout=5.0)
                log(f"Message de bienvenue: {welcome}")
            except asyncio.TimeoutError:
                log("Pas de message de bienvenue")

            # Test 2: Envoyer un ping
            log("Envoi ping...")
            await ws.send(json.dumps({"type": "ping"}))

            try:
                response = await asyncio.wait_for(ws.recv(), timeout=5.0)
                log(f"Réponse ping: {response}")
            except asyncio.TimeoutError:
                log("Pas de réponse au ping")

            # Test 3: Subscribe à un channel (ActionCable format)
            log("Test subscription ActionCable...")
            subscribe_msg = {
                "command": "subscribe",
                "identifier": json.dumps({
                    "channel": "PoolChannel",
                    "pool_address": "0x88e6a0c2ddd26feeb64f039a2c41296fcb3f5640"  # ETH/USDC Uniswap V3
                })
            }
            await ws.send(json.dumps(subscribe_msg))

            # Écouter les messages pendant 30 secondes
            log("Écoute des messages pendant 30s...")
            try:
                for i in range(30):
                    msg = await asyncio.wait_for(ws.recv(), timeout=1.0)
                    log(f"Message reçu: {msg[:200]}...")  # Limiter l'affichage
            except asyncio.TimeoutError:
                log("Timeout - pas de messages")

    except Exception as e:
        log(f"❌ Erreur: {e}")

async def main():
    log("🔍 CoinGecko Terminal WebSocket Explorer")
    log("=" * 60)

    # Tester chaque URL
    for url in POSSIBLE_WS_URLS:
        await explore_websocket(url)
        log("-" * 60)
        await asyncio.sleep(2)

    log("✅ Exploration terminée")

if __name__ == "__main__":
    asyncio.run(main())
