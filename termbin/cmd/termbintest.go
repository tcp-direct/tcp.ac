package main

import (
	"github.com/rs/zerolog/log"
	"tcp.ac/termbin"
)

func init() {
	termbin.UseChannel = true
}

func incoming() {
	var (
		msg      termbin.Message
		deflated []byte
		err      error
	)

	select {
	case msg = <-termbin.Msg:
		switch msg.Type {
		case "ERROR":
			log.Error().
				Str("RemoteAddr", msg.RAddr).
				Int("Size", msg.Size).
				Msg(msg.Content)
		case "INCOMING_DATA":
			log.Debug().
				Str("RemoteAddr", msg.RAddr).
				Int("Size", msg.Size).
				Msg("INCOMING_DATA")
		case "FINISH":
			log.Debug().
				Str("RemoteAddr", msg.RAddr).
				Int("Size", msg.Size).
				Msg(msg.Content)

		case "DEBUG":
			log.Debug().
				Str("RemoteAddr", msg.RAddr).
				Int("Size", msg.Size).
				Msg(msg.Content)

		case "FINAL":
			log.Info().
				Str("RemoteAddr", msg.RAddr).
				Int("Size", msg.Size).
				Msg(msg.Content)
			if termbin.Gzip {
				if deflated, err = termbin.Deflate(msg.Bytes); err != nil {
					log.Error().Err(err).Msg("DEFLATE_ERROR")
				}
				println(string(deflated))
			} else {
				println(string(msg.Bytes))
			}
		}
	}
}

//func serveTermbin() {
func main() {
	if termbin.UseChannel {
		go func() {
			for {
				incoming()
			}
		}()
	}

	err := termbin.Listen("127.0.0.1", "8888")
	if err != nil {
		println(err.Error())
		return
	}
}
