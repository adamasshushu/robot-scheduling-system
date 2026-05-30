#!/usr/bin/env python3
"""
机器人底盘模拟器 — 通过 MQTT 接入调度系统
用法: python3 robot_sim.py <robot_code> [--broker localhost] [--port 1883]

MQTT 协议:
  发布 → robot/{robot_code}/status    状态上报 (每秒)
  发布 → robot/{robot_code}/heartbeat 心跳 (每5秒)
  发布 → robot/{robot_code}/alert     告警 (按需)
  订阅 ← robot/{robot_code}/command   接收指令
  发布 → robot/{robot_code}/response  指令回复
"""

import json
import time
import random
import argparse
import threading

try:
    import paho.mqtt.client as mqtt
except ImportError:
    print("请安装 paho-mqtt: pip3 install paho-mqtt")
    exit(1)


class RobotSimulator:
    def __init__(self, robot_code: str, broker: str = "localhost", port: int = 1883):
        self.robot_code = robot_code
        self.broker = broker
        self.port = port

        # 初始状态
        self.x = random.uniform(0, 100)
        self.y = random.uniform(0, 100)
        self.battery = 100
        self.status = "standby"  # standby / running / charging / fault
        self.speed = 0.0
        self.task_id = 0
        self.progress = 0
        self.temperature = 25.0
        self.uptime = 0
        self.running = True

        # MQTT 客户端
        self.client = mqtt.Client(client_id=f"sim-{robot_code}")
        self.client.on_connect = self._on_connect
        self.client.on_message = self._on_command

    def _on_connect(self, client, userdata, flags, rc):
        if rc == 0:
            print(f"✅ 已连接到 MQTT Broker {self.broker}:{self.port}")
            # 订阅指令
            topic = f"robot/{self.robot_code}/command"
            client.subscribe(topic, qos=1)
            print(f"  📡 订阅: {topic}")
        else:
            print(f"❌ MQTT 连接失败: code={rc}")

    def _on_command(self, client, userdata, msg):
        try:
            cmd = json.loads(msg.payload.decode())
            command = cmd.get("command", "")
            params = cmd.get("params", {})
            print(f"\n📥 收到指令: {command} {json.dumps(params, ensure_ascii=False)}")

            response = {"command_id": cmd.get("command",""), "status": "ok", "message": "", "timestamp": int(time.time()*1000)}

            if command == "start":
                self.status = "running"
                self.speed = params.get("speed", 1.0)
                response["message"] = "已启动"

            elif command == "stop":
                self.status = "standby"
                self.speed = 0.0
                response["message"] = "已停止"

            elif command == "pause":
                self.status = "standby"
                self.speed = 0.0
                response["message"] = "已暂停"

            elif command == "goto":
                self.status = "running"
                tx = params.get("target_x", self.x)
                ty = params.get("target_y", self.y)
                self.speed = params.get("speed", 1.0)
                response["message"] = f"前往 ({tx}, {ty})"
                # 模拟移动 (后台线程)
                threading.Thread(target=self._move_to, args=(tx, ty), daemon=True).start()

            elif command == "charge":
                self.status = "charging"
                self.speed = 0.0
                response["message"] = "开始充电"

            elif command == "upgrade":
                response["message"] = "升级指令已接收，将在空闲时执行"
                response["status"] = "ok"

            else:
                response["status"] = "error"
                response["message"] = f"未知指令: {command}"

            # 回复
            resp_topic = f"robot/{self.robot_code}/response"
            self.client.publish(resp_topic, json.dumps(response, ensure_ascii=False), qos=1)
            print(f"  📤 回复: {response['status']} {response['message']}")

        except Exception as e:
            print(f"  ❌ 指令处理异常: {e}")

    def _move_to(self, target_x: float, target_y: float):
        """模拟移动到目标点"""
        dist = ((target_x - self.x)**2 + (target_y - self.y)**2)**0.5
        steps = max(int(dist / (self.speed or 1)), 1)
        for i in range(steps):
            if self.status != "running":
                break
            time.sleep(0.5)
            self.x += (target_x - self.x) / (steps - i)
            self.y += (target_y - self.y) / (steps - i)
            self.progress = int((i+1) / steps * 100)
        if self.status == "running":
            self.status = "standby"
            self.speed = 0.0
            self.progress = 100
            print("  ✅ 到达目标")

    def _status_loop(self):
        """每秒上报状态"""
        while self.running:
            try:
                # 模拟状态变化
                if self.status == "running":
                    self.battery = max(0, self.battery - 0.05)
                    self.temperature = 25 + random.uniform(-1, 10)
                elif self.status == "charging":
                    self.battery = min(100, self.battery + 1.0)
                    self.temperature = 25 + random.uniform(-1, 3)

                report = {
                    "robot_code": self.robot_code,
                    "battery_pct": int(self.battery),
                    "location_x": round(self.x, 2),
                    "location_y": round(self.y, 2),
                    "map_zone": "A区-仓库" if self.x < 33 else "B区-装配线" if self.x < 66 else "C区-办公区",
                    "speed": round(self.speed, 1),
                    "status": self.status,
                    "task_id": self.task_id,
                    "progress_pct": self.progress,
                    "temperature": round(self.temperature, 1),
                    "light_mode": "auto",
                    "timestamp": int(time.time() * 1000),
                }

                topic = f"robot/{self.robot_code}/status"
                self.client.publish(topic, json.dumps(report, ensure_ascii=False), qos=1)

                time.sleep(1)
            except Exception as e:
                print(f"  ⚠️ 状态上报异常: {e}")
                time.sleep(3)

    def _heartbeat_loop(self):
        """每5秒发送心跳"""
        while self.running:
            try:
                self.uptime += 5
                hb = {
                    "robot_code": self.robot_code,
                    "version": "fw-v2.1.0",
                    "uptime": self.uptime,
                }
                topic = f"robot/{self.robot_code}/heartbeat"
                self.client.publish(topic, json.dumps(hb, ensure_ascii=False), qos=1)
                time.sleep(5)
            except Exception as e:
                print(f"  ⚠️ 心跳异常: {e}")
                time.sleep(5)

    def _alert_loop(self):
        """每30秒随机概率发送告警 (演示用)"""
        while self.running:
            time.sleep(30)
            if self.running and random.random() < 0.15:  # 15% 概率
                alert_types = [
                    ("warning", "battery", "电池电量低，建议充电"),
                    ("info", "sensor", "前方检测到障碍物，已减速"),
                    ("error", "motor", "左轮电机电流异常"),
                ]
                level, atype, msg = random.choice(alert_types)
                alert = {
                    "robot_code": self.robot_code,
                    "level": level,
                    "type": atype,
                    "message": msg,
                    "data": "",
                    "timestamp": int(time.time() * 1000),
                }
                topic = f"robot/{self.robot_code}/alert"
                self.client.publish(topic, json.dumps(alert, ensure_ascii=False), qos=1)
                print(f"\n🚨 发送告警: [{level}] {msg}")

    def run(self):
        """启动模拟器"""
        print(f"""
╔══════════════════════════════════════════╗
║   🤖 机器人底盘模拟器                   ║
║   编号: {self.robot_code:<30} ║
║   MQTT: {self.broker}:{self.port:<29} ║
╚══════════════════════════════════════════╝
""")
        self.client.connect(self.broker, self.port, 60)
        self.client.loop_start()

        # 启动各个后台线程
        threading.Thread(target=self._status_loop, daemon=True).start()
        threading.Thread(target=self._heartbeat_loop, daemon=True).start()
        threading.Thread(target=self._alert_loop, daemon=True).start()

        print("▶  模拟器运行中... 等待指令\n")
        print("  可用指令:")
        print("    系统 → POST /api/v1/mqtt/command?robot_code=RB-A001")
        print("    JSON → {\"command\":\"goto\",\"params\":{\"target_x\":80,\"target_y\":60}}")
        print()

        try:
            while True:
                time.sleep(1)
        except KeyboardInterrupt:
            print("\n🛑 模拟器关闭...")
            self.running = False
            self.client.loop_stop()
            self.client.disconnect()


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="机器人底盘模拟器")
    parser.add_argument("robot_code", help="机器人编号 (如 RB-A001)")
    parser.add_argument("--broker", default="localhost", help="MQTT Broker 地址 (默认: localhost)")
    parser.add_argument("--port", type=int, default=1883, help="MQTT 端口 (默认: 1883)")
    args = parser.parse_args()

    sim = RobotSimulator(args.robot_code, args.broker, args.port)
    sim.run()
