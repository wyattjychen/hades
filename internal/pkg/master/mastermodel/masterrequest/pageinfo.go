package masterrequest

type (
	PageInfo struct {
		Page     int `json:"page" form:"page"`
		PageSize int `json:"page_size" form:"page_size"`
	}
	ByID struct {
		ID int `json:"id" form:"id"`
	}
	ByIDS struct {
		IDS []int `json:"ids" form:"ids"`
	}
)

func (page *PageInfo) Check() {
	if page.PageSize <= 0 {
		page.PageSize = 20
	}
	if page.Page <= 0 {
		page.Page = 1
	}
}
