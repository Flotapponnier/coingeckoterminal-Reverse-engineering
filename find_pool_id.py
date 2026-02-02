#!/usr/bin/env python3
"""
Script pour trouver le pool_id numérique depuis l'adresse du pool
"""

import requests
import json
from bs4 import BeautifulSoup

# Pools du benchmark à trouver
BENCHMARK_POOLS = [
    {
        "name": "ETH/USDC Uniswap V3",
        "network": "eth",
        "address": "0x88e6a0c2ddd26feeb64f039a2c41296fcb3f5640",
        "known_pool_id": "147971598"  # Déjà trouvé dans Network tab
    },
    {
        "name": "SOL/USDC Raydium",
        "network": "solana",
        "address": "7qbRF6YsyGuLUVs6Y1q64bdVrfe4ZcUUz1JRdoVNUJnm",
    },
    {
        "name": "WETH/USDC Base",
        "network": "base",
        "address": "0x4c36388be6f416a29c8d8eee81c771ce6be14b18",
    },
    {
        "name": "WBNB/BUSD PancakeSwap",
        "network": "bsc",
        "address": "0x58f876857a02d6762e0101bb5c46a8c1ed44dc16",
    },
    {
        "name": "WETH/USDC Arbitrum",
        "network": "arbitrum",
        "address": "0xc6962004f452be9203591991d15f6b388e09e8d0",
    },
]

def fetch_pool_page(network, address):
    """Fetch la page HTML du pool pour extraire le pool_id"""
    url = f"https://www.geckoterminal.com/{network}/pools/{address}"

    headers = {
        "User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36"
    }

    print(f"[+] Fetching {url}")

    try:
        response = requests.get(url, headers=headers, timeout=10)

        if response.status_code != 200:
            print(f"    ❌ Status: {response.status_code}")
            return None

        # Chercher pool_id dans le HTML
        html = response.text

        # Méthode 1: Chercher dans les meta tags
        soup = BeautifulSoup(html, 'html.parser')

        # Méthode 2: Chercher "pool_id" dans le texte
        if 'pool_id' in html:
            # Trouver toutes les occurrences de pool_id
            import re
            matches = re.findall(r'"pool_id["\s:]+(\d+)"', html)
            if matches:
                pool_id = matches[0]
                print(f"    ✅ Found pool_id: {pool_id}")
                return pool_id

        # Méthode 3: Chercher dans window.__NEXT_DATA__ ou autre
        if '__NEXT_DATA__' in html:
            start = html.find('__NEXT_DATA__')
            if start != -1:
                # Extraire le JSON
                start = html.find('{', start)
                end = html.find('</script>', start)
                if start != -1 and end != -1:
                    json_str = html[start:end]
                    try:
                        data = json.loads(json_str)
                        # Chercher pool_id récursivement
                        pool_id = find_pool_id_recursive(data)
                        if pool_id:
                            print(f"    ✅ Found pool_id in __NEXT_DATA__: {pool_id}")
                            return pool_id
                    except json.JSONDecodeError:
                        pass

        print(f"    ❌ pool_id not found")
        return None

    except Exception as e:
        print(f"    ❌ Error: {e}")
        return None

def find_pool_id_recursive(data, depth=0):
    """Cherche pool_id récursivement dans une structure de données"""
    if depth > 10:  # Limite de profondeur
        return None

    if isinstance(data, dict):
        if 'pool_id' in data:
            pool_id = data['pool_id']
            if isinstance(pool_id, (int, str)) and str(pool_id).isdigit():
                return str(pool_id)

        for value in data.values():
            result = find_pool_id_recursive(value, depth + 1)
            if result:
                return result

    elif isinstance(data, list):
        for item in data:
            result = find_pool_id_recursive(item, depth + 1)
            if result:
                return result

    return None

def main():
    print("🔍 GeckoTerminal Pool ID Finder")
    print("=" * 60)

    results = []

    for pool in BENCHMARK_POOLS:
        print(f"\n📍 {pool['name']}")
        print(f"   Network: {pool['network']}")
        print(f"   Address: {pool['address']}")

        if 'known_pool_id' in pool:
            print(f"   ✅ Known pool_id: {pool['known_pool_id']}")
            results.append({
                **pool,
                'pool_id': pool['known_pool_id'],
                'found': True
            })
        else:
            pool_id = fetch_pool_page(pool['network'], pool['address'])

            if pool_id:
                results.append({
                    **pool,
                    'pool_id': pool_id,
                    'found': True
                })
            else:
                results.append({
                    **pool,
                    'pool_id': None,
                    'found': False
                })

    # Afficher le résumé
    print("\n" + "=" * 60)
    print("📊 RÉSUMÉ")
    print("=" * 60)

    for result in results:
        status = "✅" if result['found'] else "❌"
        pool_id = result.get('pool_id', 'NOT FOUND')
        print(f"{status} {result['name']:<30} pool_id={pool_id}")

    # Générer le code Go
    print("\n" + "=" * 60)
    print("💾 Code Go pour le benchmark:")
    print("=" * 60)

    print("var geckoTerminalPools = []struct {")
    print("\tName      string")
    print("\tNetwork   string")
    print("\tAddress   string")
    print("\tPoolID    string")
    print("}{")

    for result in results:
        if result['found']:
            print(f"\t{{")
            print(f"\t\tName:    \"{result['name']}\",")
            print(f"\t\tNetwork: \"{result['network']}\",")
            print(f"\t\tAddress: \"{result['address']}\",")
            print(f"\t\tPoolID:  \"{result['pool_id']}\",")
            print(f"\t}},")

    print("}")

if __name__ == "__main__":
    main()
