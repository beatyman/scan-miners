import sqlite3
import requests
from requests.auth import HTTPDigestAuth, HTTPBasicAuth
import concurrent.futures
import json
import time
import argparse
import sys

# Configuration
DB_NAME = "antpool_workers.db"
# Common endpoints from api.txt
ENDPOINTS_TO_TRY = [
    "/cgi-bin/stats.cgi",
    "/cgi-bin/miner_stats.cgi",
    "/cgi-bin/summary.cgi",
    "/cgi-bin/get_system_info.cgi"
]
USERNAME = "root"
PASSWORD = "root"
MAX_WORKERS = 20  # Number of parallel threads
TIMEOUT = 3 # Reduced timeout for faster scanning of dead IPs

def init_db():
    conn = sqlite3.connect(DB_NAME)
    cursor = conn.cursor()
    
    # Create table for miner details
    cursor.execute("""
    CREATE TABLE IF NOT EXISTS miner_details (
        ip_address TEXT PRIMARY KEY,
        status TEXT,
        endpoint_used TEXT,
        response_json TEXT,
        updated_at TIMESTAMP
    );
    """)
    conn.commit()
    return conn

def get_ips_from_db():
    conn = sqlite3.connect(DB_NAME)
    cursor = conn.cursor()
    cursor.execute("SELECT DISTINCT ip_address FROM workers WHERE ip_address IS NOT NULL AND ip_address != ''")
    ips = [row[0] for row in cursor.fetchall()]
    conn.close()
    return ips

def scan_miner(ip):
    # print(f"Scanning {ip}...")
    result = {
        "ip_address": ip,
        "status": "failed",
        "endpoint_used": "",
        "response_json": "",
        "updated_at": time.strftime('%Y-%m-%d %H:%M:%S')
    }
    
    for endpoint in ENDPOINTS_TO_TRY:
        url = f"http://{ip}{endpoint}"
        try:
            # Try Digest Auth first (most common for miners)
            try:
                response = requests.get(
                    url, 
                    auth=HTTPDigestAuth(USERNAME, PASSWORD), 
                    timeout=TIMEOUT
                )
            except requests.exceptions.RequestException:
                continue

            # If 401, try Basic Auth
            if response.status_code == 401:
                try:
                    response = requests.get(
                        url, 
                        auth=HTTPBasicAuth(USERNAME, PASSWORD), 
                        timeout=TIMEOUT
                    )
                except requests.exceptions.RequestException:
                    continue

            if response.status_code == 200:
                result["status"] = "success"
                result["endpoint_used"] = endpoint
                try:
                    try:
                        json_data = response.json()
                        result["response_json"] = json.dumps(json_data)
                    except json.JSONDecodeError:
                        result["response_json"] = response.text
                except Exception:
                    result["response_json"] = response.text[:1000]
                
                print(f"[SUCCESS] {ip} found at {endpoint}")
                return result # Found a working endpoint
            elif response.status_code == 401:
                 print(f"[AUTH FAIL] {ip} at {endpoint} (Credentials rejected)")
        except requests.exceptions.RequestException:
            pass
            
    print(f"[FAIL] {ip} unreachable or no endpoint found")
    return result

def save_result(conn, result):
    cursor = conn.cursor()
    cursor.execute("""
        INSERT OR REPLACE INTO miner_details (ip_address, status, endpoint_used, response_json, updated_at)
        VALUES (?, ?, ?, ?, ?)
    """, (
        result["ip_address"],
        result["status"],
        result["endpoint_used"],
        result["response_json"],
        result["updated_at"]
    ))
    conn.commit()

def main():
    parser = argparse.ArgumentParser(description="Scan miners for stats")
    parser.add_argument("--ip", help="Scan a single IP address")
    args = parser.parse_args()

    conn = init_db()
    
    if args.ip:
        print(f"Scanning single IP: {args.ip}")
        result = scan_miner(args.ip)
        print(json.dumps(result, indent=4))
        save_result(conn, result)
    else:
        ips = get_ips_from_db()
        print(f"Found {len(ips)} IPs to scan.")
        
        # Use ThreadPoolExecutor for parallel scanning
        with concurrent.futures.ThreadPoolExecutor(max_workers=MAX_WORKERS) as executor:
            future_to_ip = {executor.submit(scan_miner, ip): ip for ip in ips}
            
            count = 0
            for future in concurrent.futures.as_completed(future_to_ip):
                result = future.result()
                save_result(conn, result)
                count += 1
                if count % 50 == 0:
                    print(f"Progress: {count}/{len(ips)}")
                    
    conn.close()
    print("Scan completed.")

if __name__ == "__main__":
    main()
