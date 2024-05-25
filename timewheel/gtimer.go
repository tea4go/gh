package timewheel

import (
	"math"
)

/*注意事项：

任何的定时任务都是有误差的，在时间轮刻度比较大，或者并发量大，负载较高的系统中尤其明显，具体请参考：https://github.com/golang/go/issues/14410
间隔不会考虑任务的执行时间。例如，如果一项工作需要3分钟才能执行完成，并且计划每隔5分钟运行一次，那么每次任务之间只有2分钟的空闲时间。
需要注意的是单例模式运行的定时任务，任务的执行时间会影响该任务下一次执行的开始时间。例如：一个每秒执行的任务，运行耗时为1秒，那么在第1秒开始运行后，下一次任务将会在第3秒开始执行。

简要说明：

New 方法用于创建自定义的任务定时器对象:
 - slot 参数用于指定每个时间轮的槽数；
 - interval 参数用于指定定时器的最小tick时间间隔；
 - level 为非必需参数，用于自定义分层时间轮的层数，默认为6；
Add 方法用于添加定时任务，其中：
 - interval 参数用于指定方法的执行的时间间隔；
 - job 参数为需要执行的任务方法(方法地址)；
AddEntry 方法添加定时任务，支持更多参数的控制；
AddSingleton 方法用于添加单例定时任务，即同时只能有一个该任务正在运行；
AddOnce 方法用于添加只运行一次的定时任务，当运行一次数后该定时任务自动销毁；
AddTimes 方法用于添加运行指定次数的定时任务，当运行times次数后该定时任务自动销毁；
Search 方法用于根据名称进行定时任务搜索(返回定时任务*Entry对象指针)；
Start 方法用于启动定时任务(Add后自动启动定时任务)；
Stop 方法用于停止定时任务；
Close 方法用于关闭定时器；
*/
const (
	STATUS_READY   = 0
	STATUS_RUNNING = 1
	STATUS_STOPPED = 2
	STATUS_CLOSED  = -1
	PANIC_EXIT     = "exit"
	DEFAULT_TIMES  = math.MaxInt32
)

// var (
// 	// 默认的wheel管理对象; slots = 时间轮槽数; 间隔时间 = 50ms; 分层大小 = 6
// 	defaultTimer = nil // NewTimer(10, 50*time.Millisecond, 6)
// )

// //增加循环任务
// func AddTimer(interval time.Duration, job JobFunc, jobp interface{}) *Entry {
// 	return defaultTimer.AddTimer(interval, job, jobp)
// }
