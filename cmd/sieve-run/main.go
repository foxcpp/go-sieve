package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"net/textproto"
	"os"
	"strings"
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

	envData := interp.EnvelopeStatic{
		From: *envFrom,
		To:   *envTo,
	}
	msgData := interp.MessageStatic{
		Size:   int(fileInfo.Size()),
		Header: msgHdr,
	}
	data := sieve.NewRuntimeData(loadedScript, interp.DummyPolicy{},
		envData, msgData)

	ctx := context.Background()
	start = time.Now()
	if err := loadedScript.Execute(ctx, data); err != nil {
		log.Fatalln(err)
	}
	end = time.Now()
	log.Println("script executed in", end.Sub(start))

	fmt.Println("redirect:", data.RedirectAddr)
	fmt.Println("fileinfo:", data.Mailboxes)
	fmt.Println("keep:", data.ImplicitKeep || data.Keep)
	fmt.Printf("flags: %s\n", strings.Join(data.Flags, " "))
}
