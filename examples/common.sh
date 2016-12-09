if ! [ -z $KUBECTL_NAME ]; then
    if [ -x "$(command -v kubectl)" ]; then
        KUBECTL_NAME='kubectl'
    else
        KUBECTL_NAME=${KUBECTL_NAME}
    fi
    export KUBECTL_NAME
fi
