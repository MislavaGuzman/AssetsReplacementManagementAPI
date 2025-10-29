package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/MislavaGuzman/AssetsReplacementManagementAPI/internal/dto"
)

type AssetReplacementTicket struct {
	ID            int64          `json:"id"`
	TicketID      int64          `json:"ticket_id"`
	CategoryID    sql.NullInt64  `json:"category_id"`
	NoSerial      sql.NullString `json:"no_serial"`
	OrderNumber   sql.NullString `json:"order_number"`
	Capex         sql.NullString `json:"capex"`
	InvoiceNumber sql.NullString `json:"invoice_number"`
	Supplier      sql.NullString `json:"supplier"`
	CenterDistID  sql.NullInt64  `json:"center_dist_id"`
	CenterDist    sql.NullString `json:"center_dist"`
	StageProcess  sql.NullString `json:"stage_process"`
	LastUpdated   sql.NullTime   `json:"last_updated"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     sql.NullString `json:"deleted_at"`
}

type TicketStore struct {
	db *sql.DB
}

func (s *TicketStore) GetByID(ctx context.Context, id int64) (*AssetReplacementTicket, error) {
	query := `
		SELECT ID, TICKET_ID, CATEGORY_ID, NO_SERIAL, ORDER_NUMBER, NULLIF(CAPEX, '0') AS CAPEX,,
			INVOICE_NUMBER, SUPPLIER, CENTER_DIST_ID, CENTER_DIST, STAGE_PROCESS
		FROM ASSETS_REPLACEMENT_TICKETS
		WHERE TICKET_ID = :1  
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
		&t.Capex,
		&t.InvoiceNumber,
		&t.Supplier,
		&t.CenterDist,
		&t.CenterDistID,
		&t.StageProcess,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("error fetching ticket by ID: %w", err)
	}
	return &t, nil

}

func (s *TicketStore) GetAll(ctx context.Context, stage, offset, limit int) ([]AssetReplacementTicket, int, error) {
	countQuery := `
		SELECT COUNT(*)
		FROM ASSETS_REPLACEMENT_TICKETS
		WHERE STAGE_PROCESS IN ('Request Initiated', 'Procurement Phase')
		  AND DELETED_AT IS NULL
	`

	var total int
	if err := s.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("error counting tickets: %w", err)
	}

	query := `
		SELECT ID, TICKET_ID, CATEGORY_ID, NO_SERIAL, ORDER_NUMBER, NULLIF(CAPEX, '0') AS CAPEX,
			   INVOICE_NUMBER, SUPPLIER, CENTER_DIST_ID, CENTER_DIST
		FROM ASSETS_REPLACEMENT_TICKETS
		WHERE STAGE_PROCESS IN ('Request Initiated', 'Procurement Phase')
		ORDER BY CREATED_AT DESC
		OFFSET :1 ROWS FETCH NEXT :2 ROWS ONLY
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, offset, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("error fetching tickets: %w", err)
	}
	defer rows.Close()

	var tickets []AssetReplacementTicket
	for rows.Next() {
		var t AssetReplacementTicket
		if err := rows.Scan(
			&t.ID,
			&t.TicketID,
			&t.CategoryID,
			&t.NoSerial,
			&t.OrderNumber,
			&t.Capex,
			&t.InvoiceNumber,
			&t.Supplier,
			&t.CenterDistID,
			&t.CenterDist,
		); err != nil {
			return nil, 0, fmt.Errorf("error scanning ticket: %w", err)
		}
		tickets = append(tickets, t)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}

	return tickets, total, nil
}

func (s *TicketStore) GetByFilters(ctx context.Context, filters map[string]interface{}, offset, limit int) ([]AssetReplacementTicket, int, error) {
	baseQuery := `
        SELECT
            ID, TICKET_ID, CATEGORY_ID, NO_SERIAL, ORDER_NUMBER, NULLIF(CAPEX, '0') AS CAPEX, 
            INVOICE_NUMBER, SUPPLIER, CENTER_DIST_ID, CENTER_DIST
        FROM ASSETS_REPLACEMENT_TICKETS
        WHERE
            STAGE_PROCESS IN ('Request Initiated', 'Procurement Phase')
            OR NVL(STAGE_PROCESS, 'NULL') NOT IN ('COMPLETED')
        `

	var args []interface{}
	argIndex := 1

	if v, ok := filters["TICKET_ID"]; ok {
		baseQuery += fmt.Sprintf("AND TICKET_ID = :%d", argIndex)
		args = append(args, v)
		argIndex++
	}

	if v, ok := filters["CENTER_DIST_ID"]; ok {
		baseQuery += fmt.Sprintf(" AND CENTER_DIST_ID = :%d", argIndex)
		args = append(args, v)
		argIndex++
	}
	if v, ok := filters["CENTER_DIST"]; ok {
		baseQuery += fmt.Sprintf(" AND CENTER_DIST = :%d", argIndex)
		args = append(args, v)
		argIndex++
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM (%s)", baseQuery)
	var total int
	if err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("error counting filtered tickets: %w", err)
	}

	baseQuery += fmt.Sprintf(" ORDER BY CREATED_AT DESC OFFSET :%d ROWS FETCH NEXT :%d ROWS ONLY", argIndex, argIndex+1)
	args = append(args, offset, limit)

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("error fetching filtered tickets: %w", err)
	}
	defer rows.Close()

	var tickets []AssetReplacementTicket
	for rows.Next() {
		var t AssetReplacementTicket
		if err := rows.Scan(
			&t.ID,
			&t.TicketID,
			&t.CategoryID,
			&t.NoSerial,
			&t.OrderNumber,
			&t.Capex,
			&t.InvoiceNumber,
			&t.Supplier,
			&t.CenterDistID,
			&t.CenterDist,
		); err != nil {
			return nil, 0, fmt.Errorf("error scanning ticket: %w", err)
		}
		tickets = append(tickets, t)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}

	return tickets, total, nil
}

func (s *TicketStore) Create(ctx context.Context, t *AssetReplacementTicket) error {
	query := `
		INSERT INTO ASSETS_REPLACEMENT_TICKETS
			(TICKET_ID, CATEGORY_ID, NO_SERIAL, ORDER_NUMBER, STAGE_PROCESS, CAPEX, INVOICE_NUMBER, SUPPLIER, CENTER_DIST_ID, CENTER_DIST)
		VALUES (:1, :2, :3, :4, :5, :6, :7, :8, :9, :10)
		RETURNING ID INTO :11
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
		t.StageProcess,
		t.Capex,
		t.InvoiceNumber,
		t.Supplier,
		t.CenterDistID,
		t.CenterDist,
		sql.Out{Dest: &t.ID},
	)
	if err != nil {
		return fmt.Errorf("error creating ticket: %w", err)
	}
	return nil
}

func (s *TicketStore) Upsert(ctx context.Context, d dto.TicketUpsertDTO) error {
	query := `
	MERGE INTO ASSETS_REPLACEMENT_TICKETS tgt
	USING (SELECT :1 AS TICKET_ID FROM dual) src
	ON (tgt.TICKET_ID = src.TICKET_ID)
	WHEN MATCHED THEN 
		UPDATE SET
			tgt.NO_SERIAL = NVL(:2, tgt.NO_SERIAL),
			tgt.ORDER_NUMBER = NVL(:3, tgt.ORDER_NUMBER),
			tgt.CAPEX = NVL(:4, tgt.CAPEX),
			tgt.INVOICE_NUMBER = NVL(:5, tgt.INVOICE_NUMBER),
			tgt.SUPPLIER = NVL(:6, tgt.SUPPLIER),
			tgt.LAST_UPDATED = SYSDATE,
			tgt.UPDATED_AT = SYSDATE
	WHEN NOT MATCHED THEN
		INSERT (TICKET_ID, NO_SERIAL, ORDER_NUMBER, CAPEX, INVOICE_NUMBER, SUPPLIER, CREATED_AT, UPDATED_AT)
		VALUES (:1, :2, :3, :4, :5, :6, SYSDATE, SYSDATE)

	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := s.db.ExecContext(ctx, query,
		d.TicketID,
		d.NoSerial,
		d.OrderNumber,
		d.Capex,
		d.InvoiceNumber,
		d.Supplier,
	)

	if err != nil {
		return fmt.Errorf("error upserting ticket: %w", err)
	}

	return nil
}

func (s *TicketStore) ExistsActiveOrderWithSerial(ctx context.Context, serial string, excludeTicketID int64) (bool, error) {
	query := `
		SELECT COUNT(1)
		FROM ASSETS_REPLACEMENT_TICKETS
		WHERE NO_SERIAL = :1
		  AND NVL(STAGE_PROCESS, 'NULL') NOT IN ('COMPLETED')
		  AND TICKET_ID <> :2
		  AND DELETED_AT IS NULL
	`
	var count int
	err := s.db.QueryRowContext(ctx, query, serial, excludeTicketID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *TicketStore) Delete(ctx context.Context, id int64) error {
	query := `
		UPDATE ASSETS_REPLACEMENT_TICKETS
			SET DELETED_AT = SYSDATE, UPDATED_AT = SYSDATE
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

func (s *TicketStore) Update(ctx context.Context, t *AssetReplacementTicket) error {
	query := `
		UPDATE ASSETS_REPLACEMENT_TICKETS
		SET
			NO_SERIAL = :2,
			ORDER_NUMBER = :3,
			CAPEX =:5,
			INVOICE_NUMBER = :6,
			SUPPLIER = :7,
			CENTER_DIST_ID = :8,
			CENTER_DIST = :9,
			STAGE_PROCESS = :10
            LAST_UPDATED = SYSDATE,
            UPDATED_AT = SYSDATE
			WHERE ID = :11

	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	res, err := s.db.ExecContext(
		ctx,
		query,
		t.NoSerial,
		t.OrderNumber,
		t.Capex,
		t.InvoiceNumber,
		t.Supplier,
		t.CenterDistID,
		t.CenterDist,
		t.StageProcess,
	)
	if err != nil {
		return fmt.Errorf("error updating ticket: %w", err)

	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *TicketStore) GetBasicTickets(ctx context.Context) ([]AssetReplacementTicket, error) {
	query := `
		SELECT 
			TICKET_ID, ORDER_NUMBER, NULLIF(CAPEX, '0') AS CAPEX, INVOICE_NUMBER, SUPPLIER
		FROM ASSETS_REPLACEMENT_TICKETS
		WHERE DELETED_AT IS NULL
		ORDER BY CREATED_AT DESC
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error fetching basic tickets: %w", err)
	}
	defer rows.Close()

	var tickets []AssetReplacementTicket
	for rows.Next() {
		var t AssetReplacementTicket
		if err := rows.Scan(
			&t.TicketID,
			&t.OrderNumber,
			&t.Capex,
			&t.InvoiceNumber,
			&t.Supplier,
		); err != nil {
			return nil, fmt.Errorf("error scanning basic ticket: %w", err)
		}
		tickets = append(tickets, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return tickets, nil
}
