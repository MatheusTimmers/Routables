package router

type Router struct {
	IP        string
	RouteTable map[string]Route
}

func NewRouter(ip string) *Router {
	return &Router{
		IP:         ip,
		RouteTable: make(map[string]Route),
	}
}

func Start()  {
  
}
