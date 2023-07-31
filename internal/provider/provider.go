package provider

import (
	"context"
	"net/http"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/qustavo/terraform-provider-voltage/internal/voltage"
)

const (
	voltageHost = "https://api.voltage.cloud"
)

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &voltageProvider{
			version: version,
		}
	}
}

type voltageProvider struct {
	version string
}

func (p *voltageProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "voltage"
	resp.Version = p.version
}

func (p *voltageProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"token": schema.StringAttribute{
				Description: "API Token",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}

}

type voltageProviderModel struct {
	Token types.String `tfsdk:"token"`
}

func (p *voltageProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config voltageProviderModel

	diags := req.Config.Get(ctx, &config)

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	token := os.Getenv("VOLTAGE_TOKEN")
	tflog.Warn(ctx, "got token", map[string]any{"token": token})

	if !config.Token.IsNull() {
		token = config.Token.ValueString()
	}

	if token == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Unknown Voltage Token",
			"The provider cannot create the Voltage API client as there is a missing or empty value for the Volatage API Token. "+
				"Set the token value in the configuration or use the VOLTAGE_TOKEN environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	requestEditorFn := func(_ context.Context, req *http.Request) error {
		req.Header.Set("X-VOLTAGE-AUTH", token)

		return nil
	}

	client, err := voltage.NewClientWithResponses(voltageHost, voltage.WithRequestEditorFn(requestEditorFn))
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not start a new Voltage API client",
			err.Error(),
		)

		return
	}

	resp.ResourceData = client
}

func (p *voltageProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewNodeResource,
	}
}

func (p *voltageProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}
