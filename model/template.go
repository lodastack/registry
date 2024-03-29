package model

var (
	Machine   = "machine"
	Collect   = "collect"
	Alarm     = "alarm"
	Deploy    = "deploy"
	Group     = "group"
	Namespace = "ns"
	User      = "user"

	Templates []string = []string{
		Machine, Alarm, Collect, Deploy, Group, User, Namespace,
	}

	PkProperty = map[string]string{
		Machine: "hostname",
		Collect: "name",
		Alarm:   "name",
		Deploy:  "name",
	}

	TemplatePrefix     string = "_template_"
	TemplateCollectNum int    = len(collectTemplate)
	TemplateAlarmNum   int    = len(alarmTemplate)
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
		"comment":          "文件系统空间使用量",
		"name":             "fs.space.used",
		"interval":         "120",
		"measurement_type": "FS",
	}, Resource{
		"comment":          "文件系统空间剩余量",
		"name":             "fs.space.free",
		"interval":         "120",
		"measurement_type": "FS",
	}, Resource{
		"comment":          "文件系统空间总量",
		"name":             "fs.space.total",
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
		"comment":          "内存Cache",
		"name":             "mem.cache",
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
	}, Resource{
		"comment":          "网卡速率",
		"interval":         "10",
		"measurement_type": "NET",
		"name":             "net.speed",
	}, Resource{
		"comment":          "僵尸进程数",
		"interval":         "10",
		"measurement_type": "PS",
		"name":             "ps.zombies.num",
	}, Resource{
		"comment":          "交换分区使用率",
		"interval":         "10",
		"measurement_type": "MEM",
		"name":             "mem.swap.used.percent",
	},

	/* Resource{
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
		"bin_path":         "/usr/local/loda/bin/loda-server",
		"comment":          "registry service",
		"interval":         "10",
		"measurement_type": "PROC",
		"name":             "loda-server",
	},*/
	}

	alarmTemplate ResourceList = ResourceList{
		Resource{
			"alert":        "mail,sms,wechat",
			"blockstep":    "5",
			"default":      "true",
			"enable":       "true",
			"endtime":      "",
			"every":        "1m",
			"expression":   ">",
			"func":         "mean",
			"groupby":      "*",
			"groups":       "groups",
			"level":        "2",
			"maxblocktime": "60",
			"md5":          "md5",
			"measurement":  "cpu.loadavg.5",
			"name":         "负载过高",
			"period":       "1m",
			"rp":           "loda",
			"starttime":    "",
			"trigger":      "threshold",
			"value":        "40",
			"where":        "",
		},
		Resource{
			"alert":        "mail,sms,wechat",
			"blockstep":    "5",
			"default":      "true",
			"enable":       "true",
			"endtime":      "",
			"every":        "1m",
			"expression":   "<",
			"func":         "mean",
			"groupby":      "*",
			"groups":       "groups",
			"level":        "2",
			"maxblocktime": "60",
			"md5":          "md5",
			"measurement":  "cpu.idle",
			"name":         "cpu空闲低",
			"period":       "1m",
			"rp":           "loda",
			"starttime":    "",
			"trigger":      "threshold",
			"value":        "30",
			"where":        "",
		},
		Resource{
			"alert":        "mail,sms,wechat",
			"blockstep":    "5",
			"default":      "true",
			"enable":       "true",
			"endtime":      "0",
			"every":        "1m",
			"expression":   ">",
			"func":         "mean",
			"groupby":      "*",
			"groups":       "groups",
			"level":        "2",
			"maxblocktime": "60",
			"md5":          "md5",
			"measurement":  "mem.used.percent",
			"name":         "内存使用过高",
			"period":       "1m",
			"rp":           "loda",
			"starttime":    "0",
			"trigger":      "threshold",
			"value":        "95",
			"where":        "",
		},
		Resource{
			"alert":        "mail,sms,wechat",
			"blockstep":    "5",
			"default":      "true",
			"enable":       "true",
			"endtime":      "",
			"every":        "1m",
			"expression":   ">",
			"func":         "mean",
			"groupby":      "*",
			"groups":       "groups",
			"level":        "2",
			"maxblocktime": "60",
			"md5":          "md5",
			"measurement":  "disk.io.util",
			"name":         "磁盘IO过高",
			"period":       "1m",
			"rp":           "loda",
			"starttime":    "",
			"trigger":      "threshold",
			"value":        "80",
			"where":        "",
		},
		Resource{
			"alert":       "mail,sms,wechat",
			"default":     "true",
			"enable":      "true",
			"every":       "1m",
			"expression":  ">",
			"func":        "mean",
			"groupby":     "*",
			"groups":      "groups",
			"level":       "2",
			"md5":         "md5",
			"measurement": "fs.inodes.used.percent",
			"name":        "文件系统innode过高",
			"period":      "1m",
			"rp":          "loda",
			"trigger":     "threshold",
			"value":       "80",
			"where":       "",
		},
		Resource{
			"alert":       "mail,sms,wechat",
			"default":     "true",
			"enable":      "true",
			"every":       "1m",
			"expression":  ">",
			"func":        "mean",
			"groupby":     "*",
			"groups":      "groups",
			"level":       "2",
			"md5":         "md5",
			"measurement": "fs.space.used.percent",
			"name":        "磁盘空间过低",
			"period":      "1m",
			"rp":          "loda",
			"trigger":     "threshold",
			"value":       "90",
			"where":       "",
		},
		Resource{
			"alert":       "mail,sms,wechat",
			"default":     "true",
			"enable":      "true",
			"every":       "2m",
			"expression":  ">",
			"func":        "mean",
			"groupby":     "*",
			"groups":      "groups",
			"level":       "1",
			"md5":         "md5",
			"measurement": "RUN.ping.loss",
			"name":        "ping监控",
			"period":      "2m",
			"rp":          "loda",
			"trigger":     "threshold",
			"value":       "60",
			"where":       "",
		},
		Resource{
			"alert":        "mail,sms,wecaht",
			"blockstep":    "10",
			"default":      "true",
			"enable":       "true",
			"endtime":      "0",
			"every":        "1m",
			"expression":   "==",
			"func":         "mean",
			"groupby":      "*",
			"groups":       "groups",
			"level":        "2",
			"maxblocktime": "60",
			"md5":          "md5",
			"measurement":  "net.speed",
			"name":         "网卡被识别成百兆",
			"period":       "1m",
			"rp":           "loda",
			"starttime":    "0",
			"trigger":      "threshold",
			"value":        "100",
			"where":        "",
		},
		Resource{
			"alert":        "mail,sms,wechat",
			"blockstep":    "10",
			"default":      "true",
			"enable":       "true",
			"endtime":      "0",
			"every":        "1m",
			"expression":   ">",
			"func":         "mean",
			"groupby":      "*",
			"groups":       "groups",
			"level":        "2",
			"maxblocktime": "60",
			"md5":          "md5",
			"measurement":  "ps.zombies.num",
			"name":         "僵尸进程数量大于15",
			"period":       "1m",
			"rp":           "loda",
			"starttime":    "0",
			"trigger":      "threshold",
			"value":        "15",
			"where":        "",
		},
		Resource{
			"alert":        "mail,sms,wechat",
			"blockstep":    "5",
			"default":      "true",
			"enable":       "true",
			"endtime":      "",
			"every":        "5m",
			"expression":   "==",
			"func":         "mean",
			"groupby":      "*",
			"groups":       "groups",
			"level":        "1",
			"maxblocktime": "60",
			"md5":          "md5",
			"measurement":  "fs.space.used.percent",
			"name":         "磁盘无剩余空间",
			"period":       "5m",
			"rp":           "loda",
			"starttime":    "",
			"trigger":      "threshold",
			"value":        "100",
			"where":        "",
		},
		Resource{
			"alert":        "mail,sms,wechat",
			"blockstep":    "5",
			"default":      "true",
			"enable":       "true",
			"endtime":      "",
			"every":        "1m",
			"expression":   "==",
			"func":         "mean",
			"groupby":      "*",
			"groups":       "groups",
			"level":        "1",
			"maxblocktime": "60",
			"md5":          "md5",
			"measurement":  "agent.alive",
			"name":         "无监控数据上报",
			"period":       "10m",
			"rp":           "loda",
			"starttime":    "",
			"trigger":      "deadman",
			"value":        "0",
			"where":        "",
		},
		Resource{
			"alert":        "mail,sms,wechat",
			"blockstep":    "5",
			"default":      "true",
			"enable":       "true",
			"endtime":      "",
			"every":        "5m",
			"expression":   "<",
			"func":         "mean",
			"groupby":      "*",
			"groups":       "groups",
			"level":        "1",
			"maxblocktime": "60",
			"md5":          "md5",
			"measurement":  "fs.files.rw",
			"name":         "文件系统损坏",
			"period":       "5m",
			"rp":           "loda",
			"starttime":    "",
			"trigger":      "threshold",
			"value":        "1",
			"where":        "",
		},
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
