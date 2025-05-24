package ws

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/lioarce01/remote_patient_monitoring_system/pkg/common/domain/entities"
)

type WSHandler struct {
	upgrader  websocket.Upgrader
	clients   map[*websocket.Conn]bool
	clientsMu sync.Mutex
}

func NewWSHandler() *WSHandler {
	return &WSHandler{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		clients: make(map[*websocket.Conn]bool),
	}
}

func (w *WSHandler) BroadcastAlert(alert *entities.Alert) {
	w.clientsMu.Lock()
	defer w.clientsMu.Unlock()
	for conn := range w.clients {
		if err := conn.WriteJSON(alert); err != nil {
			log.Println("Error al enviar alerta por WebSocket:", err)
			conn.Close()
			delete(w.clients, conn)
		}
	}
}

func (w *WSHandler) Handler() http.HandlerFunc {
	return func(wr http.ResponseWriter, r *http.Request) {
		wsConn, err := w.upgrader.Upgrade(wr, r, nil)
		if err != nil {
			log.Println("Error al actualizar conexión WebSocket:", err)
			return
		}

		w.clientsMu.Lock()
		w.clients[wsConn] = true
		w.clientsMu.Unlock()

		defer func() {
			w.clientsMu.Lock()
			delete(w.clients, wsConn)
			w.clientsMu.Unlock()
			wsConn.Close()
		}()

		for {
			if _, _, err := wsConn.NextReader(); err != nil {
				break
			}
		}
	}
}
