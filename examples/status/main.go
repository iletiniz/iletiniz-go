// Mesaj durumu sorgulama örneği.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/iletiniz/iletiniz-go"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Kullanım: status <job_id>")
		os.Exit(2)
	}

	client, err := iletiniz.NewClient(os.Getenv("ILETINIZ_API_KEY"))
	if err != nil {
		log.Fatal(err)
	}

	info, err := client.Messages.Retrieve(context.Background(), os.Args[1])
	if err != nil {
		if errors.Is(err, iletiniz.ErrNotFound) {
			fmt.Fprintln(os.Stderr, "Mesaj bulunamadı.")
			os.Exit(1)
		}
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", info)
}
