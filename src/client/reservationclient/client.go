package reservationclient

import (
	"bookem-rating-service/util"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type ReservationClient interface {
	CanUserRateHost(ctx context.Context, guestID, hostID uint) (bool, error)
	CanUserRateRoom(ctx context.Context, guestID, roomID uint) (bool, error)
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
	return c.doEligibilityGET(ctx, "/reservations/guest-stayed-with-host", map[string]string{
		"guestId": fmt.Sprintf("%d", guestID),
		"hostId":  fmt.Sprintf("%d", hostID),
	})
}

func (c *reservationClient) CanUserRateRoom(ctx context.Context, guestID, roomID uint) (bool, error) {
	return c.doEligibilityGET(ctx, "/reservations/guest-stayed-in-room", map[string]string{
		"guestId": fmt.Sprintf("%d", guestID),
		"roomId":  fmt.Sprintf("%d", roomID),
	})
}

func (c *reservationClient) doEligibilityGET(ctx context.Context, path string, q map[string]string) (bool, error) {
	util.TEL.Info("eligibility request", "path", path, "query", q)

	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		util.TEL.Error("invalid base url or path", err, "base", c.baseURL, "path", path)
		return false, err
	}
	qs := u.Query()
	for k, v := range q {
		qs.Set(k, v)
	}
	u.RawQuery = qs.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		util.TEL.Error("cannot create http request", err, "url", u.String())
		return false, err
	}
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		util.TEL.Error("http do failed", err, "url", u.String())
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		util.TEL.Error("non-200 eligibility", nil, "status", resp.StatusCode, "url", u.String())
		return false, fmt.Errorf("eligibility failed: %d", resp.StatusCode)
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