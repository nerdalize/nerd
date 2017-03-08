package nerd

//User holds a reference to a user session
type User struct {
	Region string
}

//GetCurrentUser returns the current user session
func GetCurrentUser() *User {
	//TODO: Get jwt from env variables
	return &User{
		//TODO: This should not be hardcoded.
		Region: "eu-west-1",
	}
}
