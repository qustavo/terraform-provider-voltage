package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/qustavo/terraform-provider-voltage/internal/voltage"
)

var nodeSchemaV1 = schema.Schema{
	Description: "Creates and manage a node in Voltage",
	Version:     1,
	Attributes: map[string]schema.Attribute{
		"node_id": schema.StringAttribute{
			Computed: true,
		},
		// "owner_id": schema.StringAttribute{
		// 	Computed: true,
		// },
		"created": schema.StringAttribute{
			Computed: true,
		},
		// "user_ip": schema.StringAttribute{
		// 	Computed: true,
		// },
		"network": schema.StringAttribute{
			Description: "Network the node is running on. Can be either 'testnet' or 'mainnet'.",
			Required:    true,
			Validators: []validator.String{
				stringvalidator.OneOf("mainnet", "testnet"),
			},
		},
		"purchased_type": schema.StringAttribute{
			Description: "Purchase type of the node. Can be either 'trial', 'paid', or 'ondemand'.",
			Required:    true,
			Validators: []validator.String{
				stringvalidator.OneOf("trial", "paid", "ondemand"),
			},
		},
		"type": schema.StringAttribute{
			Description: "Type of node, either 'standard' or 'lite'",
			Required:    true,
			Validators: []validator.String{
				stringvalidator.OneOf("standard", "lite"),
			},
		},
		"name": schema.StringAttribute{
			Description: "User defined node name given at creation",
			Required:    true,
		},
		"settings": schema.SingleNestedAttribute{
			Description: "Settings for the Lightning Node",
			Required:    true,
			Attributes: map[string]schema.Attribute{
				// Required fields.
				"autopilot": schema.BoolAttribute{
					Description: "When enabled, LND will turn on its autopilot feature",
					Required:    true,
				},
				"grpc": schema.BoolAttribute{
					Description: "When enabled, LND will active the gRPC API",
					Required:    true,
				},
				"rest": schema.BoolAttribute{
					Description: "When enabled, LND will active the REST API",
					Required:    true,
				},
				"keysend": schema.BoolAttribute{
					Description: "When enabled, LND will enable the Keysend feature",
					Required:    true,
				},
				"whitelist": schema.ListAttribute{
					Description: "A list of IPs that are allowed to talk to your node",
					Required:    true,
					ElementType: types.StringType,
				},
				"alias": schema.StringAttribute{
					Description: "Your node's Alias on the peer to peer network",
					Required:    true,
				},
				"color": schema.StringAttribute{
					Description: "Your node's Color on the peer to peer network",
					Required:    true,
				},

				// Optional fields.
				"wumbo": schema.BoolAttribute{
					Description: "When enabled, LND will accept Wumbo channels",
					Optional:    true,
				},
				"webhook": schema.StringAttribute{
					Description: "Your webhook endpoint if you wish to receive webhook events",
					Optional:    true,
				},
				"webhook_secret": schema.StringAttribute{
					Description: "Webhook secret used to validate the webhook is coming from us",
					Optional:    true,
					Sensitive:   true,
				},
				"minchansize": schema.StringAttribute{
					Description: "The minimum channel size your node will accept",
					Optional:    true,
				},
				"maxchansize": schema.StringAttribute{
					Description: "The maximum channel size your node will accept",
					Optional:    true,
				},
				"autocompaction": schema.BoolAttribute{
					Description: "When enabled, LND will automatically compact the databases on startup",
					Optional:    true,
				},
				"defaultfeerate": schema.StringAttribute{
					Description: "Your default fee rate for your channels",
					Optional:    true,
				},
				"basefee": schema.StringAttribute{
					Description: "Your base fee rate for your channels",
					Optional:    true,
				},
				"amp": schema.BoolAttribute{
					Description: "Enables AMP",
					Optional:    true,
				},
				"wtclient": schema.BoolAttribute{
					Description: "Enables the watchtower client",
					Optional:    true,
				},
				"maxpendingchannels": schema.StringAttribute{
					Description: "Maximum number of pending channels allowed for a single peer",
					Optional:    true,
				},
				"allowcircularroute": schema.BoolAttribute{
					Description: "If enabled, allows a payment to exit and enter the same channel",
					Optional:    true,
				},
				"numgraphsyncpeers": schema.StringAttribute{
					Description: "Number of peers used for syncing the graph",
					Optional:    true,
				},
				"gccanceledinvoicesonstartup": schema.BoolAttribute{
					Description: "If enabled, deletes cancelled invoices only when LND starts up",
					Optional:    true,
				},
				"gccanceledinvoicesonthefly": schema.BoolAttribute{
					Description: "If enabled, deletes cancelled invoices while LND is running",
					Optional:    true,
				},
				"torskipproxyforclearnettargets": schema.BoolAttribute{
					Description: "Optimization for clearnet peers. See LND Docs.",
					Optional:    true,
				},
				"rpcmiddleware": schema.BoolAttribute{
					Description: "Enables the rpcmiddleware, which can interecept certain rpc calls. See LND Docs.",
					Optional:    true,
				},
				"optionscidalias": schema.BoolAttribute{
					Description: "If enabled, and optionscidalias is also enabled, it is possible to create zeroconf channels. See lnd docs.",
					Optional:    true,
				},
				"zeroconf": schema.BoolAttribute{
					Description: "If enabled, and zeroconf is also enabled, it is possible to create zeroconf channels. See lnd docs.",
					Optional:    true,
				},
			},
		},
	},
}

type nodeModel struct {
	NodeID types.String `tfsdk:"node_id"`
	// OwnerID       types.String `tfsdk:"owner_id"`
	Created types.String `tfsdk:"created"`
	// UserIP        types.String `tfsdk:"user_ip"`
	Network       types.String `tfsdk:"network"`
	PurchasedType types.String `tfsdk:"purchased_type"`
	Type          types.String `tfsdk:"type"`
	Name          types.String `tfsdk:"name"`
	Settings      struct {
		AutoPilot                      types.Bool     `tfsdk:"autopilot"`
		Grpc                           types.Bool     `tfsdk:"grpc"`
		Rest                           types.Bool     `tfsdk:"rest"`
		Keysend                        types.Bool     `tfsdk:"keysend"`
		Whitelist                      []types.String `tfsdk:"whitelist"`
		Alias                          types.String   `tfsdk:"alias"`
		Color                          types.String   `tfsdk:"color"`
		Wumbo                          types.Bool     `tfsdk:"wumbo"`
		Webhook                        types.String   `tfsdk:"webhook"`
		WebhookSecret                  types.String   `tfsdk:"webhook_secret"`
		MinChanSize                    types.String   `tfsdk:"minchansize"`
		MaxChanSize                    types.String   `tfsdk:"maxchansize"`
		AutoCompactation               types.Bool     `tfsdk:"autocompaction"`
		DefaultFeeRate                 types.String   `tfsdk:"defaultfeerate"`
		BaseFee                        types.String   `tfsdk:"basefee"`
		Amp                            types.Bool     `tfsdk:"amp"`
		WtClient                       types.Bool     `tfsdk:"wtclient"`
		MaxPendingChannels             types.String   `tfsdk:"maxpendingchannels"`
		AllowCircularRoute             types.Bool     `tfsdk:"allowcircularroute"`
		NumGraphSyncPeers              types.String   `tfsdk:"numgraphsyncpeers"`
		GCCanceledInvoicesOnStartUp    types.Bool     `tfsdk:"gccanceledinvoicesonstartup"`
		GCCanceledInvoicesOnTheFly     types.Bool     `tfsdk:"gccanceledinvoicesonthefly"`
		TorSkipProxyForClearnetTargets types.Bool     `tfsdk:"torskipproxyforclearnettargets"`
		RPCMiddleware                  types.Bool     `tfsdk:"rpcmiddleware"`
		OptionSCIDAlias                types.Bool     `tfsdk:"optionscidalias"`
		ZeroConf                       types.Bool     `tfsdk:"zeroconf"`
	} `tfsdk:"settings"`
}

type NodeResource struct {
	client *Client
}

func NewNodeResource() resource.Resource {
	return &NodeResource{}
}

func (r *NodeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_node"
}

func (r *NodeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = nodeSchemaV1
}

func (r *NodeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*voltage.ClientWithResponses)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected '*voltage.Client', got: '%T'. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = NewClient(client)
}

func errToDiags(err error) diag.Diagnostics {
	if err == nil {
		return nil
	}

	var (
		diags   diag.Diagnostics
		cErr    *ClientError
		summary string
	)

	if errors.As(err, &cErr) {
		summary = cErr.op
	} else if errors.Is(err, ErrInvalidAPIResponseBody) {
		summary = "The API server response was invalid"
	} else {
		summary = "There was an API error"
	}

	diags.AddError(summary, err.Error())

	return diags
}

func (r *NodeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan nodeModel

	// Get the state from the plan.
	diag := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diag...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.CreateNode(ctx, &plan); err != nil {
		resp.Diagnostics.Append(errToDiags(err)...)

		return
	}

	resp.Diagnostics.Append(
		resp.State.Set(ctx, &plan)...,
	)
}

func (r *NodeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state nodeModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.ReadNode(ctx, state.NodeID.ValueString()); err != nil {
		resp.Diagnostics.Append(errToDiags(err)...)

		return
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}
func (r *NodeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update not implemented", "You cannot update a node")
}

func (r *NodeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state nodeModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteNode(ctx, state.NodeID.ValueString()); err != nil {
		resp.Diagnostics.Append(errToDiags(err)...)

		return
	}
}
