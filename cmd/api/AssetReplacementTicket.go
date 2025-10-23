package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"

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
}

type UpdateTicketPayload struct {
	OrderNumber *string `json:"order_number,omitempty" validate:"omitempty,max=100"`
	OrderStage  *string `json:"order_stage,omitempty" validate:"omitempty,max=100"`
	Capex       *string `json:"capex,omitempty" validate:"omitempty,max=50"`
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
		OrderStage:    store.SqlString(payload.OrderStage),
		Capex:         store.SqlString(payload.Capex),
		InvoiceNumber: store.SqlString(payload.InvoiceNumber),
	}

	ctx := r.Context()
	if err := app.store.Tickets.Create(ctx, t); err != nil {
		app.logger.Errorw("Error creando ticket", "error", err)

		app.internalServerError(w, r, err)
		return
	}
	app.logger.Infof("Ticket creado: %+v", t)

	_ = app.jsonResponse(w, http.StatusCreated, t)
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

	_ = app.jsonResponse(w, http.StatusOK, ticket)
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
	if payload.OrderStage != nil {
		t.OrderStage = store.SqlString(payload.OrderStage)
	}
	if payload.Capex != nil {
		t.Capex = store.SqlString(payload.Capex)
	}

	if err := app.store.Tickets.Update(ctx, t); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	_ = app.jsonResponse(w, http.StatusOK, t)
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
	stage, _ := strconv.Atoi(stageParam)
	if stage == 0 {
		stage = 1
	}

	ctx := context.Background()
	tickets, err := app.store.Tickets.GetAll(ctx, stage)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	_ = app.jsonResponse(w, http.StatusOK, tickets)
}
