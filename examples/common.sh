KUBECTL_NAME=${KUBECTL_NAME:-}
APPC=${APPC:-"k8s-appcontroller"}

if [ -z "$KUBECTL_NAME" ]; then
    if [ -x "$(command -v kubectl)" ]; then
        KUBECTL_NAME='kubectl'
    fi
fi
echo "Using following kubectl - ${KUBECTL_NAME}"

function wait-appcontroller {
   echo "Waiting for pod $APPC to start running"
   wait-until "$KUBECTL_NAME get pods | grep Running | grep -q $APPC" 20
   echo "Waiting for tprs to register URLs"
   wait-until "$KUBECTL_NAME get definitions &> /dev/null" 5
   wait-until "$KUBECTL_NAME get dependency &> /dev/null" 5
}

function wait-until {
  local limit=$((SECONDS+$2))
  while [ $SECONDS -lt $limit ]; do
    eval $1 && return
    printf '.'
  done
  echo "FAILED: $1"
  exit 2
}
