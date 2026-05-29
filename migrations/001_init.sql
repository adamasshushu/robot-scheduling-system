-- ── 机器人调度系统 — 数据库初始化脚本 ──

-- 创建 TimescaleDB 扩展
CREATE EXTENSION IF NOT EXISTS timescaledb;

-- ── 机器人表 ──
CREATE TABLE IF NOT EXISTS robots (
    id              BIGSERIAL PRIMARY KEY,
    robot_code      VARCHAR(32) NOT NULL UNIQUE,
    model           VARCHAR(64) NOT NULL,
    name            VARCHAR(128),
    battery_pct     DECIMAL(5,2) DEFAULT 100.00,
    battery_status  VARCHAR(16) DEFAULT 'normal',
    location_x      DECIMAL(10,3),
    location_y      DECIMAL(10,3),
    location_theta  DECIMAL(8,3),
    map_zone        VARCHAR(64),
    speed           DECIMAL(6,2) DEFAULT 0.00,
    light_mode      VARCHAR(16) DEFAULT 'off',
    light_color     VARCHAR(16) DEFAULT 'white',
    light_brightness INT DEFAULT 100,
    status          VARCHAR(16) DEFAULT 'standby',
    comm_status     VARCHAR(16) DEFAULT 'online',
    last_heartbeat  TIMESTAMP WITH TIME ZONE,
    ip_address      INET,
    firmware_ver    VARCHAR(32),
    max_payload     DECIMAL(8,2),
    max_speed       DECIMAL(6,2),
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at      TIMESTAMP WITH TIME ZONE
);

-- ── 任务表 ──
CREATE TABLE IF NOT EXISTS tasks (
    id              BIGSERIAL PRIMARY KEY,
    task_code       VARCHAR(32) NOT NULL UNIQUE,
    task_type       VARCHAR(32) NOT NULL,
    robot_id        BIGINT REFERENCES robots(id),
    priority        INT DEFAULT 5,
    status          VARCHAR(16) DEFAULT 'pending',
    source_location VARCHAR(128),
    target_location VARCHAR(128) NOT NULL,
    source_x        DECIMAL(10,3),
    source_y        DECIMAL(10,3),
    target_x        DECIMAL(10,3),
    target_y        DECIMAL(10,3),
    description     TEXT,
    expected_start  TIMESTAMP WITH TIME ZONE,
    expected_end    TIMESTAMP WITH TIME ZONE,
    actual_start    TIMESTAMP WITH TIME ZONE,
    actual_end      TIMESTAMP WITH TIME ZONE,
    progress_pct    INT DEFAULT 0,
    created_by      BIGINT,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ── 告警表 ──
CREATE TABLE IF NOT EXISTS alerts (
    id              BIGSERIAL PRIMARY KEY,
    alert_type      VARCHAR(32) NOT NULL,
    severity        VARCHAR(16) NOT NULL,
    robot_id        BIGINT REFERENCES robots(id),
    task_id         BIGINT REFERENCES tasks(id),
    title           VARCHAR(256) NOT NULL,
    content         TEXT,
    status          VARCHAR(16) DEFAULT 'unack',
    acknowledged_by BIGINT,
    acknowledged_at TIMESTAMP WITH TIME ZONE,
    resolved_at     TIMESTAMP WITH TIME ZONE,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ── 遥测时序表（TimescaleDB hypertable）──
CREATE TABLE IF NOT EXISTS robot_telemetry (
    time            TIMESTAMPTZ NOT NULL,
    robot_id        BIGINT NOT NULL,
    battery_pct     DECIMAL(5,2),
    location_x      DECIMAL(10,3),
    location_y      DECIMAL(10,3),
    location_theta  DECIMAL(8,3),
    speed           DECIMAL(6,2),
    light_mode      VARCHAR(16),
    light_color     VARCHAR(16),
    light_brightness INT,
    status          VARCHAR(16),
    comm_status     VARCHAR(16)
);

SELECT create_hypertable('robot_telemetry', 'time', if_not_exists => TRUE, chunk_time_interval => INTERVAL '1 day');

-- ── 用户表 ──
CREATE TABLE IF NOT EXISTS users (
    id              BIGSERIAL PRIMARY KEY,
    username        VARCHAR(64) NOT NULL UNIQUE,
    password        VARCHAR(256) NOT NULL,
    role            VARCHAR(32) DEFAULT 'operator',
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at      TIMESTAMP WITH TIME ZONE
);

-- ── 索引 ──
CREATE INDEX IF NOT EXISTS idx_robots_status ON robots(status);
CREATE INDEX IF NOT EXISTS idx_robots_comm_status ON robots(comm_status);
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_robot_id ON tasks(robot_id);
CREATE INDEX IF NOT EXISTS idx_alerts_status ON alerts(status);
CREATE INDEX IF NOT EXISTS idx_alerts_robot_id ON alerts(robot_id);
CREATE INDEX IF NOT EXISTS idx_telemetry_robot_time ON robot_telemetry(robot_id, time DESC);

-- ── 默认管理员 ──
INSERT INTO users (username, password, role) 
VALUES ('admin', 'admin123', 'admin')
ON CONFLICT (username) DO NOTHING;

-- ── 示例机器人 ──
INSERT INTO robots (robot_code, model, name, location_x, location_y, map_zone, status)
VALUES 
    ('RBT-001', 'AGV-2000', '搬运一号', 10.0, 20.0, 'A区-仓库', 'standby'),
    ('RBT-002', 'AGV-2000', '搬运二号', 50.0, 30.0, 'B区-装配线', 'standby'),
    ('RBT-003', 'AMR-100', '巡逻一号', 100.0, 80.0, 'C区-办公区', 'standby')
ON CONFLICT (robot_code) DO NOTHING;
