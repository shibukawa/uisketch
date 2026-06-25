#!/bin/sh
set -eu

npm run build:web
rm -rf desktop_assets/assets
cp packages/web/dist/index.html desktop_assets/index.html
cp -R packages/web/dist/assets desktop_assets/assets
