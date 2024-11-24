package models

type LoginRequest struct {
	Phone     string      `json:"phone"`
	Password  string      `json:"password"`
	VCode     string      `json:"v_code"`
	PhoneArea string      `json:"phone_area"`
	PictureID interface{} `json:"picture_id"`
}

type LoginResponse struct {
	Code string    `json:"code"`
	Data LoginData `json:"data"`
	Msg  string    `json:"msg"`
}

type LoginData struct {
	Ticket string `json:"ticket"`
}

type PassportRequest struct {
	Ticket string `json:"ticket"`
}

type PassportResponse struct {
	Code string       `json:"code"`
	Data PassportData `json:"data"`
	Msg  string       `json:"msg"`
}

type PassportData struct {
	Token string `json:"token"`
}

type InstanceRequest struct {
	DateFrom   string   `json:"date_from"`
	DateTo     string   `json:"date_to"`
	PageIndex  int      `json:"page_index"`
	PageSize   int      `json:"page_size"`
	Status     []string `json:"status"`
	ChargeType []string `json:"charge_type"`
}

type Instance struct {
	MachineAlias string `json:"machine_alias"`
	RegionName   string `json:"region_name"`
	GpuAllNum    int    `json:"gpu_all_num"`
	GpuIdleNum   int    `json:"gpu_idle_num"`
	UUID         string `json:"uuid"`
	StoppedAt    struct {
		Time string `json:"time"`
	} `json:"stopped_at"`
}

type InstanceResponse struct {
	Code string `json:"code"`
	Data struct {
		List []Instance `json:"list"`
	} `json:"data"`
	Msg string `json:"msg"`
}

type AutoDLConfig struct {
	Username string
	Password string
}
