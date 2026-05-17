// Tek mesaj gönderme örneği.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/iletiniz/iletiniz-go"
)

func main() {
	client, err := iletiniz.NewClient(os.Getenv("ILETINIZ_API_KEY"))
	if err != nil {
		log.Fatal(err)
	}

	res, err := client.Messages.Send(context.Background(), iletiniz.SendMessageParams{
		To:   "+905551234567",
		Body: "Merhaba! Bu Iletiniz SDK ile gönderilen test mesajıdır.",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Job: %s Status: %s\n", res.JobID, res.Status)
}
