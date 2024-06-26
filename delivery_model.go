package deliveryupdate

import (
	"encoding/json"
	"time"
)

type DeliveryItem struct {
	ID          string
	RouteId     string     `json:"routeid"`
	StopId      string     `json:"stopid"`
	Customerid  string     `json:"customerid"`
	Cpfcnpj     string     `json:"cpfcnpj"`
	Dated       TimeMillis `json:"dated"`
	Status      string     `json:"status"`
	Branch_code string     `json:"branch_code"`
	Latitude    float64    `json:"latitude"`
	Longitude   float64    `json:"longitude"`
	Sku         string     `json:"sku"`
	Volume      float32    `json:"volume"`
	Products    []Product  `json:"products"`
	Update      time.Time
}

type Product struct {
	Sku    string  `json:"sku"`
	Volume float32 `json:"volume"`
}

type TimeMillis struct {
	time.Time
}

func (t *TimeMillis) UnmarshalJSON(b []byte) error {
	var ms int64

	if err := json.Unmarshal(b, &ms); err != nil {
		return err
	}

	t.Time = time.Unix(ms/1000, (ms%1000)*1000000).UTC()
	return nil

}
