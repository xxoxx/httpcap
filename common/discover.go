package common

const (
	Service_Type_Redis     = 0
	Service_Type_Memcache  = 1
	Service_Type_Twemproxy = 2
	Service_Type_Mysql     = 3
	Service_Type_Mongodb   = 4
)

type Service struct {
	Port int
	Type int
	Pid  int
}
