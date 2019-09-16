package cmd

type API int

type Ping struct {
	Id []byte
}

type Exit struct {
	Id []byte
}

func (a *API) Ping(ping Ping, reply *bool) error {
	*reply = true
	return nil
	//TODO
}

func (a *API) Exit(exit Exit, reply *bool) error {
	*reply = true
	return nil
	//TODO
}
