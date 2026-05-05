use chrono::Local;
use clap::Parser;
use rand::Rng;
use std::io::Write;
use std::net::{TcpStream, ToSocketAddrs};
use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::{Arc, Mutex};
use std::thread;
use std::time::Duration;

/// Log4go TCP test client (Rust version) — 与 testserver.go 配合使用
#[derive(Parser, Debug)]
#[command(version, about = "与 testserver.go 配合的日志发送客户端", long_about = None)]
struct Args {
    /// 日志级别 (0-7, 数字越大越详细)
    #[arg(short = 'l', default_value = "5")]
    #[allow(dead_code)]
    log_level: u8,

    /// 日志服务器地址
    #[arg(long, default_value = "127.0.0.1")]
    host: String,

    /// 日志服务器端口
    #[arg(long, default_value_t = 9514)]
    port: u16,

    /// 并发线程数
    #[arg(short = 't', default_value_t = 1)]
    threads: u32,

    /// 每个线程的批次数量
    #[arg(short = 'b', default_value_t = 1)]
    batch: u32,

    /// 是否显示调试信息
    #[arg(long)]
    #[allow(dead_code)]
    debug: bool,
}

// 日志级别前缀 (与 log4go levelPrefix 一致)
const LEVEL_PREFIX: &[&str] = &["[M]", "[A]", "[C]", "[E]", "[W]", "[N]", "[I]", "[D]"];

// 日志级别名称
const LEVEL_NAMES: &[&str] = &[
    "Emergency", "Alert", "Critical", "Error",
    "Warning", "Notice", "Info", "Debug",
];

// ANSI 颜色定义 — 与 Go log4go console.go 的 colors 数组完全一致
const LEVEL_COLORS: &[&str] = &[
    "1;37;41", // Emergency  高亮白+红底
    "1;37;45", // Alert      高亮白+紫红底
    "1;33;46", // Critical   高亮黄+青底
    "1;31",    // Error      高亮红
    "1;33",    // Warning    高亮黄
    "1;32",    // Notice     高亮绿
    "1;34",    // Info       高亮蓝
    "1;37",    // Debug      高亮白
];

/// 用 ANSI 颜色包裹消息 (与 Go newBrush 逻辑一致)
fn colorize(msg: &str, level: usize) -> String {
    // Go 的 newBrush: 去掉末尾 \n -> 包裹颜色 -> 加 \n
    let stripped = msg.strip_suffix('\n').unwrap_or(msg);
    format!("\x1b[{}m{}\x1b[0m\n", LEVEL_COLORS[level], stripped)
}

/// 构建与 Go log4go connWriter.WriteMsg 格式一致的带颜色日志消息
/// 格式: `HH.MM.SS (文件名:行号) [LEVEL]> 消息\n` (包裹 ANSI 颜色)
fn format_log_msg(level: usize, b: u32, i: u32) -> String {
    let now = Local::now();
    let time_str = now.format("%H.%M.%S").to_string();
    let prefix = LEVEL_PREFIX[level];
    let name = LEVEL_NAMES[level];
    let raw = format!(
        "{} (testclient.rs:1) {}> TestLog {:04}-{:04}({})\n",
        time_str, prefix, b, i, name
    );
    colorize(&raw, level)
}

fn main() {
    let args = Args::parse();

    // 设置 Ctrl+C 信号处理
    let running = Arc::new(AtomicBool::new(true));
    let r = running.clone();
    ctrlc::set_handler(move || {
        eprintln!("\n-= 退出 =-");
        r.store(false, Ordering::SeqCst);
    })
    .expect("设置信号处理失败");

    // 连接服务器
    let addr = format!("{}:{}", args.host, args.port);
    let addr2 = addr
        .to_socket_addrs()
        .expect("地址解析失败")
        .next()
        .expect("无效地址");

    let stream = TcpStream::connect_timeout(&addr2, Duration::from_secs(5))
        .expect("连接日志服务器失败");
    stream
        .set_write_timeout(Some(Duration::from_secs(3)))
        .ok();
    // TCP_NODELAY 确保小数据包立即发送
    stream.set_nodelay(true).ok();
    let stream = Arc::new(Mutex::new(stream));

    // 发送日志名称 (与 Go connWriter.connect 一致)
    {
        let mut s = stream.lock().unwrap();
        s.write_all(b"{LogName}testlog{LogName}\n")
            .expect("发送日志名称失败");
        s.flush().ok();
    }

    println!("Start Log4go Client ...... ({}:{})", args.host, args.port);
    println!(
        "= 开启 {} 个线程，每个线程执行 {} 批次日志。",
        args.threads, args.batch
    );
    println!("...... 请按 Ctrl+C 结束 ......");
    println!("================================================================");

    // 心跳 goroutine (每5秒发送 {HeartBeat})
    let stream_hb = stream.clone();
    let running_hb = running.clone();
    thread::spawn(move || {
        while running_hb.load(Ordering::SeqCst) {
            thread::sleep(Duration::from_secs(5));
            let mut s = match stream_hb.lock() {
                Ok(s) => s,
                Err(_) => break,
            };
            if s.write_all(b"{HeartBeat}\n").is_err() {
                break;
            }
            s.flush().ok();
        }
    });

    let mut handles = vec![];

    for t in 1..=args.threads {
        let stream_t = stream.clone();
        let running_t = running.clone();
        let batch = args.batch;

        // 线程间随机休眠 0~1000 毫秒 (与 Go 端一致)
        {
            let mut rng = rand::thread_rng();
            let sleep_ms = rng.gen_range(0..1000);
            thread::sleep(Duration::from_millis(sleep_ms));
        }

        let handle = thread::spawn(move || {
            let mut rng = rand::thread_rng();

            for i in 1..=batch {
                if !running_t.load(Ordering::SeqCst) {
                    break;
                }

                println!("= 第{}线程，第{}批测试", t, i);

                // 逐级别发送日志 (0=Emergency .. 7=Debug)
                for level in 0..8 {
                    let msg = format_log_msg(level, t, i);
                    let mut s = stream_t.lock().unwrap();
                    if s.write_all(msg.as_bytes()).is_err() {
                        return;
                    }
                    s.flush().ok();
                }

                // 发送分隔线 (Go 端通过 log.Notice 发送，带颜色)
                {
                    let now = Local::now();
                    let time_str = now.format("%H.%M.%S").to_string();
                    let raw_sep = format!(
                        "{} (testclient.rs:1) [N]> --------------------------\n",
                        time_str
                    );
                    let sep = colorize(&raw_sep, 5); // 5 = Notice, 对应绿色
                    let mut s = stream_t.lock().unwrap();
                    s.write_all(sep.as_bytes()).ok();
                    s.flush().ok();
                }

                // 随机休眠 1000~2000 毫秒 (与 Go 端一致)
                let sleep_ms = rng.gen_range(1000..2000);
                thread::sleep(Duration::from_millis(sleep_ms));
            }
        });

        handles.push(handle);
    }

    // 等待所有线程完成
    for h in handles {
        h.join().unwrap();
    }

    println!("所有日志发送完成");
}
