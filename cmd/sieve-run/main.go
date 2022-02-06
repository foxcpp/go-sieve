package main

import (
	"bufio"
	"context"
	"flag"
	"log"
	"net/textproto"
	"os"
	"time"

	"github.com/foxcpp/go-sieve"
	"github.com/foxcpp/go-sieve/interp"
)

func main() {
	msgPath := flag.String("eml", "", "msgPath message to process")
	scriptPath := flag.String("scriptPath", "", "scriptPath to run")
	envFrom := flag.String("from", "", "envelope from")
	envTo := flag.String("to", "", "envelope to")
	flag.Parse()

	msg, err := os.Open(*msgPath)
	if err != nil {
		log.Fatalln(err)
	}
	defer msg.Close()
	fileInfo, err := msg.Stat()
	if err != nil {
		log.Fatalln(err)
	}
	msgHdr, err := textproto.NewReader(bufio.NewReader(msg)).ReadMIMEHeader()
	if err != nil {
		log.Fatalln(err)
	}

	script, err := os.Open(*scriptPath)
	if err != nil {
		log.Fatalln(err)
	}
	defer script.Close()

	start := time.Now()
	loadedScript, err := sieve.Load(script, sieve.DefaultOptions())
	end := time.Now()
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("script loaded in", end.Sub(start))

	data := interp.NewRuntimeData(loadedScript, interp.Callback{
		RedirectAllowed: func(ctx context.Context, d *interp.RuntimeData, addr string) (bool, error) {
			return true, nil
		},
		HeaderGet: func(key string) (string, bool, error) {
			vals, ok := msgHdr[key]
			if !ok {
				return "", false, nil
			}
			return vals[0], true, nil
		},
	})
	data.MessageSize = int(fileInfo.Size())
	data.SMTP.From = *envFrom
	data.SMTP.To = *envTo

	ctx := context.Background()
	if err := loadedScript.Execute(ctx, data); err != nil {
		log.Fatalln(err)
	}

	log.Println("redirect:", data.RedirectAddr)
	log.Println("fileinfo:", data.Mailboxes)
	log.Println("keep:", data.ImplicitKeep || data.Keep)
}
