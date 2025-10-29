package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/MislavaGuzman/AssetsReplacementManagementAPI/cmd/api/dto"
	internalDTO "github.com/MislavaGuzman/AssetsReplacementManagementAPI/internal/dto"
	"github.com/MislavaGuzman/AssetsReplacementManagementAPI/internal/store"
	"github.com/go-chi/chi/v5"
)

type CreateTicketPayload struct {
	TicketID      int64   `json:"ticket_id" validate:"required"`
	CategoryID    *int64  `json:"category_id,omitempty"`
	NoSerial      *string `json:"no_serial,omitempty"`
	OrderNumber   *string `json:"order_number,omitempty"`
	OrderStage    *string `json:"order_stage,omitempty"`
	Capex         *string `json:"capex,omitempty"`
	InvoiceNumber *string `json:"invoice_number,omitempty"`
	Supplier      *string `json:"supplier,omitempty"`
}

type UpdateTicketPayload struct {
	OrderNumber   *string `json:"order_number,omitempty" validate:"omitempty,max=100"`
	OrderStage    *string `json:"order_stage,omitempty" validate:"omitempty,max=100"`
	Capex         *string `json:"capex,omitempty" validate:"omitempty,max=50"`
	InvoiceNumber *string `json:"invoice_number,omitempty" validate:"omitempty,max=50"`
	Supplier      *string `json:"supplier,omitempty" validate:"omitempty,max=100"`
}

func (app *application) createAssetReplacementTicketHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.Info("POST /v1/asset-replacement-tickets recibido")

	var payload CreateTicketPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.logger.Errorw("Error leyendo JSON", "error", err)
		app.badRequestResponse(w, r, err)
		return
	}
	app.logger.Infof("Payload recibido: %+v", payload)

	if err := Validate.Struct(payload); err != nil {
		app.logger.Errorw("Validación fallida", "error", err)

		app.badRequestResponse(w, r, err)
		return
	}

	t := &store.AssetReplacementTicket{
		TicketID:      payload.TicketID,
		CategoryID:    store.SqlInt64(payload.CategoryID),
		NoSerial:      store.SqlString(payload.NoSerial),
		OrderNumber:   store.SqlString(payload.OrderNumber),
		Capex:         store.SqlString(payload.Capex),
		InvoiceNumber: store.SqlString(payload.InvoiceNumber),
		Supplier:      store.SqlString(payload.Supplier),
	}

	ctx := r.Context()
	if err := app.store.Tickets.Create(ctx, t); err != nil {
		app.logger.Errorw("Error creando ticket", "error", err)

		app.internalServerError(w, r, err)
		return
	}
	app.logger.Infof("Ticket creado: %+v", t)

	_ = app.jsonResponse(w, http.StatusCreated, dto.FromEntity(t))
}

func (app *application) getAssetReplacementTicketHandler(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "ticketID")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()
	ticket, err := app.store.Tickets.GetByID(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	_ = app.jsonResponse(w, http.StatusOK, dto.FromEntity(ticket))
}

func (app *application) updateAssetReplacementTicketHandler(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "ticketID")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	var payload UpdateTicketPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()
	t, err := app.store.Tickets.GetByID(ctx, id)
	if err != nil {
		app.notFoundResponse(w, r, err)
		return
	}

	if payload.OrderNumber != nil {
		t.OrderNumber = store.SqlString(payload.OrderNumber)
	}
	if payload.Capex != nil {
		t.Capex = store.SqlString(payload.Capex)
	}

	if err := app.store.Tickets.Update(ctx, t); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	_ = app.jsonResponse(w, http.StatusOK, dto.FromEntity(t))
}

func (app *application) deleteAssetReplacementTicketHandler(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "ticketID")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	ctx := r.Context()
	if err := app.store.Tickets.Delete(ctx, id); err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (app *application) getAllAssetReplacementTicketsHandler(w http.ResponseWriter, r *http.Request) {
	stageParam := r.URL.Query().Get("stage")
	pageParam := r.URL.Query().Get("page")
	limitParam := r.URL.Query().Get("limit")

	stage, _ := strconv.Atoi(stageParam)
	if stage <= 0 {
		stage = 1
	}

	page, _ := strconv.Atoi(pageParam)
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(limitParam)
	if limit <= 0 {
		limit = 50
	}

	offset := (page - 1) * limit

	ctx := r.Context()
	tickets, total, err := app.store.Tickets.GetAll(ctx, stage, offset, limit)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	totalPages := 1
	if limit > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}

	response := map[string]interface{}{
		"page":       page,
		"limit":      limit,
		"total":      total,
		"totalPages": totalPages,
		"tickets":    dto.FromEntities(tickets),
	}

	_ = app.jsonResponse(w, http.StatusOK, response)
}

func (app *application) upsertBatchHandler(w http.ResponseWriter, r *http.Request) {
	var dtos []internalDTO.TicketUpsertDTO

	if err := json.NewDecoder(r.Body).Decode(&dtos); err != nil {
		app.badRequestResponse(w, r, fmt.Errorf("error decodificando JSON: %w", err))
		return
	}

	resp, err := app.ticketService.UpsertBatch(r.Context(), dtos)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	_ = app.jsonResponse(w, http.StatusOK, resp)
}

func (app *application) upsertBatchCSVHandler(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	if err != nil {
		app.badRequestResponse(w, r, fmt.Errorf("error leyendo archivo CSV: %w", err))
		return
	}
	defer file.Close()

	tempDir := os.TempDir()
	tempPath := filepath.Join(tempDir, header.Filename)

	out, err := os.Create(tempPath)
	if err != nil {
		app.internalServerError(w, r, fmt.Errorf("error creando archivo temporal: %w", err))
		return
	}
	defer out.Close()

	_, err = out.ReadFrom(file)
	if err != nil {
		app.internalServerError(w, r, fmt.Errorf("error guardando CSV: %w", err))
		return
	}

	resp, err := app.ticketService.UpsertBatchCSV(r.Context(), tempPath)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	_ = os.Remove(tempPath)

	_ = app.jsonResponse(w, http.StatusOK, resp)
}

func (app *application) getBasicTicketsHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.Info("GET/v1/asset-replacement-tickets/basic recibido")
	ctx := r.Context()

	tickets, err := app.store.Tickets.GetBasicTickets(ctx)
	if err != nil {
		app.logger.Errorw("Error obteniendo tickets básicos", "error", err)
		app.internalServerError(w, r, err)
		return
	}

	type BasicTicket struct {
		TicketID      int64   `json:"ticket_id"`
		OrderNumber   *string `json:"order_number,omitempty"`
		Capex         *string `json:"capex,omitempty"`
		InvoiceNumber *string `json:"invoice_number,omitempty"`
		Supplier      *string `json:"supplier,omitempty"`
	}

	var response []BasicTicket
	for _, t := range tickets {
		response = append(response, BasicTicket{
			TicketID:      t.TicketID,
			OrderNumber:   StringOrNil(t.OrderNumber),
			Capex:         StringOrNil(t.Capex),
			InvoiceNumber: StringOrNil(t.InvoiceNumber),
			Supplier:      StringOrNil(t.Supplier),
		})
	}
	_ = app.jsonResponse(w, http.StatusOK, response)
}

func StringOrNil(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}
