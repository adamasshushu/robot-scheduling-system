package service

import (
	"math"
	"sort"

	"gorm.io/gorm"
	"github.com/adamasshushu/robot-scheduling-system/internal/model"
)

// Scheduler 智能调度引擎
type Scheduler struct {
	db *gorm.DB
}

func NewScheduler(db *gorm.DB) *Scheduler {
	return &Scheduler{db: db}
}

// Candidate 候选机器人（用于排序）
type Candidate struct {
	Robot    model.Robot
	Score    float64 // 综合评分，越高越好
	Distance float64 // 距离任务目标的距离
}

// AutoAssign 智能指派 — 多因子评分算法
//
// 因子:
//   1. 基础条件：standby 状态 + 电量 > 20%
//   2. 电量因子：电量越高越好 (权重 0.3)
//   3. 距离因子：距离目标越近越好 (权重 0.4)
//   4. 负载因子：当前任务数越少越好 (权重 0.3)
func (s *Scheduler) AutoAssign(task *model.Task) (*model.Robot, error) {
	// 1. 获取可用机器人
	var robots []model.Robot
	s.db.Where("status = ? AND battery_pct > ?", "standby", 20.0).
		Find(&robots)

	if len(robots) == 0 {
		return nil, nil // 无可用机器人
	}

	// 2. 收集每个机器人的当前任务数
	taskCounts := make(map[uint]int)
	for _, r := range robots {
		var count int64
		s.db.Model(&model.Task{}).
			Where("robot_id = ? AND status IN ?", r.ID, []string{"assigned", "running"}).
			Count(&count)
		taskCounts[r.ID] = int(count)
	}

	// 3. 计算评分
	candidates := make([]Candidate, 0, len(robots))
	for _, r := range robots {
		dist := s.distance(r.LocationX, r.LocationY, task.TargetX, task.TargetY)

		batteryScore := r.BatteryPct / 100.0                              // 0-1，越高越好
		distanceScore := 1.0 / (1.0 + dist/100.0)                          // 0-1，越近越好
		loadScore := 1.0 / (1.0 + float64(taskCounts[r.ID]))               // 0-1，越少越好

		score := batteryScore*0.3 + distanceScore*0.4 + loadScore*0.3

		candidates = append(candidates, Candidate{
			Robot:    r,
			Score:    math.Round(score*1000) / 1000,
			Distance: math.Round(dist*100) / 100,
		})
	}

	// 4. 排序取最优
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Score > candidates[j].Score
	})

	return &candidates[0].Robot, nil
}

// distance 欧几里得距离
func (s *Scheduler) distance(x1, y1 float64, targetX, targetY *float64) float64 {
	if targetX == nil || targetY == nil {
		return 999999 // 无法计算距离，给极低分
	}
	dx := x1 - *targetX
	dy := y1 - *targetY
	return math.Sqrt(dx*dx + dy*dy)
}

// GetCandidates 获取所有候选机器人及其评分（用于调试/展示）
func (s *Scheduler) GetCandidates(task *model.Task) ([]Candidate, error) {
	var robots []model.Robot
	s.db.Where("status = ? AND battery_pct > ?", "standby", 20.0).
		Find(&robots)

	taskCounts := make(map[uint]int)
	for _, r := range robots {
		var count int64
		s.db.Model(&model.Task{}).
			Where("robot_id = ? AND status IN ?", r.ID, []string{"assigned", "running"}).
			Count(&count)
		taskCounts[r.ID] = int(count)
	}

	candidates := make([]Candidate, 0, len(robots))
	for _, r := range robots {
		dist := s.distance(r.LocationX, r.LocationY, task.TargetX, task.TargetY)
		batteryScore := r.BatteryPct / 100.0
		distanceScore := 1.0 / (1.0 + dist/100.0)
		loadScore := 1.0 / (1.0 + float64(taskCounts[r.ID]))
		score := batteryScore*0.3 + distanceScore*0.4 + loadScore*0.3

		candidates = append(candidates, Candidate{
			Robot:    r,
			Score:    math.Round(score*1000) / 1000,
			Distance: math.Round(dist*100) / 100,
		})
	}
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Score > candidates[j].Score
	})
	return candidates, nil
}
