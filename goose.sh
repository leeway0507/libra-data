#!/bin/bash

export GOOSE_DRIVER="postgres"
export GOOSE_DBSTRING="postgres://postgres:1q2w3e4r@localhost:5433/library"

# 서버 정보 설정
REMOTE_USER="ubuntu" # 원격 서버의 사용자 이름
REMOTE_HOST="134.185.119.195" # 원격 서버의 호스트 이름 또는 IP
REMOTE_PORT="5432" # PostgreSQL 기본 포트
LOCAL_PORT="5433" # 로컬에서 사용할 포트
SSH_KEY="~/.key/postgresql-oracle.key" 

cleanup_port() {
    echo "Checking if port ${LOCAL_PORT} is in use..."
    PID=$(lsof -ti :${LOCAL_PORT})
    if [ -n "$PID" ]; then
        echo "Port ${LOCAL_PORT} is in use by process $PID. Terminating process..."
        kill -9 $PID
        echo "Process $PID terminated."
    else
        echo "Port ${LOCAL_PORT} is not in use."
    fi
}

# SSH 터널 연결
start_tunnel() {
    cleanup_port  # 기존 포트 점유 프로세스 종료
    echo "Starting SSH tunnel..."
    ssh -N -L ${LOCAL_PORT}:localhost:${REMOTE_PORT} -i ${SSH_KEY} ${REMOTE_USER}@${REMOTE_HOST} &
    TUNNEL_PID=$!
    sleep 2
    if ps -p ${TUNNEL_PID} > /dev/null; then
        echo $TUNNEL_PID > ssh_tunnel.pid
        echo "Tunnel started successfully. PID: ${TUNNEL_PID}"
    else
        echo "Failed to start SSH tunnel."
        exit 1
    fi
}


# SSH 터널 해제
stop_tunnel() {
    if [ -f ssh_tunnel.pid ]; then
        TUNNEL_PID=$(cat ssh_tunnel.pid)
        echo "Stopping SSH tunnel with PID: ${TUNNEL_PID}"
        kill $TUNNEL_PID
        rm ssh_tunnel.pid
        echo "Tunnel stopped."
    else
        echo "No active tunnel found."
    fi
}

# SQL 업데이트
run_update() {
    if [ -d "pkg/db/migration" ]; then
        echo "Starting SQL update..."
        cd pkg/db/migration
        goose up
        goose status
        echo "SQL update completed successfully."
    else
        echo "Migration directory not found. Please check the path."
        exit 1
    fi
}

# 메인 스크립트 실행
echo "Starting tunnel setup..."
start_tunnel

echo "Tunnel setup completed."

echo "Starting SQL update..."
run_update

echo "Stopping tunnel..."
stop_tunnel

echo "Tunnel stopped and script completed."
