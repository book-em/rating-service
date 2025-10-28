package reservationclient

import (
	"bookem-rating-service/util"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type ReservationClient interface {
	CanUserRateHost(ctx context.Context, guestID, hostID uint) (bool, error)
}

type reservationClient struct {
	baseURL string
}

func NewReservationClient() ReservationClient {
	return &reservationClient{
		baseURL: "http://reservation-service:8080/api",
	}
}

func (c *reservationClient) CanUserRateHost(ctx context.Context, guestID, hostID uint) (bool, error) {
	util.TEL.Info("eligibility: can user rate host", "guest_id", guestID, "host_id", hostID)

	url := fmt.Sprintf("%s/reservations/guest-stayed-with-host?guestId=%d&hostId=%d", c.baseURL, guestID, hostID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		util.TEL.Error("cannot create eligibility request", err)
		return false, err
	}
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		util.TEL.Error("eligibility request failed", err)
		return false, err
	}

	if resp.StatusCode != http.StatusOK {
		util.TEL.Error("eligibility non-200", nil, "status", resp.StatusCode, "url", url)
		return false, fmt.Errorf("eligibility host failed: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		util.TEL.Error("could not parse bytes from response", err)
		return false, err
	}

	defer resp.Body.Close()

	var obj EligibilityDTO
	if err := json.Unmarshal(bodyBytes, &obj); err != nil {
		util.TEL.Error("could not unmarshall JSON", err)
		return false, err
	}
	return obj.Eligible, nil
}
