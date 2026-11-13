package rest

import (
	"encoding/json"
	"net/http"
	apiv1 "webhook-delivery/gen/rest/v1"
	"webhook-delivery/internal/delivery/rest/common"
	"webhook-delivery/internal/delivery/rest/deliveries"
	"webhook-delivery/internal/delivery/rest/endpoints"
	"webhook-delivery/internal/delivery/rest/events"
	"webhook-delivery/internal/delivery/rest/subscriptions"
)

// aliases give embedded handlers distinct field names
type (
	deliveriesHandler    = deliveries.Handler
	endpointsHandler     = endpoints.Handler
	subscriptionsHandler = subscriptions.Handler
	eventsHandler        = events.Handler
)

type Services struct {
	Deliveries    deliveries.Interface
	Endpoints     endpoints.Interface
	Subscriptions subscriptions.Interface
	Events        events.Interface
}

type Handler struct {
	*deliveriesHandler
	*endpointsHandler
	*subscriptionsHandler
	*eventsHandler
}

func NewHandler(services Services) apiv1.ServerInterface {
	handler := &Handler{
		deliveriesHandler:    deliveries.NewHandler(services.Deliveries),
		endpointsHandler:     endpoints.NewHandler(services.Endpoints),
		subscriptionsHandler: subscriptions.NewHandler(services.Subscriptions),
		eventsHandler:        events.NewHandler(services.Events),
	}

	return apiv1.NewStrictHandlerWithOptions(handler, nil, apiv1.StrictHTTPServerOptions{
		RequestErrorHandlerFunc: func(w http.ResponseWriter, _ *http.Request, err error) {
			writeError(w, http.StatusBadRequest, apiv1.Error{Code: new("invalid_request"), Message: err.Error()})
		},

		ResponseErrorHandlerFunc: func(w http.ResponseWriter, _ *http.Request, _ error) {
			writeError(w, http.StatusInternalServerError, common.InternalError())
		},
	})
}

func writeError(w http.ResponseWriter, status int, body apiv1.Error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
