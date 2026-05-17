// Template ile mesaj gönderme örneği.
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
		To:       "+905551234567",
		Template: "order_shipped",
		Variables: map[string]any{
			"name":        "Ayşe",
			"tracking_no": "TR123456789",
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Sent via template: %s -> %s\n", res.TemplateKey, res.Status)
}
