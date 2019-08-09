module github.com/lodastack/registry

go 1.13

require (
	github.com/BurntSushi/toml v0.3.0
	github.com/go-ldap/ldap v0.0.0-20180523145351-6e1f1f02400e
	github.com/julienschmidt/httprouter v0.0.0-20180411154501-adbc77eec0d9
	github.com/lodastack/log v0.0.0-20161025094532-b25a4d2e8c22
	github.com/lodastack/models v0.0.0-20190318082849-326fc099cf31
	github.com/lodastack/sdk-go v0.0.0-20170303095045-58b4c40298f6
	github.com/lodastack/store v0.0.0-20190809164939-1389518ecd49
	github.com/miekg/dns v1.0.7
	github.com/pquerna/ffjson v0.0.0-20171002144729-d49c2bc1aa13
	github.com/satori/go.uuid v1.2.0
	golang.org/x/crypto v0.0.0-20180820150726-614d502a4dac // indirect
	gopkg.in/asn1-ber.v1 v1.0.0-20170511165959-379148ca0225 // indirect
	gopkg.in/ldap.v2 v2.5.1 // indirect
	labix.org/v2/mgo v0.0.0-20140701140051-000000000287 // indirect
	launchpad.net/gocheck v0.0.0-20140225173054-000000000087 // indirect
)

replace (
	golang.org/x/crypto v0.0.0-20180820150726-614d502a4dac => github.com/golang/crypto v0.0.0-20180820150726-614d502a4dac
	golang.org/x/net v0.0.0-20180826012351-8a410e7b638d => github.com/golang/net v0.0.0-20180826012351-8a410e7b638d
	golang.org/x/sync v0.0.0-20180314180146-1d60e4601c6f => github.com/golang/sync v0.0.0-20180314180146-1d60e4601c6f
	golang.org/x/sys v0.0.0-20180824143301-4910a1d54f87 => github.com/golang/sys v0.0.0-20180824143301-4910a1d54f87
	google.golang.org/appengine v1.1.0 => github.com/golang/appengine v1.1.0
)
