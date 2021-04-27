package httpserver

import (
	"embed"
	"io/fs"
	"log"
	"net/http"

	"github.com/OpenVoIP/baresip-go/api"
	"github.com/OpenVoIP/baresip-go/pkg/wsserver"
)

//go:embed web
var static embed.FS

func CreateServer() {
	// websocket 处理
	hub := wsserver.NewHub()
	go hub.Run()

	serverRoot, _ := fs.Sub(static, "web")
	mux := http.NewServeMux()

	mux.HandleFunc("/baresip/index", api.GetConfig)
	mux.HandleFunc("/baresip/ws", func(rw http.ResponseWriter, r *http.Request) {
		wsserver.ServeWs(hub, rw, r)

	})
	mux.Handle("/", http.FileServer(http.FS(serverRoot)))

	log.Printf("http server run on %s", "localhost: 8988")
	err := http.ListenAndServe(":8988", mux)
	if err != nil {
		log.Fatal(err)
	}
}
