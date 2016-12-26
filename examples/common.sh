KUBECTL_NAME=${KUBECTL_NAME:-}
if [ -z "$KUBECTL_NAME" ]; then
    if [ -x "$(command -v kubectl)" ]; then
        KUBECTL_NAME='kubectl'
    fi
    if [ -x "$(command -v kubectl.sh)" ]; then
        KUBECTL_NAME='kubectl.sh'
    fi
fi
echo "Using following kubectl - ${KUBECTL_NAME}"
