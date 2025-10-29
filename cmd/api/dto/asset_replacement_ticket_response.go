package dto

import (
	"github.com/MislavaGuzman/AssetsReplacementManagementAPI/internal/store"
)

type TicketResponse struct {
	ID            int64   `json:"id"`
	TicketID      int64   `json:"ticket_id"`
	CategoryID    *int64  `json:"category_id,omitempty"`
	NoSerial      *string `json:"no_serial,omitempty"`
	OrderNumber   *string `json:"order_number,omitempty"`
	OrderStage    *string `json:"order_stage,omitempty"`
	Capex         *string `json:"capex,omitempty"`
	InvoiceNumber *string `json:"invoice_number,omitempty"`
	Supplier      *string `json:"supplier,omitempty"`
}

func FromEntity(t *store.AssetReplacementTicket) TicketResponse {
	var (
		categoryID    *int64
		noSerial      *string
		orderNumber   *string
		orderStage    *string
		capex         *string
		invoiceNumber *string
		supplier      *string
	)

	if t.CategoryID.Valid {
		categoryID = &t.CategoryID.Int64
	}
	if t.NoSerial.Valid {
		noSerial = &t.NoSerial.String
	}
	if t.OrderNumber.Valid {
		orderNumber = &t.OrderNumber.String
	}
	if t.Capex.Valid {
		capex = &t.Capex.String
	}
	if t.InvoiceNumber.Valid {
		invoiceNumber = &t.InvoiceNumber.String
	}
	if t.Supplier.Valid {
		supplier = &t.Supplier.String
	}

	return TicketResponse{
		ID:            t.ID,
		TicketID:      t.TicketID,
		CategoryID:    categoryID,
		NoSerial:      noSerial,
		OrderNumber:   orderNumber,
		OrderStage:    orderStage,
		Capex:         capex,
		InvoiceNumber: invoiceNumber,
		Supplier:      supplier,
	}
}

func FromEntities(tickets []store.AssetReplacementTicket) map[string]interface{} {
	result := make([]TicketResponse, len(tickets))
	for i, ticket := range tickets {
		result[i] = FromEntity(&ticket)
	}

	return map[string]interface{}{
		"tickets": result,
		"count":   len(tickets),
	}
}
