***方便记录和调试***

# HTTP API Doc


### 0 管理接口
---

#### 0.1 加入成员

向集群中加入一个成员。

```
curl -x POST -d '{"addr":"127.0.0.2:9991"}' "http://127.0.0.1:9991/api/v1/join"
```
#### 0.2 删除成员

在集群中删除一个成员。

```
curl -x DELETE -d '{"addr":"127.0.0.2:9991"}' "http://127.0.0.1:9991/api/v1/join"
```

#### 0.3 备份数据（只能在leader上操作）
备份整个数据库。

```
curl "http://127.0.0.1:9991/api/v1/backup" > /data/backup.db
```

#### 0.4 恢复数据（只能在leader上操作）

从备份的文件中恢复操作会恢复整个集群中每个节点的数据。

```
curl "http://127.0.0.1:9991/api/v1/restore?file=/data/backup.db"
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
- matchreg: 机器政策匹配规则，如果新机器匹配到规则，则注册到该节点下。默认不进行匹配

成功返回：返回新建节点ID

例子  （初始节点ID为`0`）

    # 新建非叶子节点
    curl -X POST "http://127.0.0.1:9991/api/v1/ns?parent=0&type=1&name=product1"
    #返回
    98a6586c-4f4f-475c-9c5c-8d4e31d14cd2
    
    #新建叶子节点
    curl -X POST "http://127.0.0.1:9991/api/v1/ns?parent=0&type=0&name=server1"
    curl -X POST "http://127.0.0.1:9991/api/v1/ns?parent=0&type=0&name=server2&machinereg=server2-machine"
    
    #在prodect1.loda下新建叶子节点
    curl -X POST "http://127.0.0.1:9991/api/v1/ns?parent=prodect1-loda-NodeID&type=0&name=server1"

#### 1.2 查询节点
查询全部节点

    curl "http://127.0.0.1:9991/api/v1/ns"
    #返回
    {
      "ID": "0",
      "Name": "loda",
      "Type": 1,
      "MachineReg": "^$",
      "Children": [
        {
          "Children": [],
          "ID": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxxx",
          "Name": "pool",
          "Type": 0,
          "MachineReg": "^$",
        },
        {
         "ID": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxxx",
          "Name": "product1",
          "Type": 1,
          "MachineReg": "^$",
          "Children": [
            {
              "Children": [],
              "ID": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxxx",
              "Name": "server1",
              "Type": 0,
              "MachineReg": "^$"
            }
          ]
        },
        {
          "Children": [],
          "ID": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxxx",
          "Name": "server1",
          "Type": 0,
          "MachineReg": "server"
        },
        {
          "Children": [],
          "ID": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxxx",
          "Name": "server2",
          "Type": 0,
          "MachineReg": "server2-machine"
        }
      ]
}

根据节点ID查询节点

    curl "http://127.0.0.1:9991/api/v1/ns?nodeid=Node-ID"
    # 返回
    {"Children":[],"ID":"xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxxx","Name":"server","Type":0,"MachineReg":"-"}

### 2 资源接口
---

#### 2.1 设置资源

只能在叶子节点下设置资源，目前只能设置全量资源，不能追加资源。
提供参数：
- query参数 ns：资源所在的叶子节点ns
- query参数resouce：资源类型
- body参数：资源内容格式为maplist，系统会给每个资源***添加资源ID***

例子：

    curl -d '[{"hostname":"127.0.0.2"},{"hostname":"127.0.0.1"}]' "http://127.0.0.1:9991/api/v1/resource?ns=pool.loda&resource=machine"
    curl -d '[{"hostname":"127.0.0.2"},{"hostname":"127.0.0.3"}]' "http://127.0.0.1:9991/api/v1/resource?ns=server1.product1.loda&resource=machine"

#### 2.2 查询资源

如果查询非叶子节点的某种资源(非模板)，则对该节点下所有叶子节点进行查询。

提供参数:
- query参数 nodeid: 资源所在节点的ID
- query参数 ns：资源所在的叶子节点ns(ID优先级较高)
- query参数resouce：资源类型

例子:

     curl "http://127.0.0.1:9991/api/v1/resource?ns=server1.product1.loda&resource=machine"
     curl "http://127.0.0.1:9991/api/v1/resource?ns=pool.loda&resource=machine"
     curl "http://127.0.0.1:9991/api/v1/resource?nodeid=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxxx&resource=machine"

#### 2.3 搜索资源

提供参数:
- query参数 ns：资源所在的叶子节点ns
- query参数resouce：资源类型
- query参数type: 搜索类型，fuzzy为模糊搜索。功能上同strings.contain()进行搜索
- query参数k/v： 搜索的属性k-v

例子:

    curl "http://127.0.0.1:9991/api/v1/resource/search?ns=loda&resource=machine&k=hostname&v=127.0.0.2&type=exact"|jq
    #返回
    {
      "pool.loda": [
        {
          "_id": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxxx"",
          "host": "127.0.0.2"
        }
      ],
      "server1.product1.loda": [
        {
          "_id": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxxx"",
          "host": "127.0.0.2"
        }
      ]
    }
    curl "http://127.0.0.1:9991/api/v1/resource/search?ns=loda&resource=machine&k=hostname&v=127.0.0.&type=fuzzy"|jq
    #返回
    {
      "pool.loda": [
        {
          "_id": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxxx"",
          "host": "127.0.0.2"
        },
        {
          "_id": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxxx"",
          "host": "127.0.0.1"
        }
      ],
      "server1.product1.loda": [
        {
          "_id": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxxx"",
          "host": "127.0.0.2"
        },
        {
          "_id": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxxx",
          "host": "127.0.0.3"
        }
      ]
    }

### 3 注册接口
---

根据hostname进行节点查找/注册:
- 如果机器已存在，则返回`map{ns:资源ID}`
- 如果机器不在任何节点下，则注册到hostname对应节点下，如果未匹配到任何节点，则注册到`pool.loda`中。

#### 3.1 注册接口

POST方法

提供参数:
- body参数 `map[string: string]`: 以hostname匹配machine资源。 例如`map["hostname":"xxx", "ips":"x.x.x.x,x.x.x.x", "status":"off"]` ***如果machine资源无hostname属性，则无法匹配***

结果返回：
- `map{ns: ResourceID}`

例子:

    curl -X POST -d '{"hostname":"pool2-machine","ips":"10.10.10.10,127.0.0.1","status":"off"}' "http://127.0.0.1:9991/api/v1/agent/ns"
    # 返回
    {"pool.loda":"f2f21847-652d-4a50-bfbb-d12df9b30b46"}
    curl -X POST -d '{"hostname":"server2-machine1","ips":"10.10.10.11,127.0.0.1","status":"off"}' "http://127.0.0.1:9991/api/v1/agent/ns"
    # 返回
    {"service1.product.loda":"b7705b32-11f4-4bef-acb1-fdbd47d2c7c0","service2.product.loda":"606df412-b043-4f12-8878-7e03089cb36e"}