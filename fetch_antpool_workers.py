import requests
import json
import csv
import os
import sqlite3
import time
from datetime import datetime

def init_db(db_name="antpool_workers.db"):
    conn = sqlite3.connect(db_name)
    cursor = conn.cursor()
    
    # Create table if not exists
    # Using 'id' from Antpool as PRIMARY KEY to avoid duplicates
    create_table_sql = """
    CREATE TABLE IF NOT EXISTS workers (
        id INTEGER PRIMARY KEY,
        worker_id TEXT,
        ip_address TEXT,
        hs_last_10min TEXT,
        hs_last_1h TEXT,
        hs_last_1d TEXT,
        reject_ratio TEXT,
        last_share_time INTEGER,
        status INTEGER,
        online_time_24h REAL,
        reconnect_24h REAL,
        fetched_at TIMESTAMP
    );
    """
    cursor.execute(create_table_sql)
    conn.commit()
    return conn

def derive_ip(worker_id):
    if "x" in worker_id:
        try:
            parts = worker_id.split("x")
            if len(parts) == 2:
                # Format: 172.16.{A}.{B}
                return f"172.16.{parts[0]}.{parts[1]}"
        except Exception:
            pass
    return ""

def fetch_antpool_data():
    url = "https://www.antpool.com/auth/v3/observer/api/worker/list"
    
    # Headers
    headers = {
        "accept": "application/json",
        "referer": "https://www.antpool.com/observer?accessKey=tHRWhY0DJFTLgfPhE9tC&coinType=BTC&observerUserId=sam001sz",
        "user-agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
    }
    
    # Cookies
    cookies = {
        "JSESSIONID": "249030E43E0239969FE66729AC74482F"
    }

    # Database connection
    conn = init_db()
    cursor = conn.cursor()
    
    page = 1
    page_size = 50 # Increased page size to reduce requests
    total_pages = 1
    total_workers_fetched = 0
    
    print(f"Starting fetch process...")

    while page <= total_pages:
        print(f"Fetching page {page}...")
        
        # Query parameters
        params = {
            "search": "",
            "workerStatus": "0",
            "accessKey": "tHRWhY0DJFTLgfPhE9tC",
            "coinType": "BTC",
            "observerUserId": "sam001sz",
            "pageNum": str(page),
            "pageSize": str(page_size)
        }
    
        try:
            response = requests.get(url, params=params, headers=headers, cookies=cookies, timeout=10)
            
            if response.status_code == 200:
                try:
                    data = response.json()
                    
                    code = data.get("code")
                    if str(code) == "0" or str(code) == "000000":
                        if "data" in data and isinstance(data["data"], dict):
                            worker_data = data["data"]
                            
                            # Calculate total pages on first request
                            if page == 1:
                                total_records = worker_data.get("totalRecord", 0)
                                total_pages = (total_records + page_size - 1) // page_size
                                print(f"Total records: {total_records}, Total pages: {total_pages}")
                            
                            items = worker_data.get("items", []) # 'items' might be 'rows' or 'list' depending on API, check previous JSON
                            # Based on previous output, it is 'items' inside 'data'
                            
                            if not items and "rows" in worker_data:
                                items = worker_data["rows"]
                            
                            if items:
                                for item in items:
                                    worker_id = item.get("workerId", "")
                                    ip_address = derive_ip(worker_id)
                                    
                                    # Insert into DB
                                    # Using INSERT OR REPLACE to update existing records
                                    cursor.execute("""
                                        INSERT OR REPLACE INTO workers (
                                            id, worker_id, ip_address, 
                                            hs_last_10min, hs_last_1h, hs_last_1d, 
                                            reject_ratio, last_share_time, status, 
                                            online_time_24h, reconnect_24h, fetched_at
                                        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
                                    """, (
                                        item.get("id"),
                                        worker_id,
                                        ip_address,
                                        item.get("hsLast10Min"),
                                        item.get("hsLast1Hour"), # JSON key is hsLast1Hour or hsLast1H? CSV header says hsLast1Hour
                                        item.get("hsLast1D"),
                                        item.get("rejectRatio"),
                                        item.get("shareLastTime"),
                                        item.get("workerStatus"),
                                        item.get("onlineTimeLast24h"),
                                        item.get("reconnectLast24h"),
                                        datetime.now()
                                    ))
                                
                                conn.commit()
                                count = len(items)
                                total_workers_fetched += count
                                print(f"Page {page}: Saved {count} workers. Total fetched: {total_workers_fetched}")
                            else:
                                print(f"Page {page}: No items found.")
                                break # Stop if no items
                        else:
                            print("Data field is missing or not a dictionary.")
                            break
                    else:
                        print(f"API returned error code: {code}")
                        print(f"Message: {data.get('msg', 'No message')}")
                        break
                        
                except json.JSONDecodeError:
                    print("Failed to decode JSON response.")
                    break
            else:
                print(f"Request failed. Status: {response.status_code}")
                break
                
            page += 1
            # Be nice to the API
            time.sleep(0.5)
            
        except requests.exceptions.RequestException as e:
            print(f"An error occurred: {e}")
            break

    conn.close()
    print("Done.")

if __name__ == "__main__":
    fetch_antpool_data()
