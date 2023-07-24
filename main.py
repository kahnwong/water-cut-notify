import os

import requests

# fetch data
headers = {
    "Host": "mobile.mwa.co.th",
    "Accept": "*/*",
    "Accept-Language": "en-US,en;q=0.9",
    "User-Agent": "mwa-mobile-ios/2 CFNetwork/1408.0.4 Darwin/22.5.0",
}

url = f"https://mobile.mwa.co.th/api/mobile/no-water-running-area/latitude/{os.getenv('LATITUDE')}/longitude/{os.getenv('LONGITUDE')}"
r = requests.get(url, headers=headers).json()

# prepare message
r = r[0]

message = f"""
area: {r['areaName']}
soi: {r['soi']}
reason: {r['reason']}

startDate: {r['startDate']}
endDate: {r['endDate']}
"""

# send notification
url = f"https://ntfy.sh/{os.getenv('NTFY_TOPIC')}"

r = requests.post(url, data=message.encode(encoding="utf-8"))
assert r.status_code == 200
