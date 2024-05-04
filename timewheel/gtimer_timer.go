package timewheel

import (
	"fmt"
	"time"

	"github.com/tea4go/gh/timewheel/gtype"
)

// Timer is a Hierarchical Timing Wheel manager for timing jobs.

// Wheel is a slot wrapper for timing job install and uninstall.
type Timer struct {
	status     *gtype.Int //定时器状态
	wheels     []*wheel   //分层时间轮对象
	length     int        //分层层数
	number     int        //每一层Slot Number
	intervalMs int64      //最小时间刻度(毫秒)
}

//单层时间轮
type wheel struct {
	timer      *Timer        //所属定时器
	level      int           //所属分层索引号
	slots      []*gtype.List //所有的定时任务项，按照Slot的数量分组，提高并发效率。
	slot_size  int64         //时间轮槽数=len(slots)
	ticks      *gtype.Int64  //当前时间轮已转动的刻度数量，一个刻度是interval。
	intervalMs int64         //间隔（以毫秒为单位），即一个时隙的持续时间。
	totalMs    int64         //一周大小，总持续时间（以毫秒为单位）=number*interval
	create     time.Time     //创建的时间。
}

func NewTimer() *Timer {
	return NewTimerPlus(10, 50*time.Millisecond, 6)
}

// 创建并返回Hierarchical Timing Wheel设计的计时器。slot - 时间轮槽数; interval - 间隔时间  level - 分层大小
func NewTimerPlus(slot int, interval time.Duration, level int) *Timer {
	if slot > 29 {
		level = 29
	}
	if level > 8 {
		level = 8
	}
	t := &Timer{
		status:     gtype.NewInt(STATUS_RUNNING),
		wheels:     make([]*wheel, level),
		length:     level,
		number:     slot,
		intervalMs: interval.Nanoseconds() / 1e6, //间隔ms
	}
	for i := 0; i < level; i++ {
		if i > 0 { //上一个轮盘的总刻度为下一个轮盘的单位刻度
			n := time.Duration(t.wheels[i-1].totalMs) * time.Millisecond
			w := t.newWheel(i, slot, n)
			t.wheels[i] = w
			t.wheels[i-1].addEntry("proceed", true, n, w.proceed, nil, false, DEFAULT_TIMES, STATUS_READY)
		} else {
			t.wheels[i] = t.newWheel(i, slot, interval)
		}
	}

	t.wheels[0].start()
	return t
}

// 创建并返回单个轮子.
func (t *Timer) newWheel(level int, slot int, interval time.Duration) *wheel {
	w := &wheel{
		timer:      t,
		level:      level,
		slots:      make([]*gtype.List, slot),
		slot_size:  int64(slot),
		ticks:      gtype.NewInt64(),
		totalMs:    int64(slot) * interval.Nanoseconds() / 1e6,
		create:     time.Now(),
		intervalMs: interval.Nanoseconds() / 1e6,
	}
	for i := int64(0); i < w.slot_size; i++ {
		w.slots[i] = gtype.NewList(true)
	}
	return w
}

func (t *Timer) String() string {
	out_str := "-= 当前定时器 =-\n"
	for _, v := range t.wheels {
		out_str += v.String()
	}
	return out_str
}

//添加定时任务
func (t *Timer) AddNow(name string, interval time.Duration, job JobFunc, jobp interface{}) *Entry {
	go job(name, time.Now(), interval, jobp)
	return t.doAddEntry(name, interval, job, jobp, false, DEFAULT_TIMES, STATUS_READY)
}

//添加定时任务
func (t *Timer) Add(name string, interval time.Duration, job JobFunc, jobp interface{}) *Entry {
	return t.doAddEntry(name, interval, job, jobp, false, DEFAULT_TIMES, STATUS_READY)
}

//添加单例定时任务，即同时只能有一个该任务正在运行；
func (t *Timer) AddSingleton(name string, interval time.Duration, job JobFunc, jobp interface{}) *Entry {
	return t.doAddEntry(name, interval, job, jobp, true, DEFAULT_TIMES, STATUS_READY)
}

//添加只运行一次的定时任务，当运行一次数后该定时任务自动销毁；
func (t *Timer) AddOnce(name string, interval time.Duration, job JobFunc, jobp interface{}) *Entry {
	return t.doAddEntry(name, interval, job, jobp, true, 1, STATUS_READY)
}

//添加运行指定次数的定时任务，当运行times次数后该定时任务自动销毁；
func (t *Timer) AddTimes(name string, interval time.Duration, times int, job JobFunc, jobp interface{}) *Entry {
	return t.doAddEntry(name, interval, job, jobp, true, times, STATUS_READY)
}

func (t *Timer) DelayAdd(name string, delay time.Duration, interval time.Duration, job JobFunc, jobp interface{}) {
	t.AddOnce(name+"(Delay)", delay, func(name1 string, create time.Time, interval time.Duration, jobp1 interface{}) {
		t.Add(name, interval, job, jobp)
	}, jobp)
}

func (t *Timer) DelayAddSingleton(name string, delay time.Duration, interval time.Duration, job JobFunc, jobp interface{}) {
	t.AddOnce(name+"(Delay)", delay, func(name1 string, create time.Time, interval time.Duration, jobp1 interface{}) {
		t.AddSingleton(name, interval, job, jobp)
	}, jobp)
}

func (t *Timer) DelayAddOnce(name string, delay time.Duration, interval time.Duration, job JobFunc, jobp interface{}) {
	t.AddOnce(name+"(Delay)", delay, func(name1 string, create time.Time, interval time.Duration, jobp1 interface{}) {
		t.AddOnce(name, interval, job, jobp)
	}, jobp)
}

func (t *Timer) DelayAddTimes(name string, delay time.Duration, interval time.Duration, times int, job JobFunc, jobp interface{}) {
	t.AddOnce(name+"(Delay)", delay, func(name1 string, create time.Time, interval time.Duration, jobp1 interface{}) {
		t.AddTimes(name, interval, times, job, jobp)
	}, jobp)
}

func (t *Timer) Start() {
	t.status.Set(STATUS_RUNNING)
}

func (t *Timer) Stop() {
	t.status.Set(STATUS_STOPPED)
}

func (t *Timer) Close() {
	t.status.Set(STATUS_CLOSED)
}

func (t *Timer) doAddEntry(name string, interval time.Duration, job JobFunc, jobp interface{}, singleton bool, times int, status int) *Entry {
	index := t.getLevelByIntervalMs(interval.Nanoseconds() / 1e6)
	return t.wheels[index].addEntry(name, false, interval, job, jobp, singleton, times, status)
}

func (t *Timer) doAddEntryByParent(interval int64, parent *Entry) *Entry {
	return t.wheels[t.getLevelByIntervalMs(interval)].addEntryByParent(interval, parent)
}

// getLevelByIntervalMs calculates and returns the level of timer wheel with given milliseconds.
// 根据intervalMs计算添加的分层索引
func (t *Timer) getLevelByIntervalMs(intervalMs int64) int {
	pos, cmp := t.binSearchIndex(intervalMs)
	switch cmp {
	case 0:
		fallthrough
	// intervalMs比最后匹配值小
	case -1:
		i := pos
		for ; i > 0; i-- {
			if intervalMs > t.wheels[i].intervalMs && intervalMs <= t.wheels[i].totalMs {
				return i
			}
		}
		return i

	// intervalMs比最后匹配值大
	case 1:
		i := pos
		for ; i < t.length-1; i++ {
			if intervalMs > t.wheels[i].intervalMs && intervalMs <= t.wheels[i].totalMs {
				return i
			}
		}
		return i
	}
	return 0
}

// 二分查找当前任务可以添加的时间轮对象索引
func (t *Timer) binSearchIndex(n int64) (index int, result int) {
	min := 0
	max := t.length - 1
	mid := 0
	cmp := -2
	for min <= max {
		mid = int((min + max) / 2)
		switch {
		case t.wheels[mid].intervalMs == n:
			cmp = 0
		case t.wheels[mid].intervalMs > n:
			cmp = -1
		case t.wheels[mid].intervalMs < n:
			cmp = 1
		}
		switch cmp {
		case -1:
			max = mid - 1
		case 1:
			min = mid + 1
		case 0:
			return mid, cmp
		}
	}
	return mid, cmp
}

//执行时间轮刻度逻辑
func (w *wheel) start() {
	go func() {
		ticker := time.NewTicker(time.Duration(w.intervalMs) * time.Millisecond)
		for {
			select {
			case <-ticker.C:
				switch w.timer.status.Val() {
				case STATUS_RUNNING:
					w.proceed("proceed", w.create, time.Duration(w.intervalMs)*time.Millisecond, nil)
				case STATUS_STOPPED:
				case STATUS_CLOSED:
					ticker.Stop()
					return
				}

			}
		}
	}()
}

func (w *wheel) GetCount() int {
	count := 0
	for _, v := range w.slots {
		item := v.Top()
		for item != nil {
			entry := item.Value.(*Entry)
			if !entry.sysflag {
				count += 1
			}
			item = item.Next()
		}
	}
	return count
}

func (w *wheel) String() string {
	out_str := fmt.Sprintf("层级=%d, 槽数=%d, 任务=%-8d,  间隔(%12dms), 一周(%13dms), 创建时间=%s\n",
		w.level, w.slot_size, w.GetCount(), w.intervalMs, w.totalMs, w.create.Format("2006-01-02 15:04:05"))
	count := w.GetCount()
	if count >= 1 && count <= 10 {
		for k, v := range w.slots {
			item := v.Top()
			for item != nil {
				entry := item.Value.(*Entry)
				if !entry.sysflag {
					out_str += fmt.Sprintf("= %d %s\n", k, entry.String())
				}
				item = item.Next()
			}
		}
	}
	return out_str
}

// 执行时间轮刻度逻辑
func (w *wheel) proceed(name string, create time.Time, interval time.Duration, param interface{}) {
	n := w.ticks.Add(1)
	l := w.slots[int(n%w.slot_size)]
	length := l.Len()
	if length > 0 {
		go func(l *gtype.List, nowTicks int64) {
			entry := (*Entry)(nil)
			nowMs := time.Now().UnixNano() / 1e6

			for i := length; i > 0; i-- {
				if v := l.PopFront(); v == nil {
					break
				} else {
					entry = v.(*Entry)
				}

				//是否满足运行条件
				runnable, addable := entry.check(nowTicks, nowMs)
				if runnable {
					//异步执行运行
					entry.runTimes.Add(1)
					go func(entry *Entry) {
						//执行完任务，把状态重置为Ready状态
						defer func() {
							if err := recover(); err != nil {
								if "exit" != PANIC_EXIT {
									panic(err)
								} else {
									fmt.Println("执行报错：", err, entry.jobp)
									entry.Close()
								}
							}
							if entry.Status() == STATUS_RUNNING { //todo:entry将running改为ready?
								entry.SetStatus(STATUS_READY)
							}
						}()

						//开始执行任务
						if !entry.sysflag {
							a := fmt.Sprintf("%.0f", float64((time.Now().UnixNano()-entry.createTime.UnixNano())/1e6)/float64(entry.runTimes.Val())/1000.0)
							b := fmt.Sprintf("%.0f", float64(entry.interval/1e6)/1000.0)
							if a != b {
								fmt.Printf("开始执行任务[%d] %s != %s\n", entry.runTimes.Val(), a, b)
							}
						}
						entry.job(entry.name, entry.createTime, entry.interval, entry.jobp)
					}(entry)
				}

				// 是否继续添运行，滚动任务
				if addable {
					//优先从chird time wheel开始添加
					entry.wheel.timer.doAddEntryByParent(entry.rawIntervalMs, entry)
				}
			}
		}(l, n)
	}
}

//将计时作业添加到方向盘。
func (w *wheel) addEntry(name string, sysflag bool, interval time.Duration, job JobFunc, jobp interface{}, singleton bool, times int, status int) *Entry {
	if times <= 0 {
		times = DEFAULT_TIMES
	}
	ms := interval.Nanoseconds() / 1e6 //间隔ms
	num := ms / w.intervalMs
	if num <= 0 {
		// 如果安装的任务间隔小于时间轮刻度，考虑任务安装的时间，那么将会在下一刻度被执行
		num = 1
	}

	nowMs := time.Now().UnixNano() / 1e6
	ticks := w.ticks.Val()
	entry := &Entry{
		wheel:         w,
		name:          name,
		sysflag:       sysflag,
		job:           job,
		jobp:          jobp,
		times:         gtype.NewInt(times),
		status:        gtype.NewInt(status),
		createTime:    time.Now(),
		runTimes:      gtype.NewInt(),
		interval:      interval,
		create:        ticks,
		interval_num:  num,
		singleton:     gtype.NewBool(singleton),
		createMs:      nowMs,
		intervalMs:    ms,
		rawIntervalMs: ms,
	}

	// 安装任务
	slot_index := (ticks + num) % w.slot_size
	// if !sysflag {
	// 	fmt.Printf("新建定时任务：%s, 参数：Level=%d, Slot=%d\n", name, w.level, slot_index)
	// }
	w.slots[slot_index].PushBack(entry)
	return entry
}

// 创建定时任务，给定父级Entry，间隔参数为毫秒数
func (w *wheel) addEntryByParent(interval int64, parent *Entry) *Entry {
	num := interval / w.intervalMs
	if num <= 0 {
		num = 1
	}
	nowMs := time.Now().UnixNano() / 1e6
	ticks := w.ticks.Val()
	entry := &Entry{
		wheel:         w,
		name:          parent.name,
		sysflag:       parent.sysflag,
		job:           parent.job,
		jobp:          parent.jobp,
		times:         parent.times,
		status:        parent.status,
		singleton:     parent.singleton,
		createTime:    parent.createTime,
		runTimes:      parent.runTimes,
		interval:      parent.interval,
		create:        ticks,
		interval_num:  num,
		createMs:      nowMs,
		intervalMs:    interval,
		rawIntervalMs: parent.rawIntervalMs,
	}

	// 安装任务
	slot_index := (ticks + num) % w.slot_size
	// if !parent.sysflag {
	// 	fmt.Printf("调整定时任务：%s, 参数：Level=%d, Slot=%d\n", parent.name, w.level, slot_index)
	// }
	w.slots[slot_index].PushBack(entry)
	return entry
}
