package hub

type Hub struct {
    Clients    map[*Client]bool
    Channels   map[string]map[*Client]bool
    Register   chan *Client
    Unregister chan *Client
    Broadcast  chan Message
}

func NewHub() *Hub {
    return &Hub{
        Clients:    make(map[*Client]bool),
        Channels:   make(map[string]map[*Client]bool),
        Register:   make(chan *Client),
        Unregister: make(chan *Client),
        Broadcast:  make(chan Message),
    }
}

func (h *Hub) Run() {
    for {
        select {
        case client := <-h.Register:
            h.Clients[client] = true
            if h.Channels[client.Channel] == nil {
                h.Channels[client.Channel] = make(map[*Client]bool)
            }
            h.Channels[client.Channel][client] = true

        case client := <-h.Unregister:
            if _, ok := h.Clients[client]; ok {
                delete(h.Clients, client)
                if clients, ok := h.Channels[client.Channel]; ok {
                    delete(clients, client)
                }
                close(client.Send)
            }

        case msg := <-h.Broadcast:
            if clients, ok := h.Channels[msg.Channel]; ok {
                for client := range clients {
                    select {
                    case client.Send <- msg.Data:
                    default:
                        close(client.Send)
                        delete(h.Clients, client)
                    }
                }
            }
        }
    }
}