package services

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/MislavaGuzman/AssetsReplacementManagementAPI/internal/dto"
	"github.com/MislavaGuzman/AssetsReplacementManagementAPI/internal/store"
	"github.com/go-gota/gota/dataframe"
	"go.uber.org/zap"
)

type TicketService struct {
	store  store.TicketRepository
	logger *zap.SugaredLogger
}

func NewTicketService(repo store.TicketRepository, logger *zap.SugaredLogger) *TicketService {
	return &TicketService{store: repo, logger: logger}
}

func (svc *TicketService) UpsertBatch(ctx context.Context, dtos []dto.TicketUpsertDTO) (dto.TicketUpsertResponse, error) {
	var updatedCount int
	var skipped []int64

	for _, d := range dtos {
		err := svc.validateUpsert(ctx, d)
		if err != nil {
			svc.logger.Warnf("ticket %d skipped: %v", d.TicketID, err)
			skipped = append(skipped, d.TicketID)
			continue
		}

		if err := svc.store.Upsert(ctx, d); err != nil {
			svc.logger.Warnf("error updating ticket %d: %v", d.TicketID, err)
			skipped = append(skipped, d.TicketID)
			continue
		}
		updatedCount++
	}

	resp := dto.TicketUpsertResponse{
		UpdatedCount: updatedCount,
		SkippedIDs:   skipped,
		Message:      fmt.Sprintf("%d tickets actualizados de forma exitosa: ", updatedCount),
	}
	return resp, nil
}

func (svc *TicketService) validateUpsert(ctx context.Context, d dto.TicketUpsertDTO) error {
	if d.NoSerial != nil && *d.NoSerial != "" {
		exists, err := svc.store.ExistsActiveOrderWithSerial(ctx, *d.NoSerial, d.TicketID)
		if err != nil {
			return fmt.Errorf("error verificando no_serial: %w", err)
		}
		if exists {
			return fmt.Errorf("no_serial %s ya asignado a una orden activa", *d.NoSerial)
		}
	}
	return nil
}

func (svc *TicketService) UpsertBatchCSV(ctx context.Context, csvPath string) (*dto.TicketUpsertResponse, error) {
	file, err := os.Open(csvPath)
	if err != nil {
		return nil, fmt.Errorf("error abrindo archivo CSV: %w", err)
	}
	defer file.Close()

	df := dataframe.ReadCSV(file)
	requiredCols := []string{"ticket_id", "no_serial", "order_number", "capex", "invoice_number", "supplier"}

	dfCols := df.Names()
	colMap := make(map[string]bool)
	for _, col := range dfCols {
		colMap[col] = true
	}

	for _, col := range requiredCols {
		if !colMap[col] {
			return nil, fmt.Errorf("columna requerida o inválida: %s", col)
		}
	}

	var dtos []dto.TicketUpsertDTO
	for i := 0; i < df.Nrow(); i++ {
		row := df.Subset(i)
		id, err := strconv.ParseInt(row.Col("ticket_id").Elem(0).String(), 10, 64)
		if err != nil || id == 0 {
			continue
		}

		dtoRow := dto.TicketUpsertDTO{
			TicketID:      id,
			NoSerial:      toPtr(row.Col("no_serial").Elem(0).String()),
			OrderNumber:   toPtr(row.Col("order_number").Elem(0).String()),
			Capex:         toPtr(row.Col("capex").Elem(0).String()),
			InvoiceNumber: toPtr(row.Col("invoice_number").Elem(0).String()),
			Supplier:      toPtr(row.Col("supplier").Elem(0).String()),
		}

		dtos = append(dtos, dtoRow)
	}

	resp, err := svc.UpsertBatch(ctx, dtos)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func toPtr(s string) *string {
	if strings.TrimSpace(s) == "" {
		return nil
	}

	return &s
}
