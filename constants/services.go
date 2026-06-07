package constants

const (
	UserService    = "user-service"
	ProductService = "product-service"
)

var ServiceMap = map[string]string{
	UserService:    "http://localhost:8081",
	ProductService: "http://localhost:8082",
}
