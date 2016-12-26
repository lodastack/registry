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
`POST`方法，url: `/api/v1/ns`
参数：
- QUERY参数 ns: 父节点的节点Ns
- QUERY参数 type: 节点类型，`0`为叶子节点，`1`为非叶子节点
- QUERY参数 name：节点名称，用于组成节点ns
- QUERY参数 matchreg: 机器政策匹配规则，如果新机器匹配到规则，则注册到该节点下。默认不进行匹配

成功返回：JSON数据`{"httpstatus": 200/400/404/500, "data":JSON, "msg":"error msg"}`，httpstatus为返回的http状态码，data为Json, msg为string **如非特别说明，ns及资源接口返回格式相同**

例子  （初始节点ID为`0`）

    # 新建非叶子节点
    curl -X POST "http://127.0.0.1:9991/api/v1/ns?ns=loda&name=product0&type=1&machinereg=product0"
    #返回
    {"httpstatus":200,"data":null,"msg":"success"}
    
    #新建叶子节点
    curl -X POST -d 'ns=loda&name=server0&type=0&matchreg=server0.loda' "http://127.0.0.1:9991/api/v1/ns"
    curl -X POST "http://127.0.0.1:9991/api/v1/ns?ns=loda&name=server1&type=0&matchreg=server1.loda"
    
    #在prodect1.loda下新建叶子节点
    curl -X POST -d 'ns=product0.loda&name=server0&type=0' "http://127.0.0.1:9991/api/v1/ns"
    curl -X POST "http://127.0.0.1:9991/api/v1/ns?ns=product0.loda&name=server1&type=0&machinereg=server1.product0"

    # 错误：在叶子节点下新建节点
    curl -X POST -d 'ns=server1.loda&name=test&type=0' "http://127.0.0.1:9991/api/v1/ns"
    # 返回
    {"httpstatus": 500, "data": null, "msg": "can not create node under leaf node"}

#### 1.2 查询节点
查询全部节点
`GET`方法, url: `/api/v1/ns`
参数：
- QUERY参数 ns: 查询的ns，`loda`则查询全部节点。

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

根据节点ID查询节点

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

根据参数修改ns的Name/MachineReg属性。如果未提供则保持不变。

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

从节点删除一个子节点。目前只允许删除**叶子节点**或者**没有子节点的非叶子节点**

`DELETE`方法, url: `/api/v1/ns`
需要参数：
- QUERY参数 ns：需要删除的ns

    curl -X DELETE "http://127.0.0.1:9991/api/v1/ns?ns=server1-test2.loda"
    # 返回
    {"httpstatus": 200, "data": "", "msg":"success"}

### 2 资源接口
---

#### 2.1 设置资源

只能在叶子节点下设置资源，目前只能设置全量资源，不能追加资源。

`POST`方法, url: `/api/v1/resource`

提供参数：
- body参数：JSON

    type bodyParam struct {
	    rl        []map[string]string `json:"resourcelist"`
	    ns        string              `json:"ns"`
	    resType   string              `json:"type"`
    }

例子：

    curl -X POST -d '{"ns":"pool.loda","type":"machine","resourcelist":[{"hostname":"127.0.0.2"},{"hostname":"127.0.0.1"}]}' "http://127.0.0.1:9991/api/v1/resource"
    # 返回
    {"httpstatus":200,"data":null,"msg":"success"}
    curl -X POST -d '{"ns":"server0.product0.loda","type":"machine","resourcelist":[{"hostname":"127.0.0.2"},{"hostname":"127.0.0.3"}]}' "http://127.0.0.1:9991/api/v1/resource"

#### 2.2 添加资源

在节点下添加一个资源项。如果要添加的pk属性缺失或已经在该节点下存在，则不予添加。
目前各资源对应的pk属性：`machine`资源： `hostname`， 其他资源: `name`

`POST`方法, url: `/api/v1/resource/add`

提供参数：
- body参数：JSON

    type bodyParam struct {
    	Ns        string             `json:"ns"`
    	ResType   string             `json:"type"`
    	R         model.Resource     `json:"resource"`
    }


   curl -X POST -d '{"ns": "pool.loda", "type":"machine", "resource": {"hostname":"127.0.0.255"}}' "http://127.0.0.1:9991/api/v1/resource/add"
    # 返回
    {"httpstatus":200,"data":null,"msg":"bebf14c6-d5ad-48df-9cfb-0c75f7d3a505"}


#### 2.3 查询资源

如果查询非叶子节点的某种资源(非模板)，则对该节点下所有叶子节点进行查询。

`GET`方法, url: `/api/v1/resource`

提供参数：
- QUERY参数 ns：资源所在的叶子节点ns
- QUERY参数 type：资源类型

例子:

    curl "http://127.0.0.1:9991/api/v1/resource?ns=pool.loda&type=machine"
    # 返回
    {"httpstatus":200,"data":[{"_id":"2d472e17-09cc-475c-a937-5f21f829c355","hostname":"127.0.0.2"},{"_id":"642944d9-34f3-499b-826e-0585b988b46f","hostname":"127.0.0.3"}]}

    curl "http://127.0.0.1:9991/api/v1/resource?ns=server0.product0.loda&type=collect"
    curl "http://127.0.0.1:9991/api/v1/resource?ns=loda&type=machine"


#### 2.4 搜索资源

`GET`方法, url: `/api/v1/resource/search`
提供参数：
- query参数 ns：资源所在的叶子节点ns
- query参数 type：资源类型
- query参数 mod: 搜索类型，fuzzy为模糊搜索。功能上同strings.contain()进行搜索
- query参数 k/v： 搜索的属性k-v

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

`PUT`方法, url: `api/v1/resource`

提供参数:
- body 参数: JSON

    type bodyParam struct {
    	Ns        string             `json:"ns"`
    	ResType   string             `json:"type"`
    	ResId     string             `json:"resourceid"`
    	UpdateMap map[string]string  `json:"update"`
    }

    curl -X PUT -d'{"ns": "pool.loda", "type": "machine", "resourceid": "1b7a5cac-a875-4062-ba9e-c24319cb27df", "update":{"comment":"new comment"}}' "http://127.0.0.1:9991/api/v1/resource"
    # 返回
    {"httpstatus":200,"data":"success"}


#### 2.6 删除资源

在节点下删除一个资源项

`DELETE`方法, url: `/api/v1/resource`

提供参数：
- Query参数 ns：资源所在的叶子节点ns
- Query参数 type： 需要删除的资源类型
- Query参数 resourceid: 需要删除的资源ID

    curl -X DELETE "http://127.0.0.1:9991/api/v1/resource?ns=pool.loda&type=machine&resourceid=1b7a5cac-a875-4062-ba9e-c24319cb27df"
    # 返回
    {"httpstatus":200,"data":null,"msg":"success"}

#### 2.7 移动资源

`PUT`方法, url: `/api/v1/resource/move`

提供参数：
- Query参数 from：资源当前所在的ns
- Query参数 to： 需要移动到的目的ns
- Query参数 type: 资源类型
- Query参数 resourceid: 资源ID

     curl -X PUT "http://127.0.0.1:9991/api/v1/resource/move?from=pool.loda&to=server0.product0.loda&type=machine&resourceid=d0f769bf-1e2c-4cae-85ad-61e24f1ea96d"

### 3 注册接口
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

    curl -d "username=libk&password=test" "http://127.0.0.1:8004/api/v1/user/signin"



#### 4.2 登出接口
`GET`方法

提供参数：
- header中的AuthToken

结果返回：
- `{
  "user": "libk",
  "token": "39dfcfb7-5f2b-45dc-b99f-6f0011d9dcc7"
}`

例子：
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
- query参数 gids: 用户的组id（不设置为不更改）
- query参数 dashboards：**保留**

    curl -X PUT -H "AuthToken: d1a02e5e-d1d3-4c9c-9fe5-e8eeccbd8ee4"  -H "NS: loda" -H "Resource: ns" "http://127.0.0.1:9991/api/v1/perm/user?username=test&gids=&dashboards="

#### 4.5 用户组查询

`GET`方法

参数：
- query参数 gids: 用户的组id

    curl -H "AuthToken: d1a02e5e-d1d3-4c9c-9fe5-e8eeccbd8ee4" -H "NS: loda" -H "Resource: ns" "http://127.0.0.1:9991/api/v1/perm/group?gid="|jq

#### 4.6 用户组设置

`PUT`方法

参数：
- query参数 gid: 要更改的用户组ID
- query参数 managers: 用户的组管理员列表（不设置为不更改）
- query参数 items：权限列表（不设置为不更改）

curl -X PUT -H "AuthToken: d1a02e5e-d1d3-4c9c-9fe5-e8eeccbd8ee4"  -H "NS: loda" -H "Resource: ns" "http://127.0.0.1:9991/api/v1/perm/group?gid=42f95995-79a9-44fe-9fbf-417bffbf2035&managers=zhangzz,libk,loda-ma"

### 5 上报接口
---

如果`update`为`true`，则变更有变化的`hostname`或`ip`。*注意: 为了定位变更机器，必须提交oldhostname并且值为变更之前的hostname。*如果提交`update`为`true`，但新旧参数无变化或者不合法则不予变更。

POST方法
提供参数:
- body参数: lodastack/models.Report

    curl -X POST -d '{"update": true, "oldhostname": "old-hostname", "oldiplist": ["127.0.0.1"], "newiplist": ["10.10.10.10", "127.0.0.1"]}' "http://127.0.0.1:9991/api/v1/agent/report"
