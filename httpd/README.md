***方便记录和调试***

# HTTP API Doc


### 0 集群管理接口
---

#### 0.1 加入成员

agent启动时如果指定加入某个集群，agent会通过接口向指定的集群成员请求加入本机。正常返回表明加入成功，配置peer启动raft则自动加入集群。

```
curl -x POST -d '{"addr":"127.0.0.2:9991"}' "http://127.0.0.1:9991/api/v1/join"
```
#### 0.2 删除成员

在集群中删除一个成员。

```
curl -x DELETE -d '{"addr":"127.0.0.2:9991"}' "http://127.0.0.1:9991/api/v1/join"
```

#### 0.3 备份数据（只能在leader上操作）

备份整个数据库，返回数据库文件。需要将数据重定向到本地文件。

```
curl "http://127.0.0.1:9991/api/v1/backup" > /data/backup.db
```

#### 0.4 恢复数据（只能在leader上操作）

从备份的文件中恢复操作会恢复整个集群中每个节点的数据。

```
curl "http://127.0.0.1:9991/api/v1/restore?file=/data/backup.db"
```

#### 0.5 查看集群成员

查看集群成员机器状态（Leader/Follower）.每个结果已raft地址作为key，包含http接口及状态信息。

例子:

    curl "http://127.0.0.1:9991/api/v1/peer"
    # 返回
    # 8001/8002/8003为服务树http地址，9001/9002/9003为集群raft地址
    {
    "httpstatus": 200,
    "data": {
        "127.0.0.1:9001": {
            "api": "127.0.0.1:8001",
            "role": "Leader"
        },
        "127.0.0.1:9002": {
            "api": "127.0.0.1:8002",
            "role": "Follower"
        },
        "127.0.0.1:9003": {
            "api": "127.0.0.1:8003",
            "role": "Follower"
        }
    },
    "msg": ""
    }

### 1 节点接口
---

节点分为`叶子节点`和`非叶子节点`。服务树会初始化根节点`loda`，所有节点都基于这个初始节点。非叶子节点下可以建立节点。叶子节点下不可以再建立节点，但是可以存储`机器`、`监控`、`报警`、`发布`等资源。

把节点名称按照所属顺序连接起来，组成字符串`NS`。`NS`可以定位到一个具体节点，并用来定位节点中的资源。比如根节点下有非叶子节点`product`,`product`下有非叶子节点`service`，`service`下有叶子节点`test`，则`test`的`NS`为`test.service.product.loda`。`NS`可以用来定位`test`节点的`机器`、`监控`、`报警`、`发布`等资源。

#### 1.1 新建节点

只能在非叶子节点下新建节点，否则返回错误。
新建节点会同时生成节点ID，用来创建`bolt Bucket`以存储节点的各种资源。

`POST`方法，url: `/api/v1/ns`
参数：
- QUERY参数 ns: 父节点的节点Ns
- QUERY参数 type: 节点类型，`0`为叶子节点，`1`为非叶子节点
- QUERY参数 name：节点名称，用于组成节点ns
- QUERY参数 matchreg: 机器政策匹配规则，如果新机器匹配到规则，则注册到该节点下。默认不进行匹配

成功返回：JSON数据`{"httpstatus": 200/400/404/500, "data":JSON, "msg":"error msg"}`，httpstatus为返回的http状态码，data为Json, msg为string **如非特别说明，ns及资源接口返回格式相同**

例子

    # 新建非叶子节点 product0.loda，机器匹配规则为：^product0
    curl -X POST "http://127.0.0.1:9991/api/v1/ns?ns=loda&name=product0&type=1&machinereg=^product0"
    #返回
    {"httpstatus":200,"data":null,"msg":"success"}
    
    # 新建叶子节点 server0.loda,机器匹配规则为：server0.loda
    curl -X POST -d 'ns=loda&name=server0&type=0&matchreg=server0.loda' "http://127.0.0.1:9991/api/v1/ns"
    # 新建叶子节点 server1.loda,机器匹配规则为：server1.loda
    curl -X POST "http://127.0.0.1:9991/api/v1/ns?ns=loda&name=server1&type=0&matchreg=server1.loda"
    
    # 在prodect0.loda下新建叶子节点 server0.product0.loda
    curl -X POST -d 'ns=product0.loda&name=server0&type=0' "http://127.0.0.1:9991/api/v1/ns"
    # 在prodect0.loda下新建叶子节点 server1.product0.loda
    curl -X POST "http://127.0.0.1:9991/api/v1/ns?ns=product0.loda&name=server1&type=0&machinereg=server1.product0"

    # 错误：在叶子节点下新建节点
    curl -X POST -d 'ns=server1.loda&name=test&type=0' "http://127.0.0.1:9991/api/v1/ns"
    # 返回
    {"httpstatus": 500, "data": null, "msg": "can not create node under leaf node"}

#### 1.2 查询节点

查询节点机器所有子节点，如果查询根节点，则返回所有节点。
`GET`方法, url: `/api/v1/ns`
参数：
- QUERY参数 ns: 查询的ns，`loda`则查询全部节点。

    # 获取所有节点
    curl "http://127.0.0.1:9991/api/v1/ns?ns=loda"
    #返回
    {
      "httpstatus": 200,
      "data": {
        "children": [
          {
            "children": [],
            "id": "3b9da565-5deb-41b7-a954-48d95695765b",
            "name": "pool",
            "type": 0,
            "machinereg": "^$"
          },
          {
            "children": [
              {
                "children": [],
                "id": "bc48398c-4270-4f62-9e02-bca2bc948582",
                "name": "server1",
                "type": 0,
                "machinereg": "^$"
              }
            ],
            "id": "ee15f627-6d28-48ac-bbcd-ab2aeaef71b7",
            "name": "product1",
            "type": 1,
            "machinereg": "product1"
          },
          {
            "children": [],
            "id": "04cc1d33-e6ca-40fd-b4b4-dc0677bcb009",
            "name": "server1",
            "type": 0,
            "machinereg": "^$"
          }
        ],
        "id": "0",
        "name": "loda",
        "type": 1,
        "machinereg": "^$"
      },
      "msg": ""
     }

    # 获取某个节点及其子节点
    curl "http://127.0.0.1:9991/api/v1/ns?ns=product0.loda"
    # 返回
    {
      "httpstatus": 200,
      "data": {
        "children": [],
        "id": "3b9da565-5deb-41b7-a954-48d95695765b",
        "name": "pool",
        "type": 0,
        "machinereg": "^$"
      },
      "msg": ""
    }


#### 1.3 修改节点

根据参数修改ns的Name/MachineReg属性，如果未提供则保持不变。

`PUT`方法, url: `/api/v1/ns`
需要提供3个参数：
- QUERY参数 ns: 带修改的节点ns
- QUERY参数 name（可选）: 节点新name **此参数会改变此节点及子节点的ns，请注意**
- QUERY参数 machinereg（可选）: 修改节点的机器匹配规则，请根据需求慎重修改

    curl -X PUT -d 'ns=server1.loda&name=server1-test&machinereg=server1.loda' "http://127.0.0.1:9991/api/v1/ns"
    # 返回
    {"httpstatus": 200, "data": null, "msg": "success"}

    curl -X PUT "http://127.0.0.1:9991/api/v1/ns?ns=server1-test.loda&name=server1-test2&machinereg=^*"

#### 1.4 节点删除

从节点删除一个子节点。目前只允许删除**叶子节点**或者**没有子节点的非叶子节点**，**不允许**删除拥有`机器资源`的叶子节点。
注意：节点删除后，节点的资源会一并删除。

`DELETE`方法, url: `/api/v1/ns`
需要参数：
- QUERY参数 ns：需要删除的ns

    curl -X DELETE "http://127.0.0.1:9991/api/v1/ns?ns=server1-test2.loda"
    # 返回
    {"httpstatus": 200, "data": "", "msg":"success"}


### 2 资源接口
---

某个服务下，所有可以抽象成json的属性都可以看做是这个服务的资源，日常工作就可以针对服务及其各种属性进行展开。

每个叶子节点拥有各种资源，包括`机器`、`监控`、`报警`、`发布`等。

非叶子节点下可以保存各种资源的模板。当建立新叶子节点时，会根据父节点中的模板进行资源初始化，例如初始化监控及报警资源等。

#### 2.1 设置资源

设置节点下的某种资源。

`POST`方法, url: `/api/v1/resource`

提供参数：
- body参数：JSON

    type bodyParam struct {
	    Ns                  string              `json:"ns"`
	    ResourceType        string              `json:"type"`
        ResourceList        []map[string]string `json:"resourcelist"`
    }

例子：

    # 设置pool节点的机器资源
    curl -X POST -d '{"ns":"pool.loda","type":"machine","resourcelist":[{"hostname":"127.0.0.2"},{"hostname":"127.0.0.1"}]}' "http://127.0.0.1:9991/api/v1/resource"
    # 返回
    {"httpstatus":200,"data":null,"msg":"success"}
    
    # 设置server0.product0.loda的机器资源
    curl -X POST -d '{"ns":"server0.product0.loda","type":"machine","resourcelist":[{"hostname":"127.0.0.2"},{"hostname":"127.0.0.3"}]}' "http://127.0.0.1:9991/api/v1/resource"

#### 2.2 添加资源

在节点下添加一个资源。要添加的资源pk不为空。如果资源pk值已经存在于该节点下，则不予添加。

目前各资源对应的pk属性：`machine资源`： `hostname`， 其他资源: `name`。
**系统会自动添加监控类型到collect资源的name前，比如：PROC.bin/PLUGIN.service/PORT.service.xx**

`POST`方法, url: `/api/v1/resource/add`

提供参数：
- body参数：JSON

    type bodyParam struct {
    	Ns             string             `json:"ns"`
    	ResourceType   string             `json:"type"`
    	Resource       model.Resource     `json:"resource"`
    }


    curl -X POST -d '{"ns": "pool.loda", "type":"machine", "resource": {"hostname":"127.0.0.255"}}' "http://127.0.0.1:9991/api/v1/resource/add"
    # 返回
    {"httpstatus":200,"data":null,"msg":"bebf14c6-d5ad-48df-9cfb-0c75f7d3a505"}


#### 2.3 查询资源

获取叶子节点的某种资源，或者非叶子节点的所有子节点的某种资源。

`GET`方法, url: `/api/v1/resource`

提供参数：
- QUERY参数 ns：资源所在的叶子节点ns
- QUERY参数 type：资源类型

例子:

    # 获取pool节点的机器资源
    curl "http://127.0.0.1:9991/api/v1/resource?ns=pool.loda&type=machine"
    # 返回
    {"httpstatus":200,"data":[{"_id":"2d472e17-09cc-475c-a937-5f21f829c355","hostname":"127.0.0.2"},{"_id":"642944d9-34f3-499b-826e-0585b988b46f","hostname":"127.0.0.3"}]}

    # 获取server0.product0.loda节点的采集资源
    curl "http://127.0.0.1:9991/api/v1/resource?ns=server0.product0.loda&type=collect"
    # 获取所有叶子节点的机器资源
    curl "http://127.0.0.1:9991/api/v1/resource?ns=loda&type=machine"


#### 2.4 搜索资源

在叶子节点或者非叶子节点的所有子节点中查找某种资源。

可以进行模糊查找，或者在资源的所有属性中进行查找。比如：
- 查找ip为10.10.10.*的机器
- 查找ip或者hostname包含10的机器

如果查询非叶子节点的某种资源(非模板)，则对该节点下所有叶子节点进行查询。

`GET`方法, url: `/api/v1/resource/search`
提供参数：
- query参数 ns：资源所在的叶子节点ns
- query参数 type：资源类型
- query参数 mod: 搜索类型，fuzzy为模糊搜索。功能上同strings.contain()进行搜索
- query参数 k/v： 搜索的属性k-v。如果k为空，则搜索资源的所有属性。

例子:

    curl "http://127.0.0.1:9991/api/v1/resource/search?ns=loda&type=machine&k=hostname&v=127.0.0.2&mod=exact"|jq
    #返回
    {
      "httpstatus": 200,
      "data": {
        "pool.loda": [
          {
            "_id": "9e324584-17ff-4a12-99d4-78841c62b0bd",
            "hostname": "127.0.0.2"
          }
        ],
        "server1.product1.loda": [
          {
            "_id": "2d472e17-09cc-475c-a937-5f21f829c355",
            "hostname": "127.0.0.2"
          }
        ]
      }
    }
    # 搜索ID
    curl "http://127.0.0.1:9991/api/v1/resource/search?ns=loda&type=machine&k=_id&v=9fc20a93-6642-4f4b-8d15-5847f2232790"
    # 返回
    {
      "httpstatus": 200,
      "data": {
        "server0.product0.loda": [
        {
            "_id": "9fc20a93-6642-4f4b-8d15-5847f2232790",
            "hostname": "127.0.0.3"
        }
        ]
      },
      "msg": ""
    }
    curl "http://127.0.0.1:9991/api/v1/resource/search?ns=pool.loda&type=collect&k=name&v=cpu.idle&mod=exact"|jq
    #返回
    {
      "httpstatus": 200,
      "data": {
        "pool.loda": [
        {
          "_id": "052d44be-6150-49f9-af0d-003d6e9f7646",
          "comment": "",
          "interval": "10",
          "measurement_type": "CPU",
          "name": "cpu.idle"
        }
        ]
      },
      "msg": ""
    }
    curl "http://127.0.0.1:9991/api/v1/resource/search?ns=loda&type=machine&k=hostname&v=127.0.0.&mod=fuzzy"|jq
    #返回
    {
      "httpstatus": 200,
      "data": {
        "pool.loda": [
          {
            "_id": "9e324584-17ff-4a12-99d4-78841c62b0bd",
            "hostname": "127.0.0.2"
          },
          {
            "_id": "689d63c2-14db-422a-b60d-ca6f66ec5348",
            "hostname": "127.0.0.1"
          }
        ],
        "server1.product1.loda": [
          {
            "_id": "2d472e17-09cc-475c-a937-5f21f829c355",
            "hostname": "127.0.0.2"
          },
          {
            "_id": "642944d9-34f3-499b-826e-0585b988b46f",
            "hostname": "127.0.0.3"
          }
        ]
      }
    }

#### 2.5 修改资源

根据`map`对资源进行修改，如果属性未出现在修改map中，则不予变更。

`PUT`方法, url: `api/v1/resource`

提供参数:
- body 参数: JSON

    type bodyParam struct {
    	Ns             string             `json:"ns"`
    	ResourceType   string             `json:"type"`
    	ResourceId     string             `json:"resourceid"`
    	UpdateMap      map[string]string  `json:"update"`
    }

例子

    # 修改pool节点的某台机器的备注
    curl -X PUT -d'{"ns": "pool.loda", "type": "machine", "resourceid": "1b7a5cac-a875-4062-ba9e-c24319cb27df", "update":{"comment":"new comment"}}' "http://127.0.0.1:9991/api/v1/resource"
    # 返回
    {"httpstatus":200,"data":"success"}


#### 2.6 删除资源

在节点下删除一个资源

`DELETE`方法, url: `/api/v1/resource`

提供参数：
- Query参数 ns：资源所在的叶子节点ns
- Query参数 type： 需要删除的资源类型
- Query参数 resourceid: 需要删除的资源ID

例子：

    curl -X DELETE "http://127.0.0.1:9991/api/v1/resource?ns=pool.loda&type=machine&resourceid=1b7a5cac-a875-4062-ba9e-c24319cb27df"
    # 返回
    {"httpstatus":200,"data":null,"msg":"success"}

#### 2.7 移动资源

将资源移动到拎一个节点。

`PUT`方法, url: `/api/v1/resource/move`

提供参数：
- Query参数 from：资源当前所在的ns
- Query参数 to： 需要移动到的目的ns
- Query参数 type: 资源类型
- Query参数 resourceid: 资源ID

例子：

    curl -X PUT "http://127.0.0.1:9991/api/v1/resource/move?from=pool.loda&to=server0.product0.loda&type=machine&resourceid=d0f769bf-1e2c-4cae-85ad-61e24f1ea96d"

#### 2.8 删除监控

删除监控资源的同时，删除该监控的采集数据。

`DELETE`方法，url:`/api/v1/resource/collect`
提供参数：
- Query参数 ns：资源所在的叶子节点ns
- Query参数 measurements: 要删除的监控项

例子：

    curl -X DELETE "http://127.0.0.1:9991/api/v1/resource/collect?ns=pool.loda&measurements=PORT.test,PLUGIN.test.cpu"


#### 2.9 修改机器状态

搜索机器名，修改这个机器在所有节点下的机器状态。设置type为`machine`，读取`resourceid`为`hostname`，用以搜索机器并更新机器资源。
`PUT`方法，url:`/api/v1/machine/status`

    curl -X PUT  -H 'Resource: machine' -H 'NS: loda' -H 'AuthToken: xxx' -d'[{"ns": "loda", "type": "machine", "resourceid": "hostname", "update":{"status":"online"}}]' 'http://127.0.0.1:9991/api/v1/resource/list'

### 3 agent相关接口
---

根据hostname进行节点查找/注册:
- 如果机器已存在，则返回`map{ns:资源ID}`
- 如果机器不在任何节点下，则注册到hostname对应节点下，如果未匹配到任何节点，则注册到`pool.loda`中。

#### 3.1 注册接口

`POST`方法

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


#### 3.2 agent获取资源

`GET`方法, url: `/api/v1/agent/resource`

提供参数：
- QUERY参数 ns：资源所在的叶子节点ns
- QUERY参数 type：资源类型

    curl "http://127.0.0.1:9991/api/v1/agent/resource?ns=pool.loda&type=collect"

#### 3.3 上报接口
---

接受将agent上报agent版本信息，并根据hostname将agent版本信息保存到机器资源中。

如果`update`为`true`，则变更有变化的`hostname`或`ip`。*注意: 为了定位变更机器，必须提交oldhostname并且值为变更之前的hostname。*如果提交`update`为`true`但新旧参数无变化或者不合法，则不予变更。

POST方法
提供参数:
- body参数: lodastack/models.Report

    curl -X POST -d '{"update": true, "newhostname": "new-hostname", "oldhostname": "old-hostname", "oldiplist": ["127.0.0.1"], "newiplist": ["10.10.10.10", "127.0.0.1"]}' "http://127.0.0.1:9991/api/v1/agent/report"


### 4 权限
---

#### 4.1 登录接口
`POST`方法

提供参数：
- username 用户名
- password 密码

结果返回：
- `{"user":"libk","token":"39dfcfb7-5f2b-45dc-b99f-6f0011d9dcc7"}`

例子：

    curl -d "username=name&password=pwd" "http://127.0.0.1:8001/api/v1/user/signin"
    {
      "httpstatus": 200,
      "data": {
            "user": "name",
            "token": "3de20444-b3b6-475c-8404-f2b471b64a85"
      },
      "msg": ""
    }



#### 4.2 登出接口
`GET`方法

提供参数：
- header中的AuthToken

结果返回：
    {
      "user": "libk",
      "token": "39dfcfb7-5f2b-45dc-b99f-6f0011d9dcc7"
    }

例子:

    curl "http://127.0.0.1:8004/api/v1/user/signout"


#### 4.3 用户查询

`GET`方法

参数：
- query参数 username

    curl -H "AuthToken: d1a02e5e-d1d3-4c9c-9fe5-e8eeccbd8ee4" -H "NS: loda" -H "Resource: ns" "http://127.0.0.1:9991/api/v1/perm/user?username=zhangzz"|jq

#### 4.4 用户设置

`PUT`方法

参数：
- query参数 username: 要更改的用户名
- query参数 dashboards：**保留**

例子:

    curl -X PUT -H "AuthToken: d1a02e5e-d1d3-4c9c-9fe5-e8eeccbd8ee4"  -H "NS: loda" -H "Resource: ns" "http://127.0.0.1:9991/api/v1/perm/user?username=test&gids=&dashboards="

#### 4.5 组创建

`POST`方法

参数：
- query参数 gname: 用户组name
- query参数 items: 用户的组权限列表，`,`分隔

    curl  -H "AuthToken: aeaec15e-5601-4bd9-a81a-b1a096244a8c" -H "NS: loda" -H "Resource: ns" -X POST "http://127.0.0.1:9991/api/v1/perm/group?gname=test&items=1,2,33"
    {
      "httpstatus": 404,
      "data": null,
      "msg": "group already exist"
    }

#### 4.6 组查询

`GET`方法

参数：
- query参数 gname: 用户组name

例子:

    curl -H "AuthToken: d1a02e5e-d1d3-4c9c-9fe5-e8eeccbd8ee4" -H "NS: loda" -H "Resource: ns" "http://127.0.0.1:9991/api/v1/perm/group?gname=loda-defaultgroup
    "|jq

#### 4.7 用户组成员管理

`PUT`方法

参数：
- query参数 gname: 要更改的用户组
- query参数 managers: 用户的组管理员列表（不设置为不更改）
- query参数 members: 用户组成员列表

例子:

    curl  -H "AuthToken: 9fc601ba-f834-4a53-a289-bb83ffd0cc15" -H "NS: loda" -H "Resource: ns" -X PUT "http://127.0.0.1:9991/api/v1/perm/group/member?gname=loda.test.leaf-test&action=add&members=fuyin"
    curl  -H "AuthToken: 9fc601ba-f834-4a53-a289-bb83ffd0cc15" -H "NS: loda" -H "Resource: ns" -X PUT "http://127.0.0.1:9991/api/v1/perm/group/member?gname=loda.test.leaf-test&action=add&managers=fuyin"

#### 4.8 用户组权限管理

`GET`方法

参数：
- query参数 gname: 要更改的用户组
- query参数 items: 用户的组权限列表，`,`分隔

例子:

    curl  -H "AuthToken: aeaec15e-5601-4bd9-a81a-b1a096244a8c" -H "NS: loda" -H "Resource: ns" -X PUT "http://127.0.0.1:9991/api/v1/perm/group/item?gname=test&items=1,2,33,444"

#### 4.9 NS下用户组查询

`GET`方法

参数：
- query参数 ns: 要查询的group列表所属ns

例子:

    curl "http://127.0.0.1:9991/api/v1/perm/group/list?ns=server0.loda"
    {
        "httpstatus": 200,
        "data": [
            {
                "name": "loda.server0-admin",
                "manager": [
                    "loda-defaultuser",
                    "zhangzz"
                ],
                "member": [
                    "loda-defaultuser",
                    "zhangzz"
                ],
                "items": [
                    "server0.loda-machine-GET",
                    "server0.loda-machine-PUT",
                    "server0.loda-machine-POST",
                    "server0.loda-machine-DELETE",
                    "server0.loda-alarm-GET",
                    "server0.loda-alarm-PUT",
                    .......
                ]
            },
            .....
            {
                "name": "loda.server0test2",
                "manager": null,
                "member": null,
                "items": [
                  "1",
                  "2",
                  "33"
                ]
          }
      ],
    "msg": ""
  }

#### 4.10 用户组删除

`DELETE`方法

参数：
- query参数 gname: 用户组name

例子:

    curl  -H "AuthToken: aeaec15e-5601-4bd9-a81a-b1a096244a8c" -H "NS: loda" -H "Resource: ns" -X DELETE "http://127.0.0.1:9991/api/v1/perm/group?gname=test"|jq

#### 4.11 权限验证接口

根据header中携带的`token`、`ns`、`resource`以及`http method`信息，验证请求有权限进行操作。方便第三方进行权限验证。
如果有权限则返回`200`，token失效则返回`401`，无权限返回`403`.

例子

    curl -X GET -H "AuthToken: xxxxx-xxx-xxx-xxxxxx" -H "NS: pool.loda" -H "Resource: machine" "http://127.0.0.1:9991/api/v1/perm/check"
    curl -X POST -H "AuthToken: xxxxx-xxx-xxx-xxxxxx" -H "NS: pool.loda" -H "Resource: machine" "http://127.0.0.1:9991/api/v1/perm/check"
    curl -X PUT -H "AuthToken: xxxxx-xxx-xxx-xxxxxx" -H "NS: pool.loda" -H "Resource: machine" "http://127.0.0.1:9991/api/v1/perm/check"
    curl -X DELETE -H "AuthToken: xxxxx-xxx-xxx-xxxxxx" -H "NS: pool.loda" -H "Resource: machine" "http://127.0.0.1:9991/api/v1/perm/check"
