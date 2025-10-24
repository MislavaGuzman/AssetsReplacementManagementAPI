package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type AssetReplacementTicket struct {
	ID            int64          `json:"id"`
	TicketID      int64          `json:"ticket_id"`
	CategoryID    sql.NullInt64  `json:"category_id"`
	NoSerial      sql.NullString `json:"no_serial"`
	OrderNumber   sql.NullString `json:"order_number"`
	OrderStage    sql.NullString `json:"order_stage"`
	Capex         sql.NullString `json:"capex"`
	InvoiceNumber sql.NullString `json:"invoice_number"`
	LastUpdated   sql.NullTime   `json:"last_updated"`
	ETLStage      sql.NullInt64  `json:"etl_stage"`
	ETLStatus     sql.NullString `json:"etl_status"`
	IsDeleted     sql.NullString `json:"is_deleted"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

type TicketStore struct {
	db *sql.DB
}

func (s *TicketStore) GetAll(ctx context.Context, stage int) ([]AssetReplacementTicket, error) {
	query := `
		SELECT 			
		    ID, 
			TICKET_ID, 
			CATEGORY_ID, 
			NO_SERIAL, 
			ORDER_NUMBER, 
			ORDER_STAGE, 
			CAPEX, 
			INVOICE_NUMBER, 
			ETL_STAGE
		FROM ASSETS_REPLACEMENT_TICKETS
		WHERE ETL_STAGE = :1 AND IS_DELETED = 'N'
		ORDER BY CREATED_AT DESC
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, stage)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tickets []AssetReplacementTicket
	for rows.Next() {
		var t AssetReplacementTicket
		err := rows.Scan(
			&t.ID,
			&t.TicketID,
			&t.CategoryID,
			&t.NoSerial,
			&t.OrderNumber,
			&t.OrderStage,
			&t.Capex,
			&t.InvoiceNumber,
			&t.ETLStage,
		)
		if err != nil {
			return nil, err
		}
		tickets = append(tickets, t)
	}

	return tickets, nil
}

func (s *TicketStore) GetByID(ctx context.Context, id int64) (*AssetReplacementTicket, error) {
	query := `
		SELECT ID, TICKET_ID, CATEGORY_ID, NO_SERIAL, ORDER_NUMBER, ORDER_STAGE, CAPEX,
		       INVOICE_NUMBER, LAST_UPDATED, ETL_STAGE, ETL_STATUS, IS_DELETED,
		       CREATED_AT, UPDATED_AT
		FROM ASSETS_REPLACEMENT_TICKETS
		WHERE ID = :1 AND IS_DELETED = 'N'
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	var t AssetReplacementTicket
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&t.ID,
		&t.TicketID,
		&t.CategoryID,
		&t.NoSerial,
		&t.OrderNumber,
		&t.OrderStage,
		&t.Capex,
		&t.InvoiceNumber,
		&t.LastUpdated,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return &t, nil
}

func (s *TicketStore) Create(ctx context.Context, t *AssetReplacementTicket) error {
	query := `
		INSERT INTO ASSETS_REPLACEMENT_TICKETS
			(TICKET_ID, CATEGORY_ID, NO_SERIAL, ORDER_NUMBER, ORDER_STAGE, CAPEX, INVOICE_NUMBER)
		VALUES (:1, :2, :3, :4, :5, :6, :7)
		RETURNING ID INTO :8
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := s.db.ExecContext(
		ctx,
		query,
		t.TicketID,
		t.CategoryID,
		t.NoSerial,
		t.OrderNumber,
		t.OrderStage,
		t.Capex,
		t.InvoiceNumber,
		sql.Out{Dest: &t.ID},
	)
	return err
}

func (s *TicketStore) Update(ctx context.Context, t *AssetReplacementTicket) error {
	query := `
		UPDATE ASSETS_REPLACEMENT_TICKETS
		SET CAPEX = :1, ORDER_NUMBER = :2, ORDER_STAGE = :3, UPDATED_AT = SYSDATE
		WHERE ID = :4
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	res, err := s.db.ExecContext(ctx, query, t.Capex, t.OrderNumber, t.OrderStage, t.ID)
	if err != nil {
		return err
	}

	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *TicketStore) Delete(ctx context.Context, id int64) error {
	query := `
		UPDATE ASSETS_REPLACEMENT_TICKETS
		SET IS_DELETED = 'Y', UPDATED_AT = SYSDATE
		WHERE ID = :1
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	res, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}

	return nil
}
