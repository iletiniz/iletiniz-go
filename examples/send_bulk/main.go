// Toplu mesaj gönderme örneği.
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

	res, err := client.Messages.SendBulk(context.Background(), iletiniz.SendBulkParams{
		Template: "low_stock_alert",
		Items: []iletiniz.BulkItemInput{
			{To: "+905551111111", Variables: map[string]any{"product": "Ürün A", "stock": 3}},
			{To: "+905552222222", Variables: map[string]any{"product": "Ürün B", "stock": 1}},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Toplam: %d, Gönderilen: %d, Başarısız: %d\n", res.Total, res.Sent, res.Failed)
}
