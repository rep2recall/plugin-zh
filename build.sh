#!/bin/bash

GOOS="$(go env GOOS | tr -d "\r\n")"

if [[ ! -f "../../dist/$GOOS/plugins/app/assets/zh.db" ]]; then
    if [[ ! -f ./assets/zh.db ]]; then
        cd ./scripts

        if [[ ! -d "./node_modules" ]]; then
            yarn || exit 1
        fi

        yarn ts-node ./src/13-tatoeba.ts
        yarn ts-node ./src/14-cedict.ts

        cd -
    fi

    mkdir -p "../../dist/$GOOS/plugins/app/assets/"
    cp ./assets/zh.db "../../dist/$GOOS/plugins/app/assets/"
fi

go build -o "../../dist/$GOOS/plugins/app/"
