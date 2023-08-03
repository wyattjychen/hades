package masterrequest

type ByUUID struct {
	UUID string `json:"uuid" form:"uuid"`
}

type ReqNodeSearch struct {
	PageInfo
	IP     string `json:"ip" form:"ip"` // node ip
	UUID   string `json:"uuid" form:"uuid"`
	UpTime int64  `json:"up" form:"up"`         // start time
	Status int    `json:"status" form:"status"` // status
}
