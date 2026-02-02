#!/usr/bin/env python3
"""
Script Playwright pour extraire automatiquement les pool_id numériques
depuis GeckoTerminal en interceptant les messages WebSocket
"""

import asyncio
import json
import re
from playwright.async_api import async_playwright

# Pools à monitorer (du benchmark)
POOLS = [
    {
        "name": "ETH/USDC Uniswap V3",
        "network": "eth",
        "address": "0x88e6a0c2ddd26feeb64f039a2c41296fcb3f5640",
        "chain": "ethereum",
    },
    {
        "name": "SOL/USDC Raydium",
        "network": "solana",
        "address": "7qbRF6YsyGuLUVs6Y1q64bdVrfe4ZcUUz1JRdoVNUJnm",
        "chain": "solana",
    },
    {
        "name": "WETH/USDC Base",
        "network": "base",
        "address": "0x4c36388be6f416a29c8d8eee81c771ce6be14b18",
        "chain": "base",
    },
    {
        "name": "WBNB/BUSD PancakeSwap",
        "network": "bsc",
        "address": "0x58f876857a02d6762e0101bb5c46a8c1ed44dc16",
        "chain": "bnb",
    },
    {
        "name": "WETH/USDC Arbitrum",
        "network": "arbitrum",
        "address": "0xc6962004f452be9203591991d15f6b388e09e8d0",
        "chain": "arbitrum",
    },
]

results = {}


async def extract_pool_id(page, pool):
    """Extrait le pool_id en interceptant les WebSocket messages"""
    url = f"https://www.geckoterminal.com/{pool['network']}/pools/{pool['address']}"

    print(f"\n📍 {pool['name']}")
    print(f"   URL: {url}")

    pool_id_found = None

    # Intercepter les WebSocket frames
    def handle_websocket(ws):
        nonlocal pool_id_found

        def on_framesent(payload):
            nonlocal pool_id_found
            try:
                if isinstance(payload, str):
                    data = json.loads(payload)

                    # Chercher dans les subscriptions
                    if data.get("command") == "subscribe":
                        identifier = data.get("identifier", "")

                        if isinstance(identifier, str):
                            # Parse l'identifier JSON
                            try:
                                ident_data = json.loads(identifier)
                                if "pool_id" in ident_data:
                                    pool_id = ident_data["pool_id"]
                                    if pool_id and not pool_id_found:
                                        pool_id_found = pool_id
                                        print(f"   ✅ Found pool_id: {pool_id}")
                            except json.JSONDecodeError:
                                pass

                        # Regex fallback
                        if not pool_id_found:
                            match = re.search(r'"pool_id"\s*:\s*"(\d+)"', identifier)
                            if match:
                                pool_id_found = match.group(1)
                                print(f"   ✅ Found pool_id (regex): {pool_id_found}")
            except (json.JSONDecodeError, Exception) as e:
                pass

        ws.on("framesent", on_framesent)

    page.on("websocket", handle_websocket)

    try:
        # Aller sur la page
        await page.goto(url, wait_until="networkidle", timeout=30000)

        # Attendre un peu pour que le WebSocket se connecte et subscribe
        await page.wait_for_timeout(5000)

        if pool_id_found:
            return pool_id_found
        else:
            print(f"   ⚠️  pool_id not found in WebSocket messages")
            return None

    except Exception as e:
        print(f"   ❌ Error: {e}")
        return None


async def main():
    print("🦎 GeckoTerminal Pool ID Extractor (Playwright)")
    print("=" * 70)
    print("Intercepting WebSocket messages to extract numeric pool_id...")
    print()

    async with async_playwright() as p:
        # Lancer browser headless
        browser = await p.chromium.launch(headless=True)

        # Créer un contexte avec User-Agent
        context = await browser.new_context(
            user_agent="Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36"
        )

        page = await context.new_page()

        # Extraire pool_id pour chaque pool
        for pool in POOLS:
            pool_id = await extract_pool_id(page, pool)

            results[pool["address"]] = {
                "name": pool["name"],
                "network": pool["network"],
                "chain": pool["chain"],
                "address": pool["address"],
                "pool_id": pool_id,
                "found": pool_id is not None,
            }

            # Petit délai entre les requêtes
            await page.wait_for_timeout(2000)

        await browser.close()

    # Afficher résumé
    print("\n" + "=" * 70)
    print("📊 RÉSUMÉ")
    print("=" * 70)

    for result in results.values():
        status = "✅" if result["found"] else "❌"
        pool_id = result.get("pool_id", "NOT FOUND")
        print(f"{status} {result['name']:<35} pool_id={pool_id}")

    # Générer code Go
    print("\n" + "=" * 70)
    print("💾 Code Go pour head_lag_monitor.go:")
    print("=" * 70)
    print()
    print("var monitoredPools = []struct {")
    print("\tName    string")
    print("\tNetwork string")
    print("\tPoolID  string")
    print("\tChain   string")
    print("}{")

    for result in results.values():
        if result["found"]:
            print(f"\t{{")
            print(f"\t\tName:    \"{result['name']}\",")
            print(f"\t\tNetwork: \"{result['network']}\",")
            print(f"\t\tPoolID:  \"{result['pool_id']}\",")
            print(f"\t\tChain:   \"{result['chain']}\",")
            print(f"\t}},")

    print("}")

    # Sauvegarder en JSON
    with open("pool_ids_found.json", "w") as f:
        json.dump(results, f, indent=2)

    print("\n✅ Résultats sauvegardés dans pool_ids_found.json")

    # Compter succès
    found_count = sum(1 for r in results.values() if r["found"])
    total_count = len(results)

    print(f"\n🎯 Trouvé: {found_count}/{total_count} pool_id")


if __name__ == "__main__":
    asyncio.run(main())
