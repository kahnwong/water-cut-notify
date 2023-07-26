import os
from typing import Any
from typing import Dict

import requests
from dotenv import load_dotenv


load_dotenv()


def _fetch_data(latitude: float, longitude: float) -> Dict[str, Any]:
    headers = {
        "Host": "mobile.mwa.co.th",
        "Accept": "*/*",
        "Accept-Language": "en-US,en;q=0.9",
        "User-Agent": "mwa-mobile-ios/2 CFNetwork/1408.0.4 Darwin/22.5.0",
    }

    url = f"https://mobile.mwa.co.th/api/mobile/no-water-running-area/latitude/{latitude}/longitude/{longitude}"
    r = requests.get(url, headers=headers).json()

    return r[0]  # first response is closest to input coordinates (hopefully)


def _prepare_notification_message(r: Dict[str, Any]) -> str:
    return f"""
    area: {r['areaName']}
    soi: {r['soi']}
    reason: {r['reason']}

    startDate: {r['startDate']}
    endDate: {r['endDate']}
    """


def _send_notification_message(message: str, ntfy_topic: str):
    url = f"https://ntfy.sh/{ntfy_topic}"

    r = requests.post(url, data=message.encode(encoding="utf-8"))
    assert r.status_code == 200


if __name__ == "__main__":
    r = _fetch_data(latitude=os.getenv("LATITUDE"), longitude=os.getenv("LONGITUDE"))

    message = _prepare_notification_message(r)
    _send_notification_message(message=message, ntfy_topic=os.getenv("NTFY_TOPIC"))
