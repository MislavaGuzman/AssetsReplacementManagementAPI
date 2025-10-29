package dto

type TicketUpsertDTO struct {
	TicketID      int64   `json:"ticket_id"`
	NoSerial      *string `json:"no_serial,omitempty"`
	OrderNumber   *string `json:"order_number,omitempty"`
	Capex         *string `json:"capex,omitempty"`
	InvoiceNumber *string `json:"invoice_number,omitempty"`
	Supplier      *string `json:"supplier,omitempty"`
}

type TicketUpsertResponse struct {
	UpdatedCount int     `json:"updated_count"`
	SkippedIDs   []int64 `json:"skipped_ids,omitempty"`
	Message      string  `json:"message"`
}
