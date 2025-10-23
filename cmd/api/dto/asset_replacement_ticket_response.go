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
	ETLStage      *int64  `json:"etl_stage,omitempty"`
	ETLStatus     *string `json:"etl_status,omitempty"`
	IsDeleted     *string `json:"is_deleted,omitempty"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

func FromEntity(t *store.AssetReplacementTicket) TicketResponse {
	var (
		categoryID, etlStage                                    *int64
		noSerial, orderNumber, orderStage, capex, invoiceNumber *string
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
	if t.OrderStage.Valid {
		orderStage = &t.OrderStage.String
	}
	if t.Capex.Valid {
		capex = &t.Capex.String
	}
	if t.InvoiceNumber.Valid {
		invoiceNumber = &t.InvoiceNumber.String
	}
	if t.ETLStage.Valid {
		etlStage = &t.ETLStage.Int64
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
		ETLStage:      etlStage,
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
