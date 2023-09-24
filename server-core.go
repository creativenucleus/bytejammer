package main

type ServerCore struct {
}

func NewServerCore() *ServerCore {
	return &ServerCore{}
}

func (s *ServerCore) linkClient(c *ClientJukebox) {
	/*
		go func() {
			for {
				select c.comms <- msg {
				case "code":
					fmt.Println("-> code:", string(msg.Data))
				}
			}
		}()
	*/
}
