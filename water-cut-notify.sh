#!/bin/bash

current_date=`TZ=":Asia/Bangkok" date +%Y-%m-%d`
res=`curl "https://gisonline.mwa.co.th/GIS1125/SRC/src/06-Map%20MWA/rest/services/content-proxy-search.php?branch_param=&start_param=&finish_param=$current_date" \
    | jq -r '.[]|select(.impactarea | contains("บางคอแหลม"))'`

curl -X POST \
    -H "Authorization: Bearer $LINE_TOKEN" \
    -F "message=$res" \
        https://notify-api.line.me/api/notify
