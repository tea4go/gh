package timewheel

import (
	"fmt"
	"time"

	"github.com/tea4go/gh/timewheel/gtype"
)

// Entry is the timing job entry to wheel.
type Entry struct {
	wheel         *wheel        //所属时间轮
	sysflag       bool          //内置方法
	name          string        //定时任务名
	job           JobFunc       //注册循环任务方法
	jobp          interface{}   //任务参数
	singleton     *gtype.Bool   //任务是否单例运行
	status        *gtype.Int    //任务状态(0: ready; 1:running; 2: stopped; -1:closed),层级entry共享状态
	times         *gtype.Int    //还需运行次数
	createTime    time.Time     //创建任务时间
	runTimes      *gtype.Int         //任务运行次数
	interval      time.Duration //任务间隔时间
	create        int64         //注册时的时间轮tick
	interval_num  int64         //设置的运行间隔(时间轮刻度数量)
	createMs      int64         //创建时间(毫秒)
	intervalMs    int64         //间隔时间(毫秒)
	rawIntervalMs int64         //原始间隔
}

type JobFunc = func(string, time.Time, time.Duration, interface{}) //作业功能

func (entry *Entry) String() string {
	out_str := fmt.Sprintf("[%s] 状态=%s， 间隔=%d, 间隔=%ds", entry.name, entry.GetStatus(), entry.interval_num, entry.intervalMs/1000)
	return out_str
}

// 获取任务状态
func (entry *Entry) GetStatus() string {
	if entry.status.Val() == STATUS_READY {
		return "准备"
	}
	if entry.status.Val() == STATUS_RUNNING {
		return "运行"
	}
	if entry.status.Val() == STATUS_STOPPED {
		return "停止"
	}
	if entry.status.Val() == STATUS_CLOSED {
		return "关闭"
	}
	return "未知"
}

// 获取任务状态
func (entry *Entry) Status() int {
	return entry.status.Val()
}

// SetStatus custom sets the status for the job.
// 设置任务状态
func (entry *Entry) SetStatus(status int) int {
	return entry.status.Set(status)
}

// Start starts the job.
func (entry *Entry) Start() {
	entry.status.Set(STATUS_READY)
}

// Stop stops the job.
// 关闭当前任务
func (entry *Entry) Stop() {
	entry.status.Set(STATUS_STOPPED)
}

// Close closes the job, and then it will be removed from the timer.
func (entry *Entry) Close() {
	entry.status.Set(STATUS_CLOSED)
}

// IsSingleton checks and returns whether the job in singleton mode.
func (entry *Entry) IsSingleton() bool {
	return entry.singleton.Val()
}

// SetSingleton sets the job singleton mode.
func (entry *Entry) SetSingleton(enabled bool) {
	entry.singleton.Set(enabled)
}

// SetTimes sets the limit running times for the job.
func (entry *Entry) SetTimes(times int) {
	entry.times.Set(times)
}

// Run runs the job.
func (entry *Entry) Run() {
	entry.job(entry.name, entry.createTime, entry.interval, entry.jobp)
}

//检测是否需要运行任务，传入：当前刻度，当前时间戳
func (entry *Entry) check(nowTicks int64, nowMs int64) (runnable, addable bool) {
	switch entry.status.Val() {
	case STATUS_STOPPED:
		return false, true
	case STATUS_CLOSED:
		return false, false
	}

	//时间轮刻度判断，是否符合运行刻度条件，刻度判断的误差会比较大
	//提高精度部分是一个原因，但是会随着时间轮的继续转动，精度会越来越精确
	if diff := nowTicks - entry.create; diff > 0 && diff%entry.interval_num == 0 {

		//分层转换处理
		if entry.wheel.level > 0 {
			diffMs := nowMs - entry.createMs
			switch {

			//表示新增(当添加任务后在下一时间轮刻度马上触发)
			case diffMs < entry.wheel.timer.intervalMs:
				entry.wheel.slots[(nowTicks+entry.interval_num)%entry.wheel.slot_size].PushBack(entry)
				return false, false

			//正常任务
			case diffMs >= entry.wheel.timer.intervalMs:
				// 任务是否有必要进行分层转换
				// 经过的时间(执行check的瞬间)在任务一个间隔期之内，并且剩余执行一个周期的时间大于当前时间轮的最小时间轮刻度
				// 这里说明任务的间隔比较大，需要提高精度
				if leftMs := entry.intervalMs - diffMs; leftMs > entry.wheel.timer.intervalMs {
					// 往底层添加，通过毫秒计算并重新添加任务到对应的时间轮上，减小运行误差
					// 当前ticks是不会执行的
					entry.wheel.timer.doAddEntryByParent(leftMs, entry)
					return false, false
				}
			}
		}

		// 是否单例
		if entry.IsSingleton() {
			// Note that it is atomic operation to ensure concurrent safety.
			// 注意原子操作结果判断
			if entry.status.Set(STATUS_RUNNING) == STATUS_RUNNING {
				return false, true
			}
		}

		// 次数限制
		times := entry.times.Add(-1)
		if times <= 0 {
			// 注意原子操作结果判断
			if entry.status.Set(STATUS_CLOSED) == STATUS_CLOSED || times < 0 {
				return false, false
			}
		}

		// 是否不限制运行次数
		if times < 2000000000 && times > 1000000000 {
			entry.times.Set(DEFAULT_TIMES)
		}
		return true, true //todo:runnable这个字段到底是是代表什么？这里又是true了？
	}
	return false, true
}
