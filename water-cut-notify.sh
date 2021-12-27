#!/bin/sh

current_date=`TZ=":Asia/Bangkok" date +%Y-%m-%d`
echo "region: $REGION"

res=`curl "https://gisonline.mwa.co.th/GIS1125/SRC/src/06-Map%20MWA/rest/services/content-proxy-search.php?branch_param=&start_param=&finish_param=$current_date" \
    | jq -r --arg REGION "$REGION" '.[]|select(\
        (.impactbranch | contains($REGION)) or \
        (.impactarea   | contains($REGION)) or \
        (.areaname     | contains($REGION)) \
        )'`

curl -X POST \
    -H "Authorization: Bearer $LINE_TOKEN" \
    -F "message=$res" \
        https://notify-api.line.me/api/notify
