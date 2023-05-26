#!/bin/sh
if [ "$(id -u)" != "0" ]; then
    exec sudo bash "$0" "$@"
fi

is_service_exists() {
    x="$1"
    if systemctl status "${x}" 2>/dev/null | grep -Fq "Active:"; then
        return 0
    else
        return 1
    fi
}

INSTALL_PATH=/opt/movies
NGINX_CONF_PATH=/etc/nginx/conf.d

# Build software
go build -o movies
ret_code=$?
if [ $ret_code != 0 ]; then
    printf "Error: [%d] when building executable. Check that you have go tools installed." $ret_code
    exit $ret_code
fi

# Check if needed files exist
if [ -f .env ] && [ -f movies ] && [ -f movies.service ]; then
    # Check if we upgrade or install for first time
    if is_service_exists 'movies.service'; then
        systemctl stop movies.service
        cp movies $INSTALL_PATH
        cp .env $INSTALL_PATH
        systemctl start movies.service
    else
        mkdir -p $INSTALL_PATH
        cp movies $INSTALL_PATH
        cp .env $INSTALL_PATH
        cp movies.service /usr/lib/systemd/system
        systemctl start movies.service
        systemctl enable movies.service
	cp movies.conf $NGINX_CONF_PATH
        nginx -s reload
    fi
else
    echo "Not all needed files found. Installation failed."
    exit 1
fi
