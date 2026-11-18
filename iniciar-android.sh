#!/data/data/com.termux/files/usr/bin/bash
set -e
launcher="$(cd -- "$(dirname -- "$0")" && pwd)"
if ! command -v java >/dev/null; then
  echo 'Instale o Java no Termux com: pkg install openjdk-21'
  exit 1
fi
exec "$launcher/mundo-na-rede-android-arm64" "$@"
