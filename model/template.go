package model

var (
	Machine = "machine"
	Collect = "collect"
	Alarm   = "alarm"

	Templates []string = []string{
		Machine, Alarm, Collect, "doc", "user", "group", "route", "ns",
		// "init", "deploy", "acl", "owner", "route",
	}

	PkProperty = map[string]string{
		Machine: "hostname",
		Collect: "name",
		Alarm:   "name",
		"doc":   "name",
	}

	TemplatePrefix string = "_template_"
)

var (
	collectTemplate ResourceList = ResourceList{Resource{
		"name":             "app.service.coredump",
		"interval":         "60",
		"measurement_type": "COREDUMP",
		"comment":          "检测 /home/coresave 下生成的 core",
	}, Resource{
		"comment":          "",
		"name":             "cpu.idle",
		"interval":         "10",
		"measurement_type": "CPU",
	}, Resource{
		"comment":          "机器单核采集",
		"name":             "cpu.idle.core",
		"interval":         "10",
		"aggregate":        "avg",
		"measurement_type": "CPU",
	}, Resource{
		"comment":          "最近一分钟服务器负载",
		"name":             "cpu.loadavg.1",
		"interval":         "10",
		"measurement_type": "CPU",
	}, Resource{
		"comment":          "最近十五分钟服务器负载",
		"name":             "cpu.loadavg.15",
		"interval":         "10",
		"measurement_type": "CPU",
	}, Resource{
		"comment":          "最近五分钟服务器负载",
		"name":             "cpu.loadavg.5",
		"interval":         "10",
		"measurement_type": "CPU",
	}, Resource{
		"comment":          "文件系统 inode 使用率",
		"name":             "fs.inodes.used.percent",
		"interval":         "120",
		"measurement_type": "FS",
	}, Resource{
		"comment":          "文件系统空间使用率",
		"name":             "fs.space.used.percent",
		"interval":         "120",
		"measurement_type": "FS",
	}, Resource{
		"comment":          "检测文件系统故障. 0 表示文件系统读写故障, 1表示文件系统正常",
		"name":             "fs.files.rw",
		"interval":         "300",
		"measurement_type": "FS",
	}, Resource{
		"comment":          "整个系统被分配的file handles",
		"name":             "kernel.files.allocated",
		"interval":         "300",
		"measurement_type": "KERNEL",
	}, Resource{
		"comment":          "整个系统剩余可以分配的 file handles",
		"name":             "kernel.files.left",
		"interval":         "300",
		"measurement_type": "KERNEL",
	}, Resource{
		"comment":          "整个系统所有进程能够打开的最多文件数",
		"name":             "kernel.files.max",
		"interval":         "300",
		"measurement_type": "KERNEL",
	}, Resource{
		"comment":          "整个系统的file handles 的使用率",
		"name":             "kernel.files.used.percent",
		"interval":         "300",
		"measurement_type": "KERNEL",
	}, Resource{
		"comment":          "CPU等待 IO 操作时间",
		"name":             "disk.io.await",
		"interval":         "10",
		"measurement_type": "DISK",
	}, Resource{
		"comment":          "io使用率",
		"name":             "disk.io.util",
		"interval":         "10",
		"measurement_type": "DISK",
	}, Resource{
		"comment":          "磁盘写IO",
		"interval":         "10",
		"measurement_type": "DISK",
		"name":             "disk.io.write_requests",
	}, Resource{
		"comment":          "磁盘读IO",
		"interval":         "10",
		"measurement_type": "DISK",
		"name":             "disk.io.read_requests",
	}, Resource{
		"collect_type":     "FLOW",
		"degree":           "0",
		"file_path":        "/var/log/kernel",
		"func":             "cnt",
		"interval":         "10",
		"measurement_type": "LOG",
		"name":             "kernel.log.OOM",
		"pattern":          "Out of memory",
		"tags":             "service",
		"tag_service":      "kill process \\d+ \\((\\S+)\\)",
		"time_format":      "yyyy-mm-ddTHH:MM:SS",
	}, Resource{
		"comment":          "内核错误日志(I/O error|EXT3-fs error|ERROR on|Medium Error|error recovery|disk error|Illegal block|Out of Memory|dead device|readonly)条数. ",
		"name":             "kernel_log_monitor",
		"interval":         "300",
		"measurement_type": "KERNEL",
	}, Resource{
		"comment":          "服务器心跳",
		"name":             "agent.alive",
		"interval":         "10",
		"measurement_type": "HEALTH",
	}, Resource{
		"comment":          "内存缓存量",
		"name":             "mem.buffers",
		"interval":         "10",
		"measurement_type": "MEM",
	}, Resource{
		"comment":          "内存空闲量",
		"name":             "mem.free",
		"interval":         "10",
		"measurement_type": "MEM",
	}, Resource{
		"comment":          "机器物理内存总量",
		"name":             "mem.total",
		"interval":         "10",
		"measurement_type": "MEM",
	}, Resource{
		"comment":          "机器内存使用率",
		"name":             "mem.used",
		"interval":         "10",
		"measurement_type": "MEM",
	}, Resource{
		"comment":          "内存使用率",
		"name":             "mem.used.percent",
		"interval":         "10",
		"measurement_type": "MEM",
	}, Resource{
		"comment":          "网卡入口流量",
		"name":             "net.in",
		"interval":         "10",
		"measurement_type": "NET",
	}, Resource{
		"comment":          "网络入口丢包数",
		"name":             "net.in.dropped",
		"interval":         "10",
		"measurement_type": "NET",
	}, Resource{
		"comment":          "网卡出口流量",
		"name":             "net.out",
		"interval":         "10",
		"aggregate":        "sum",
		"measurement_type": "NET",
	}, Resource{
		"comment":          "网络出口丢包数",
		"name":             "net.out.dropped",
		"interval":         "10",
		"measurement_type": "NET",
	}, Resource{
		"comment":          "正在使用（正在侦听）的TCP socket 数量",
		"name":             "net.tcp.inuse",
		"interval":         "10",
		"measurement_type": "NET",
	}, Resource{
		"comment":          "已使用的所有协议socket总量",
		"name":             "net.sockets.used",
		"interval":         "10",
		"measurement_type": "NET",
	}, Resource{
		"comment":          "机器timewait连接数",
		"name":             "net.tcp.timewait",
		"interval":         "15",
		"measurement_type": "NET",
	}, Resource{
		"comment":          "机器和 ntp server 时间差(ms)",
		"name":             "time.offset",
		"interval":         "300",
		"measurement_type": "TIME",
	}, /* Resource{
		"connect_timeout":  "3",
		"interval":         "10",
		"measurement_type": "PORT",
		"name":             "port.sshd.22",
		"port":             "22",
	}, Resource{
		"git":              "git@git.loda.com:plugins/process.git",
		"interval":         "10",
		"measurement_type": "PLUGIN",
		"name":             "loda-plugin",
		"parameters":       "-x loda",
	}, Resource{
		"bin_path":         "/usr/local/registry/bin/registry",
		"comment":          "registry service",
		"interval":         "10",
		"measurement_type": "PROC",
		"name":             "registry",
	},*/
	}

	alarmTemplate ResourceList = ResourceList{
		Resource{"name": "todo", "default": "true", "trigger": "threshold", "enable": "true", "every": "1m", "period": "1m", "measurement": "cpu.idle", "function": "mean", "expression": "<", "value": "10", "groupby": "todo", "groups": "todo", "level": "2", "message": "todo", "md5": "todo", "rp": "loda", "shift": "5", "alert": "todo", "where": ""},
	}
	RootTemplate map[string]ResourceList
)

func init() {
	RootTemplate = make(map[string]ResourceList)
	for _, resType := range Templates {
		RootTemplate[TemplatePrefix+resType] = nil
	}
	RootTemplate[TemplatePrefix+Collect] = collectTemplate
	RootTemplate[TemplatePrefix+Alarm] = alarmTemplate
}
