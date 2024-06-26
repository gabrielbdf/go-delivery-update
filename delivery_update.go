package deliveryupdate

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

var re *regexp.Regexp = regexp.MustCompile("[^a-zA-Z0-9]+")
var ctx = context.Background()
var client *firestore.Client

func init() {
	functions.HTTP("DeliveryUpdate", DeliveryUpdate)
	functions.HTTP("DeliveryPurge", DeliveryPurge)
}

func DeliveryPurge(w http.ResponseWriter, r *http.Request) {
	err := authorizeAndInitializeClient(w, r, ctx)
	if err != nil {
		return
	}

	collectionRef := client.Collection("deliveries")
	// time now minus 90 days
	var toRemove = time.Now().Add(-time.Hour * 24 * 90)
	var query = collectionRef.Where("Dated.Time", "<", toRemove).Limit(20000).OrderBy("Dated.Time", firestore.Desc)

	var docs []*firestore.DocumentSnapshot
	docs, err = query.Documents(ctx).GetAll()
	if err == nil {
		for _, doc := range docs {
			doc.Ref.Delete(ctx)
		}
	}

	if err := json.NewEncoder(w).Encode(len(docs)); err != nil {
		http.Error(w, "Error encoding JSON: "+err.Error(), http.StatusInternalServerError)
	}
}

func DeliveryUpdate(w http.ResponseWriter, r *http.Request) {

	err := authorizeAndInitializeClient(w, r, ctx)
	if err != nil {
		return
	}

	items, err := decodeItems(r)
	if err != nil {
		http.Error(w, "Error decoding JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	items = updateItems(client, items)

	log.Default().Print("Itens no JSON: " + strconv.Itoa(len(items)))
	if err := json.NewEncoder(w).Encode(items); err != nil {
		http.Error(w, "Error encoding JSON: "+err.Error(), http.StatusInternalServerError)
	}

}

func authorizeAndInitializeClient(w http.ResponseWriter, r *http.Request, ctx context.Context) error {
	w.Header().Add("Content-Type", "application/json")
	auth := r.Header.Get("Authorization")
	if auth == "" || auth != os.Getenv("AUTH") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return fmt.Errorf("unauthorized")
	}

	var err error
	if client == nil {
		client, err = firestore.NewClient(ctx, os.Getenv("CLIENT"))
		if err != nil {
			http.Error(w, "Error initializing client: "+err.Error(), http.StatusInternalServerError)
			return fmt.Errorf("error initializing Firestore client: %v", err)
		}
	}
	return nil
}

func decodeItems(r *http.Request) ([]DeliveryItem, error) {
	var items []DeliveryItem
	err := json.NewDecoder(r.Body).Decode(&items)
	return items, err
}

func updateItems(client *firestore.Client, items []DeliveryItem) []DeliveryItem {
	bulk := client.BulkWriter(context.Background())
	collectionRef := client.Collection("deliveries")
	defer bulk.End()
	defer bulk.Flush()

	j := 0
	for _, item := range items {
		if item.Cpfcnpj != "" {
			item.ID = item.Cpfcnpj + "_" + re.ReplaceAllString(item.StopId, "")
			items[j] = item
			item.Update = time.Now()
			docRef := collectionRef.Doc(item.ID)
			bulk.Set(docRef, item)

			itemJson, err := json.Marshal(item)
			if err != nil {
				log.Fatal(err)
			}
			log.Default().Print(string(itemJson))

			if j%20 == 0 {
				bulk.Flush()
			}
			j++
		}
	}

	return items[:j]
}
