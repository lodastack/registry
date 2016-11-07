***方便记录和调试***

# HTTP API Doc


### 0 管理接口
---

#### 0.1 加入成员

向集群中加入一个成员。

```
curl -x POST -d '{"addr":"127.0.0.2:9991"}' "http://127.0.0.1:9991/join"
```
#### 0.2 删除成员

在集群中删除一个成员。

```
curl -x DELETE -d '{"addr":"127.0.0.2:9991"}' "http://127.0.0.1:9991/join"
```

#### 0.3 备份数据（只能在leader上操作）
备份整个数据库。

```
curl "http://127.0.0.1:9991/backup" > /data/backup.db
```

#### 0.4 恢复数据（只能在leader上操作）

从备份的文件中恢复操作会恢复整个集群中每个节点的数据。

```
curl "http://127.0.0.1:9991/restore?file=/data/backup.db"
```


### 1 节点接口
---
程序会初始化一个初始节点，ID为`0`，name为`loda`。所有节点都基于这个初始节点。

每个节点都会对应一个ns。比如：`test.service.product.loda`

#### 1.1 新建节点
只能在非叶子节点下新建节点。
新建节点会同时创建节点ID，并创建节点对用存储bucket。
需要提供3个参数：
- parent: 父节点的节点ID
- type: 节点类型，0为叶子节点，1为非叶子节点
- name：节点名称，用于组成节点ns

成功返回：返回新建节点ID

例子  （初始节点ID为`0`）

    # 新建非叶子节点
    curl -X POST "http://127.0.0.1:9991/ns?parent=0&type=1&name=zzznl"
    #返回
    98a6586c-4f4f-475c-9c5c-8d4e31d14cd2
    
	#新建叶子节点
	curl -X POST "http://127.0.0.1:9991/ns?parent=0&type=0&name=zzzl"
    
    #在zzznl.loda下新建叶子节点
	curl -X POST "http://127.0.0.1:9991/ns?parent=98a6586c-4f4f-475c-9c5c-8d4e31d14cd2&type=0&name=zzzl2"

#### 1.2 查询节点
查询全部节点

    curl "http://127.0.0.1:9991/ns"
    #返回
    {"Children":[{"Children":[{"Children":[],"ID":"c357e90e-641b-4576-8df6-3325fdffe6b8","Name":"zzzl2","Type":0,"MachineReg":"-"}],"ID":"f06eb976-34a9-4680-b7a5-40c5eb3ebdac","Name":"zzznl","Type":1,"MachineReg":"-"},{"Children":[],"ID":"7fa82cda-098a-4baf-8f96-3561d52c26f4","Name":"zzzl","Type":0,"MachineReg":"-"}],"ID":"0","Name":"loda","Type":1,"MachineReg":"*"}

根据节点ID查询节点

    curl "http://127.0.0.1:9991/ns?nodeid=c357e90e-641b-4576-8df6-3325fdffe6b8"
    # 返回
    {"Children":[],"ID":"c357e90e-641b-4576-8df6-3325fdffe6b8","Name":"zzzl2","Type":0,"MachineReg":"-"}

### 2 资源接口
---

#### 2.1 设置资源

只能在叶子节点下设置资源，目前只能设置全量资源，不能追加资源。
提供参数：
- query参数 ns：资源所在的叶子节点ns
- query参数resouce：资源类型
- body参数：资源内容格式为maplist，系统会给每个资源***添加资源ID***

例子：

    curl -d '[{"host":"127.0.0.2"},{"host":"127.0.0.1"}]' "http://127.0.0.1:9991/resource?ns=zzzl.loda&resource=machine"
    curl -d '[{"host":"127.0.0.2"},{"host":"127.0.0.3"}]' "http://127.0.0.1:9991/resource?ns=zzzl2.zzznl.loda&resource=machine"

#### 2.2 查询资源

提供参数：
- query参数 ns：资源所在的叶子节点ns
- query参数resouce：资源类型

例子：

     curl "http://127.0.0.1:9991/resource?ns=zzzl2.zzznl.loda&resource=machine"
     curl "http://127.0.0.1:9991/resource?ns=zzzl.loda&resource=machine"

#### 2.3 搜索资源

提供参数：
- query参数 ns：资源所在的叶子节点ns
- query参数resouce：资源类型
- query参数type: 搜索类型，fuzzy为模糊搜索。功能上同strings.contain()进行搜索
- query参数k/v： 搜索的属性k-v

例子：

     curl "http://127.0.0.1:9991/search?ns=loda&resource=machine&k=host&v=127.0.0.2&type=exact"|jq
     curl "http://127.0.0.1:9991/search?ns=loda&resource=machine&k=host&v=127.0.0.&type=fuzzy"|jq