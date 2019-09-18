package cmd

type API int

type Ping struct {
	Id []byte
}

type Put struct {
	Id  []byte
	Val string
}

type Get struct {
	Id  []byte
	Key []byte
}

type Exit struct {
	Id []byte
}

func (a *API) Ping(ping Ping, reply *bool) error {
	*reply = true
	return nil
	//TODO
}

func (a *API) Put(put Put, reply *bool) error {
	*reply = true
	return nil
	//TODO
}

func (a *API) Get(get Get, reply *bool) error {
	*reply = true
	return nil
	//TODO
}

func (a *API) Exit(exit Exit, reply *bool) error {
	*reply = true
	return nil
	//TODO
}
