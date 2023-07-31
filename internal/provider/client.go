package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/qustavo/terraform-provider-voltage/internal/voltage"
)

type Client struct {
	voltage *voltage.ClientWithResponses
}

func NewClient(v *voltage.ClientWithResponses) *Client {
	return &Client{voltage: v}
}

type ClientError struct {
	op  string
	err error
}

func newClientError(op string, err error) *ClientError {
	return &ClientError{op: op, err: err}
}

func (e *ClientError) Error() string {
	return fmt.Sprintf("%s: %s", e.op, e.err.Error())
}

var (
	ErrInvalidAPIResponseBody = errors.New("invalid API response body")
)

func (c *Client) assertOK(r *http.Response, body []byte) error {
	ok := http.StatusOK
	s := r.StatusCode
	if s == ok {
		return nil
	}

	op := fmt.Sprintf("calling %s %s", r.Request.Method, r.Request.URL.Path)
	err := fmt.Errorf("Wanted StatusCode=%d, got %d (%s)", ok, s, string(body))

	return newClientError(op, err)
}

func (c *Client) CreateNode(ctx context.Context, m *nodeModel) error {
	body := voltage.PostNodeCreateJSONRequestBody{
		Name:          m.Name.ValueString(),
		Network:       m.Network.ValueString(),
		PurchasedType: m.PurchasedType.ValueString(),
		Type:          m.Type.ValueString(),
		Settings: voltage.NodeSettings{
			Autopilot: m.Settings.AutoPilot.ValueBoolPointer(),
			Grpc:      m.Settings.Grpc.ValueBoolPointer(),
			Rest:      m.Settings.Rest.ValueBoolPointer(),
			Keysend:   m.Settings.Keysend.ValueBoolPointer(),
			Whitelist: toPtr(each(
				m.Settings.Whitelist, func(w types.String) string { return w.ValueString() },
			)),
			Alias:                          m.Settings.Alias.ValueStringPointer(),
			Color:                          m.Settings.Color.ValueStringPointer(),
			Wumbo:                          m.Settings.Wumbo.ValueBoolPointer(),
			Webhook:                        m.Settings.Webhook.ValueStringPointer(),
			WebhookSecret:                  m.Settings.WebhookSecret.ValueStringPointer(),
			Minchansize:                    m.Settings.MinChanSize.ValueStringPointer(),
			Maxchansize:                    m.Settings.MaxChanSize.ValueStringPointer(),
			Autocompaction:                 m.Settings.AutoCompactation.ValueBoolPointer(),
			Defaultfeerate:                 m.Settings.DefaultFeeRate.ValueStringPointer(),
			Basefee:                        m.Settings.BaseFee.ValueStringPointer(),
			Amp:                            m.Settings.Amp.ValueBoolPointer(),
			Wtclient:                       m.Settings.WtClient.ValueBoolPointer(),
			Maxpendingchannels:             m.Settings.MaxChanSize.ValueStringPointer(),
			Allowcircularroute:             m.Settings.AllowCircularRoute.ValueBoolPointer(),
			Numgraphsyncpeers:              m.Settings.NumGraphSyncPeers.ValueStringPointer(),
			Gccanceledinvoicesonstartup:    m.Settings.GCCanceledInvoicesOnStartUp.ValueBoolPointer(),
			Gccanceledinvoicesonthefly:     m.Settings.GCCanceledInvoicesOnTheFly.ValueBoolPointer(),
			Torskipproxyforclearnettargets: m.Settings.TorSkipProxyForClearnetTargets.ValueBoolPointer(),
			Rpcmiddleware:                  m.Settings.RPCMiddleware.ValueBoolPointer(),
			Optionscidalias:                m.Settings.OptionSCIDAlias.ValueBoolPointer(),
			Zeroconf:                       m.Settings.ZeroConf.ValueBoolPointer(),
		},
	}

	tflog.Info(ctx, "Creating Node", map[string]any{"body": body})
	resp, err := c.voltage.PostNodeCreateWithResponse(ctx, body)
	if err != nil {
		return newClientError("creating node", err)
	}

	if err := c.assertOK(resp.HTTPResponse, resp.Body); err != nil {
		return err
	}

	if resp.JSON200.NodeId == nil {
		return fmt.Errorf("field `node_id` can't be nil: %w", ErrInvalidAPIResponseBody)
	}
	nodeID := *resp.JSON200.NodeId

	ctx = tflog.SetField(ctx, "node_id", nodeID)
	tflog.Info(ctx, "Node Created, waiting initialization")

	// Wait for the desired state.
	var nodeStatus string
	for nodeStatus != "waiting_init" {
		// Do not kill the API.
		time.Sleep(3 * time.Second)

		node, err := c.voltage.PostNodeWithResponse(ctx, voltage.PostNodeJSONRequestBody{
			NodeId: nodeID,
		})
		if err != nil {
			return newClientError("retrieving node", err)
		}

		if err := c.assertOK(node.HTTPResponse, node.Body); err != nil {
			return err
		}

		if node.JSON200.Status == nil {
			return fmt.Errorf("field node_id can't be nil: %w", ErrInvalidAPIResponseBody)
		}

		nodeStatus = *node.JSON200.Status
	}
	tflog.Info(ctx, "Node initialized correctly!")

	if resp.JSON200.Created == nil {
		return fmt.Errorf("field `created` can't be nil: %w", ErrInvalidAPIResponseBody)
	}
	created := *resp.JSON200.Created

	m.NodeID = types.StringValue(nodeID)
	m.Created = types.StringValue(created)

	// TODO: upload seed.
	return nil
}

func (c *Client) ReadNode(ctx context.Context, nodeID string) error {
	resp, err := c.voltage.PostNodeWithResponse(ctx, voltage.NodeRequest{
		NodeId: nodeID,
	})
	if err != nil {
		return newClientError("retrieving node", err)
	}

	return c.assertOK(resp.HTTPResponse, resp.Body)
}

func (c *Client) DeleteNode(ctx context.Context, nodeID string) error {
	resp, err := c.voltage.PostNodeDeleteWithResponse(ctx, voltage.PostNodeDeleteJSONRequestBody{
		NodeId: nodeID,
	})
	if err != nil {
		return newClientError("deleting node", err)
	}

	return c.assertOK(resp.HTTPResponse, resp.Body)
}
