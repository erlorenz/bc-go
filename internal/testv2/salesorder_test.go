package testv2

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/erlorenz/bc-go/bc"
	"github.com/google/uuid"
)

type salesOrder struct {
	ID                      bc.GUID          `json:"id" validate:"required"`
	Number                  string           `json:"number" validate:"required"`
	OrderDate               bc.Date          `json:"orderDate" validate:"required"`
	TotalAmountExcludingTax *float64         `json:"totalAmountExcludingTax" validate:"required"`
	LastModifiedDateTime    time.Time        `json:"lastModifiedDateTime" validate:"required"`
	SalesOrderLines         []salesOrderLine `json:"salesOrderLines" validate:"required,dive"`
	Customer                *Customer        `json:"customer"`
}

func (s salesOrder) Validate() error {
	return bc.ValidateStruct(s)
}

type salesOrderLine struct {
	ID               bc.GUID `json:"id" validate:"required"`
	LineObjectNumber string  `json:"lineObjectNumber" validate:"required"`
	Quantity         int     `json:"quantity" validate:"required"`
}

func (s salesOrderLine) Validate() error {
	return bc.ValidateStruct(s)
}

type Customer struct {
	ID     bc.GUID `json:"id" validate:"required"`
	Number string  `json:"number" validate:"required"`
}

type newSalesOrderRequest struct {
	Number         string `json:"number"`
	CustomerNumber string `json:"customerNumber"`
}

type updateSalesOrderRequest struct {
	OrderDate bc.Date `json:"orderDate"`
}

func TestSalesOrderValidation(t *testing.T) {
	id := uuid.New().String()
	b := []byte(fmt.Sprintf(`{
		"id":%q,
		"orderDate":"2024-02-01",
		"number": "XXXX",
		"totalAmountExcludingTax":0.00,
		"lastModifiedDateTime":%q,
		"salesOrderLines":[
			{
				"id": %q,
	 			"lineObjectNumber": "XXXX",
				"quantity": 1
			}
		]
		}`, id, time.Now().Format(time.RFC3339), id))

	var order salesOrder
	err := json.Unmarshal(b, &order)
	if err != nil {
		t.Error(err)
	}
	err = order.Validate()
	if err != nil {
		t.Logf("salesOrder %+v", order)
		t.Error(err)
	}

}
