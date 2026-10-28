#!/usr/bin/env bash
# Copy the Aegisbox database from Neon (AEGISBOX_NEON_DATABASE) to the
# Dokploy Postgres (AEGISBOX_DOKPLOY_DATABASE) and verify every table.
#
#   ./migrate_neon_to_dokploy.sh check     # read only: versions, sizes, table clashes
#   ./migrate_neon_to_dokploy.sh migrate   # dump Neon, restore into Dokploy, verify
#
# Stop the smpp-gateway, manager-api and portal before "migrate" so nothing is
# written to Neon after the dump. Neon is never modified.
# Uses Docker's postgres:17 client if available (works with Postgres 17 and older),
# otherwise the local pg_dump/pg_restore/psql.

set -euo pipefail

MODE="${1:-check}"
SRC="${AEGISBOX_NEON_DATABASE:?set AEGISBOX_NEON_DATABASE}"
DST="${AEGISBOX_DOKPLOY_DATABASE:?set AEGISBOX_DOKPLOY_DATABASE}"
# pg_dump should use Neon's direct endpoint, not the "-pooler" one.
SRC_DIRECT="${SRC/-pooler./.}"
WORKDIR="${WORKDIR:-./aegisbox-migration-$(date +%Y%m%d-%H%M%S)}"
DUMP="$WORKDIR/aegisbox.dump"

mkdir -p "$WORKDIR"
WORKDIR="$(cd "$WORKDIR" && pwd)"
DUMP="$WORKDIR/aegisbox.dump"

if command -v docker >/dev/null 2>&1 && docker info >/dev/null 2>&1; then
  pg() { docker run --rm -i --network host -v "$WORKDIR:$WORKDIR" postgres:17 "$@"; }
else
  pg() { "$@"; }
fi
q() { pg psql "$1" -X -v ON_ERROR_STOP=1 -At -c "$2"; }

log() { printf '\n== %s\n' "$*"; }

tables_sql="select table_schema||'.'||table_name from information_schema.tables
  where table_type='BASE TABLE' and table_schema not in ('pg_catalog','information_schema')
  order by 1"

log "Source (Neon)"
q "$SRC_DIRECT" "select version()"
q "$SRC_DIRECT" "select current_database()||'  '||pg_size_pretty(pg_database_size(current_database()))"
src_tables="$(q "$SRC_DIRECT" "$tables_sql")"
echo "$(echo "$src_tables" | grep -c . || true) tables"

log "Target (Dokploy)"
q "$DST" "select version()"
q "$DST" "select current_database()||'  '||pg_size_pretty(pg_database_size(current_database()))"
dst_tables="$(q "$DST" "$tables_sql")"
echo "$(echo "$dst_tables" | grep -c . || true) tables"
[ -n "$dst_tables" ] && echo "$dst_tables" | sed 's/^/  existing: /'

clashes="$(comm -12 <(echo "$src_tables" | sort) <(echo "$dst_tables" | sort) | grep . || true)"
if [ -n "$clashes" ]; then
  log "STOP: these tables already exist in the target"
  echo "$clashes" | sed 's/^/  /'
  echo "Restoring would overwrite or mix data. Use an empty database (or a separate schema) as the target."
  exit 1
fi
[ -n "$dst_tables" ] && echo "Target is not empty, but no table names clash."

if [ "$MODE" = "check" ]; then
  log "Check passed. Run with 'migrate' after stopping the gateway, manager-api and portal."
  exit 0
fi
[ "$MODE" = "migrate" ] || { echo "unknown mode: $MODE"; exit 1; }

read -r -p "Gateway, manager-api and portal stopped? Type yes to continue: " ok
[ "$ok" = "yes" ] || exit 1

log "Dumping Neon to $DUMP"
pg pg_dump "$SRC_DIRECT" --format=custom --no-owner --no-acl --file="$DUMP"
ls -lh "$DUMP"

log "Restoring into Dokploy (single transaction: all or nothing)"
pg pg_restore --dbname="$DST" --no-owner --no-acl --exit-on-error --single-transaction "$DUMP"

log "Verifying exact row counts"
count_sql="select string_agg(t||'='||n, E'\n' order by t) from (
  select table_schema||'.'||table_name as t,
    (xpath('/row/c/text()', query_to_xml(format('select count(*) as c from %I.%I', table_schema, table_name), false, true, '')))[1]::text::bigint as n
  from information_schema.tables
  where table_type='BASE TABLE' and table_schema not in ('pg_catalog','information_schema')) s"
q "$SRC_DIRECT" "$count_sql" | grep -v '^public.goose' | grep . | grep -Fxv -f <(echo "$dst_tables") > "$WORKDIR/src_counts.txt" || true
q "$DST" "$count_sql" | grep . | grep -E "^($(cut -d= -f1 "$WORKDIR/src_counts.txt" | sed 's/\./\\./g' | paste -sd'|'))=" > "$WORKDIR/dst_counts.txt" || true
if diff "$WORKDIR/src_counts.txt" "$WORKDIR/dst_counts.txt"; then
  echo "All $(wc -l < "$WORKDIR/src_counts.txt") tables match."
else
  echo "MISMATCH: do not switch over. Neon is unchanged."
  exit 1
fi

log "Money check: wallet balances total"
for db in "$SRC_DIRECT" "$DST"; do q "$db" "select coalesce(sum(balance),0) from wallets"; done

log "Migration version"
q "$DST" "select max(version_id) from goose_db_version where is_applied"

log "Done. Point DATABASE_URL for smpp-gateway, manager-api, migration and the portal at AEGISBOX_DOKPLOY_DATABASE, then start them."
