package payload

//TaskCreateInput is used as input to task creation
type TaskCreateInput struct {
	Image string `json:"image" valid:"min=1,max=64,required"`
}

//TaskCreateOutput is returned from
type TaskCreateOutput struct {
	ID    string `json:"id"`
	Image string `json:"image"`
}
