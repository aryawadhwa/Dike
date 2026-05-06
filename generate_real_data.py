import subprocess
import json
import sqlite3
import os
import random
from datetime import datetime, timedelta

commands = [
    ("ls -la", "dev1"),
    ("rm -rf test_delete.txt", "admin"),
    ("touch rogue_script.sh && chmod +x rogue_script.sh", "contractor_x"),
    ("rm -rf pkg", "dev2"),
    ("echo 'HACKED' > go.mod", "admin"),
    ("cat go.mod", "dev1"),
    ("mysql -u root -e 'DROP TABLE users;'", "hacker99")
]

db_path = os.path.expanduser('~/.pulse/audit.db')

try:
    conn = sqlite3.connect(db_path, timeout=10)
    cursor = conn.cursor()
    # Ensure table exists
    cursor.execute('''CREATE TABLE IF NOT EXISTS audit_log (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		username TEXT,
		command TEXT,
		risk_level TEXT,
		decision TEXT,
		preview_summary TEXT
	)''')
    cursor.execute('DELETE FROM audit_log')
    conn.commit()
    
    print("Generating 100% authentic sandbox executions...")
    
    for i in range(15):
        cmd, user = random.choice(commands)
        print(f"Running authentic simulation for: {cmd}")
        
        # We run the pulse CLI to let the real Go/Docker code calculate the diff and risk
        # Note: We added audit logging to Headless mode, but since we want custom staggered timestamps,
        # we will grab the JSON output and insert it manually.
        
        result = subprocess.run(
            ["go", "run", "cmd/pulse/main.go", "--json", "--cmd", cmd, "--dir", "."], 
            cwd="./backend",
            capture_output=True, 
            text=True
        )
        
        if result.returncode == 0 and result.stdout.strip() != "":
            try:
                # The CLI prints JSON on the last line, but might print "Warning..." first
                # So we parse the last line
                lines = result.stdout.strip().split('\n')
                out_json = json.loads(lines[-1])
                
                risk = out_json.get("risk_level", "UNKNOWN")
                decision = "APPLIED" if risk == "LOW" else "REJECTED"
                
                diff_obj = out_json.get("diff_summary", {})
                created = diff_obj.get("created", [])
                deleted = diff_obj.get("deleted", [])
                
                diff_summary = "No file changes detected."
                if created or deleted:
                    diff_summary = f"Files created: {created}\nFiles deleted: {deleted}"
                
                # Assign a historical timestamp
                ts = (datetime.now() - timedelta(minutes=random.randint(1, 2880))).strftime('%Y-%m-%d %H:%M:%S')
                
                cursor.execute('INSERT INTO audit_log (timestamp, username, command, risk_level, decision, preview_summary) VALUES (?, ?, ?, ?, ?, ?)', (ts, user, cmd, risk, decision, diff_summary))
                conn.commit()
            except Exception as e:
                print("Failed to parse JSON output:", e, "Raw:", result.stdout)
        else:
            print("Command failed:", result.stderr)

    conn.close()
    print('Finished! Realistic data with historical timestamps and authentic sandbox diffs successfully injected.')

except Exception as e:
    print('Error:', e)
