
if [ -f ssh_tunnel.pid ]; then
    TUNNEL_PID=$(cat ssh_tunnel.pid)
    echo "Stopping SSH tunnel with PID: ${TUNNEL_PID}"
    kill $TUNNEL_PID
    rm ssh_tunnel.pid
    echo "Tunnel stopped."
else
    echo "No active tunnel found."
fi
