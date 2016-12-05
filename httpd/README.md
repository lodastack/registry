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
`POST`方法，url: `/api/v1/ns/:parentNs`
参数：
- URI参数 parentNs: 父节点的节点Ns
- QUERY参数 type: 节点类型，0为叶子节点，1为非叶子节点
- QUERY参数 name：节点名称，用于组成节点ns
- QUERY参数 matchreg: 机器政策匹配规则，如果新机器匹配到规则，则注册到该节点下。默认不进行匹配

成功返回：JSON数据`{"httpstatus": 200/400/404/500, "data":JSON, "msg":"error msg"}`，httpstatus为返回的http状态码，data为Json, msg为string **如非特别说明，ns及资源接口返回格式相同**

例子  （初始节点ID为`0`）

    # 新建非叶子节点
    curl -X POST "http://127.0.0.1:9991/api/v1/ns/loda?type=1&name=product1"
    #返回
    {"httpstatus":200,"data":null,"msg":"816442ae-5c9d-44fe-b03c-6bd6a4df7fc7"}
    
    #新建叶子节点
    curl -X POST "http://127.0.0.1:9991/api/v1/ns/loda?type=0&name=server1"
    curl -X POST "http://127.0.0.1:9991/api/v1/ns/loda?&type=0&name=server2&machinereg=server2-machine"
    
    #在prodect1.loda下新建叶子节点
    curl -X POST "http://127.0.0.1:9991/api/v1/ns/product1.loda?type=0&name=server1"

#### 1.2 查询节点
查询全部节点
`GET`方法, url: `/api/v1/ns/:ns`
参数：
- URI参数 ns: 查询的ns，`/loda`则查询全部节点。

    curl "http://127.0.0.1:9991/api/v1/ns/loda"
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

    curl "http://127.0.0.1:9991/api/v1/ns/product1.loda"
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

`PUT`方法, url: `/api/v1/ns/:ns`
需要提供3个参数：
- URI参数 ns: 带修改的节点ns
- name（可选）: 节点新name **此参数会改变此节点及子节点的ns，请注意**
- machinereg（可选）: 修改节点的机器匹配规则，请根据需求慎重修改

    curl -X PUT "http://127.0.0.1:9991/api/v1/ns/product1.loda?machinereg=product1"
    # 返回
    {"httpstatus": 200, "data": "", "msg": "success"}

    curl -X PUT "http://127.0.0.1:9991/api/v1/ns/product3.loda?name=product2&machinereg=product2"

#### 1.4 节点删除

从节点删除一个子节点。

`DELETE`方法, url: `/api/v1/ns/:ns`
需要参数：
- URI参数 ns：需要删除的ns

    curl -X DELETE "http://127.0.0.1:9991/api/v1/ns/server2.loda"
    # 返回
    {"httpstatus": 200, "data": "", "msg":"success"}

### 2 资源接口
---

#### 2.1 设置资源

只能在叶子节点下设置资源，目前只能设置全量资源，不能追加资源。

`POST`方法, url: `/api/v1/resource/:ns/:type`
sss
提供参数：
- URI参数 ns：资源所在的叶子节点ns
- URI参数 type：资源类型
- body参数：资源内容格式为maplist，系统会给每个资源 ***添加资源ID***

例子：

    curl -d '[{"hostname":"127.0.0.2"},{"hostname":"127.0.0.1"}]' "http://127.0.0.1:9991/api/v1/resource/pool.loda/machine"
    # 返回
    {"httpstatus":200,"data":"success"}
    
    curl -d '[{"hostname":"127.0.0.2"},{"hostname":"127.0.0.3"}]' "http://127.0.0.1:9991/api/v1/resource/server1.product1.loda/machine"

#### 2.2 添加资源

在节点下添加一个资源项

`POST`方法, url: `/api/v1/addresource/:ns/:type`

提供参数：
- URI参数 ns：资源所在的叶子节点ns
- URI参数 type： 需要添加的资源类型
- body参数：需要添加的资源数据`map[string]string`, 自动生成ID

    curl -X POST -d '{"comment": "loda", "action": "add resource"}' "http://127.0.0.1:9991/api/v1/addresource/pool.loda/doc"
    # 返回
    {"httpstatus":200,"data":null,"msg":"bebf14c6-d5ad-48df-9cfb-0c75f7d3a505"}

#### 2.3 查询资源

如果查询非叶子节点的某种资源(非模板)，则对该节点下所有叶子节点进行查询。

`GET`方法, url: `/api/v1/resource/:ns/:type`

提供参数：
- URI参数 ns：资源所在的叶子节点ns
- URI参数 resouce：资源类型

例子:

     curl "http://127.0.0.1:9991/api/v1/resource/server1.product1.loda/machine"
     # 返回
     {"httpstatus":200,"data":[{"_id":"2d472e17-09cc-475c-a937-5f21f829c355","hostname":"127.0.0.2"},{"_id":"642944d9-34f3-499b-826e-0585b988b46f","hostname":"127.0.0.3"}]}

     curl "http://127.0.0.1:9991/api/v1/resource/pool.loda/machine"
     curl "http://127.0.0.1:9991/api/v1/resource/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxxx/machine"

#### 2.4 搜索资源

`GET`方法, url: `/api/v1/search/:ns/:type`
提供参数：
- URI参数 ns：资源所在的叶子节点ns
- URI参数 type：资源类型
- query参数 mod: 搜索类型，fuzzy为模糊搜索。功能上同strings.contain()进行搜索
- query参数 k/v： 搜索的属性k-v

例子:

    curl "http://127.0.0.1:9991/api/v1/search/loda/machine?k=hostname&v=127.0.0.2&mod=exact"|jq
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
    curl "http://127.0.0.1:9991/api/v1/search/loda/machine?k=hostname&v=127.0.0.&mod=fuzzy"|jq
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

`PUT`方法, url: `api/v1/resource/:ns/:type/:ID`

提供参数:
- URI参数 ns: 修改节点
- URI参数 type: 修改资源类型
- URI参数 ID: 修改资源ID
- body map[string]string: 需要修改的k-v。会忽略修改ID的请求，并修改提交的其他数据。

    curl -X PUT -d'{"comment":"new comment"}' "http://127.0.0.1:9991/api/v1/resource/pool.loda/collect/bd64f882-db3e-4da3-b7ee-40ac7d966726"
    # 返回
    {"httpstatus":200,"data":"success"}

#### 2.6 删除资源

在节点下删除一个资源项

`DELETE`方法, url: `/api/v1/resource/:ns/:type/:ID`

提供参数：
- URI参数 ns：资源所在的叶子节点ns
- URI参数 type： 需要删除的资源类型
- URI参数 ID: 需要删除的资源ID

    curl -X DELETE "http://127.0.0.1:9991/api/v1/resource/pool.loda/doc/xxxx-xxx-xxx-xxxxxxxxxxxxx"
    # 返回
    {"httpstatus":200,"data":null,"msg":"success"}

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


### 4 用户登录接口
---

#### 4.1 登录接口
POST方法

提供参数：
- username 用户名
- password 密码

结果返回：
- `{"user":"libk","token":"39dfcfb7-5f2b-45dc-b99f-6f0011d9dcc7"}`

例子：

    curl -d "username=libk&password=test" "http://127.0.0.1:8004/api/v1/user/signin"



#### 4.2 登出接口
GET方法

提供参数：
- header中的AuthToken

结果返回：
- `{
  "user": "libk",
  "token": "39dfcfb7-5f2b-45dc-b99f-6f0011d9dcc7"
}`

例子：
  curl "http://127.0.0.1:8004/api/v1/user/signout"

### 5 上报接口
---

如果新旧机器名不符，则将全树上的旧机器名改为新机器名

PUT方法
提供参数:
- body参数: lodastack/models.Report

    curl -X PUT -d '{"newhostname":"pool-newname","oldhostname":"pool-machine"}' "http://127.0.0.1:9991/api/v1/agent/report"
