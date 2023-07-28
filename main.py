import json
import os
from typing import Any
from typing import Dict

import geopandas as gpd
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

    if r:
        return r[0]  # first response is closest to input coordinates (hopefully)
    else:
        return None


def _is_within_affected_area(
    latitude: float, longitude: float, r: Dict[str, Any]
) -> bool:
    coordinates = []
    for i in r["polygons"]:
        raw = i["coordinates"]
        value = [[float(i["longitude"]), float(i["latitude"])] for i in raw]

        coordinates.append(value)
    # ----------------------------------------------------------------
    result_geojson = {
        "type": "FeatureCollection",
        "crs": {"type": "name", "properties": {"name": "results"}},
        "features": [
            {
                "type": "Feature",
                "geometry": {"type": "Polygon", "coordinates": coordinates},
            },
        ],
    }

    input_geojson = {
        "type": "FeatureCollection",
        "crs": {"type": "name", "properties": {"name": "results"}},
        "features": [
            {
                "type": "Feature",
                "geometry": {
                    "type": "Point",
                    "coordinates": [
                        float(os.getenv("LONGITUDE")),
                        float(os.getenv("LATITUDE")),
                    ],
                },
            },
        ],
    }
    # ----------------------------------------------------------------
    result_df = gpd.read_file(json.dumps(result_geojson), driver="GeoJSON")
    input_df = gpd.read_file(json.dumps(input_geojson), driver="GeoJSON")

    return input_df.within(result_df).to_list()[0]


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
    LATITUDE = float(os.getenv("LATITUDE"))
    LONGITUDE = float(os.getenv("LONGITUDE"))

    r = _fetch_data(latitude=LATITUDE, longitude=LONGITUDE)

    if r:
        if _is_within_affected_area(latitude=LATITUDE, longitude=LONGITUDE, r=r):
            message = _prepare_notification_message(r)

            NTFY_TOPIC = os.getenv("NTFY_TOPIC")
            _send_notification_message(message=message, ntfy_topic=NTFY_TOPIC)
