package models

type LoginRequest struct {
	Phone     string      `json:"phone"`
	Password  string      `json:"password"`
	VCode     string      `json:"v_code"`
	PhoneArea string      `json:"phone_area"`
	PictureID interface{} `json:"picture_id"`
}

type LoginResponse struct {
	Code string `json:"code"`
	Data struct {
		Ticket string `json:"ticket"`
	} `json:"data"`
	Msg string `json:"msg"`
}

type PassportRequest struct {
	Ticket string `json:"ticket"`
}

type PassportResponse struct {
	Code string `json:"code"`
	Data struct {
		Token string `json:"token"`
	} `json:"data"`
	Msg string `json:"msg"`
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
}

type InstanceResponse struct {
	Code string `json:"code"`
	Data struct {
		List []Instance `json:"list"`
	} `json:"data"`
	Msg string `json:"msg"`
}
