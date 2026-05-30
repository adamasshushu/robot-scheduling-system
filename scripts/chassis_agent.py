#!/usr/bin/env python3
"""
机器人底盘 Agent — Ubuntu ARM
部署路径: /opt/robot-agent/agent.py
"""

import json
import time
import socket
import subprocess
import os
import sys
import threading
import logging

try:
    import paho.mqtt.client as mqtt
except ImportError:
    print("请先安装: pip3 install paho-mqtt")
    sys.exit(1)

# ── 配置 ──────────────────────────────────────────
ROBOT_CODE  = os.getenv("ROBOT_CODE", "RBT-A001")
BROKER_HOST = os.getenv("MQTT_BROKER", "192.168.201.230")
BROKER_PORT = int(os.getenv("MQTT_PORT", "1883"))
MQTT_USER   = os.getenv("MQTT_USER", "")
MQTT_PASS   = os.getenv("MQTT_PASS", "")

# ── 日志 ──
logging.basicConfig(level=logging.INFO, format="%(asctime)s [%(levelname)s] %(message)s")
log = logging.getLogger("agent")

# ── MQTT ──────────────────────────────────────────
client = mqtt.Client(client_id=f"chassis-{ROBOT_CODE}")
if MQTT_USER:
    client.username_pw_set(MQTT_USER, MQTT_PASS)

# ── 底盘状态 ──────────────────────────────────────
state = {
    "x": 0.0, "y": 0.0,
    "battery": 100.0,
    "speed": 0.0,
    "status": "standby",  # standby / running / charging / fault
    "task_id": 0,
    "progress": 0,
    "temperature": 35.0,
    "uptime": 0,
}
running = True

# ── 硬件数据采集 — 根据实际硬件修改 ────────────────

def read_battery():
    """读取电池电量 — 替换为实际 ADC/I2C 代码"""
    # 示例: 读取 /sys/class/power_supply/BAT0/capacity
    try:
        with open("/sys/class/power_supply/BAT1/capacity", "r") as f:
            return float(f.read().strip())
    except:
        pass
    # 没有电池则模拟下降
    return max(0, state["battery"] - 0.01)

def read_temperature():
    """读取温度 — 替换为实际传感器代码"""
    # 示例: /sys/class/thermal/thermal_zone0/temp (单位: 毫°C)
    try:
        for i in range(5):
            path = f"/sys/class/thermal/thermal_zone{i}/temp"
            if os.path.exists(path):
                with open(path) as f:
                    return float(f.read().strip()) / 1000.0
    except:
        pass
    return 35.0

def read_position():
    """读取位置 — 替换为里程计/SLAM/IMU输出"""
    # 示例：读取 ROS topic 或 串口里程计数据
    # return odom_x, odom_y
    return state["x"], state["y"]

def read_speed():
    """读取速度 — 替换为编码器数据"""
    return state["speed"]

# ── MQTT 回调 ─────────────────────────────────────

def on_connect(client, userdata, flags, rc):
    if rc == 0:
        log.info(f"✅ 已连接调度系统 {BROKER_HOST}:{BROKER_PORT}")
        client.subscribe(f"robot/{ROBOT_CODE}/command", qos=1)
        log.info(f"📡 监听指令: robot/{ROBOT_CODE}/command")
    else:
        log.error(f"❌ 连接失败 code={rc}，15s 后重试...")
        time.sleep(15)
        client.reconnect()

def on_command(client, userdata, msg):
    """处理调度系统下发的指令"""
    try:
        cmd = json.loads(msg.payload.decode())
        command = cmd.get("command", "")
        params = cmd.get("params", {})
        log.info(f"📥 指令: {command} {json.dumps(params, ensure_ascii=False)}")

        resp = {"command_id": command, "status": "ok", "message": "", "timestamp": int(time.time()*1000)}

        if command == "start":
            state["status"] = "running"
            state["speed"] = params.get("speed", 1.0)
            resp["message"] = "已启动"

        elif command == "stop":
            state["status"] = "standby"
            state["speed"] = 0.0
            resp["message"] = "已停止"

        elif command == "goto":
            state["status"] = "running"
            state["speed"] = params.get("speed", 1.0)
            tx, ty = params.get("target_x", 0), params.get("target_y", 0)
            resp["message"] = f"前往 ({tx},{ty})"
            # ── 替换为实际底盘运动控制 ──────────
            # 例如: 通过串口/ROS发送目标坐标
            # subprocess.run(["rostopic", "pub", "/move_base_simple/goal", ...])
            threading.Thread(target=_simulate_move, args=(tx, ty), daemon=True).start()

        elif command == "charge":
            state["status"] = "charging"
            state["speed"] = 0.0
            resp["message"] = "开始充电"

        elif command == "upgrade":
            resp["message"] = "OTA 升级指令已接收"
            # ── 替换为实际 OTA 流程 ──
            # subprocess.Popen(["systemctl", "restart", "robot-firmware"])

        else:
            resp["status"] = "error"
            resp["message"] = f"未知指令: {command}"

        # 回复
        client.publish(f"robot/{ROBOT_CODE}/response", json.dumps(resp, ensure_ascii=False), qos=1)
        log.info(f"  📤 {resp['status']}: {resp['message']}")

    except Exception as e:
        log.error(f"指令处理异常: {e}")

# ── 模拟移动（替换为实际运动控制后删除）─────────
def _simulate_move(tx, ty):
    steps = int(max(abs(tx - state["x"]), abs(ty - state["y"])) / (state["speed"] or 1))
    for i in range(max(steps, 1)):
        if state["status"] != "running": break
        time.sleep(0.5)
        state["x"] += (tx - state["x"]) / (steps - i + 1)
        state["y"] += (ty - state["y"]) / (steps - i + 1)
        state["progress"] = int((i+1) / max(steps,1) * 100)
    if state["status"] == "running":
        state["status"] = "standby"
        state["speed"] = 0.0
        state["progress"] = 100

# ── 状态上报 ─────────────────────────────────────
def status_loop():
    while running:
        try:
            state["x"], state["y"] = read_position()
            state["battery"] = read_battery()
            state["speed"] = read_speed()
            state["temperature"] = read_temperature()

            payload = {
                "robot_code": ROBOT_CODE,
                "battery_pct": int(state["battery"]),
                "location_x": round(state["x"], 2),
                "location_y": round(state["y"], 2),
                "map_zone": "",
                "speed": round(state["speed"], 1),
                "status": state["status"],
                "task_id": state["task_id"],
                "progress_pct": state["progress"],
                "temperature": round(state["temperature"], 1),
                "light_mode": "auto",
                "timestamp": int(time.time() * 1000),
            }
            client.publish(f"robot/{ROBOT_CODE}/status", json.dumps(payload, ensure_ascii=False), qos=1)
            time.sleep(1)
        except Exception as e:
            log.error(f"状态上报异常: {e}")
            time.sleep(3)

def heartbeat_loop():
    while running:
        try:
            state["uptime"] += 5
            client.publish(f"robot/{ROBOT_CODE}/heartbeat", json.dumps({
                "robot_code": ROBOT_CODE,
                "version": os.getenv("FW_VERSION", "fw-v1.0"),
                "uptime": state["uptime"],
            }, ensure_ascii=False), qos=1)
            time.sleep(5)
        except:
            time.sleep(5)

# ── 异常告警 ─────────────────────────────────────
def alert_loop():
    while running:
        time.sleep(10)
        if state["battery"] < 15:
            client.publish(f"robot/{ROBOT_CODE}/alert", json.dumps({
                "robot_code": ROBOT_CODE,
                "level": "warning", "type": "battery",
                "message": f"电量过低: {int(state['battery'])}%",
                "timestamp": int(time.time()*1000),
            }, ensure_ascii=False))
        if state["temperature"] > 70:
            client.publish(f"robot/{ROBOT_CODE}/alert", json.dumps({
                "robot_code": ROBOT_CODE,
                "level": "error", "type": "system",
                "message": f"温度过高: {state['temperature']}°C",
                "timestamp": int(time.time()*1000),
            }, ensure_ascii=False))

# ── 主入口 ───────────────────────────────────────
def main():
    print(f"""
╔══════════════════════════════════════════╗
║   🤖 机器人底盘 Agent (ARM)              ║
║   编号: {ROBOT_CODE:<30} ║
║   调度: {BROKER_HOST}:{BROKER_PORT:<26} ║
╚══════════════════════════════════════════╝
""")
    client.on_connect = on_connect
    client.on_message = on_command

    while True:
        try:
            client.connect(BROKER_HOST, BROKER_PORT, keepalive=30)
            break
        except Exception as e:
            log.error(f"连接失败: {e}，10s 后重试...")
            time.sleep(10)

    client.loop_start()
    threading.Thread(target=status_loop, daemon=True).start()
    threading.Thread(target=heartbeat_loop, daemon=True).start()
    threading.Thread(target=alert_loop, daemon=True).start()

    log.info("🚀 Agent 运行中...")
    try:
        while running: time.sleep(1)
    except KeyboardInterrupt:
        log.info("🛑 正在关闭...")
        client.loop_stop()
        client.disconnect()

if __name__ == "__main__":
    main()
