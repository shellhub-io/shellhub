env_override=${ENV_OVERRIDE:-./.env.override}
cloud_dir=${CLOUD_DIR:-../cloud}

SHELLHUB_EDITION=$(
    . ./.env 2>/dev/null
    # shellcheck source=/dev/null # ENV_OVERRIDE names the file at run time
    [ -f "$env_override" ] && . "$env_override"
    eval "printf '%s' \"\${SHELLHUB_EDITION:-community}\""
)

SHELLHUB_EDITION=$(printf '%s' "$SHELLHUB_EDITION" | tr '[:upper:]' '[:lower:]' | tr -d '[:space:]')

case "$SHELLHUB_EDITION" in
    community|enterprise|cloud) ;;
    *)
        echo "🚫 ERROR: invalid SHELLHUB_EDITION '$SHELLHUB_EDITION': must be community, enterprise, or cloud." >&2
        exit 1
        ;;
esac

env_var() {
    var=$1; default=$2
    edition=$SHELLHUB_EDITION
    (
        . ./.env 2>/dev/null
        [ "$edition" != "community" ] && [ -f "$cloud_dir/.env" ] \
            && . "$cloud_dir/.env"
        # shellcheck source=/dev/null # ENV_OVERRIDE names the file at run time
        [ -f "$env_override" ] && . "$env_override"
        eval "printf '%s' \"\${${var}:-${default}}\""
    )
}
