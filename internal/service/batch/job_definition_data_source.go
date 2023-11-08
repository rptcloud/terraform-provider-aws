// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch

// **PLEASE DELETE THIS AND ALL TIP COMMENTS BEFORE SUBMITTING A PR FOR REVIEW!**
//
// TIP: ==== INTRODUCTION ====
// Thank you for trying the skaff tool!
//
// You have opted to include these helpful comments. They all include "TIP:"
// to help you find and remove them when you're done with them.
//
// While some aspects of this file are customized to your input, the
// scaffold tool does *not* look at the AWS API and ensure it has correct
// function, structure, and variable names. It makes guesses based on
// commonalities. You will need to make significant adjustments.
//
// In other words, as generated, this is a rough outline of the work you will
// need to do. If something doesn't make sense for your situation, get rid of
// it.

import (
	// TIP: ==== IMPORTS ====
	// This is a common set of imports but not customized to your code since
	// your code hasn't been written yet. Make sure you, your IDE, or
	// goimports -w <file> fixes these imports.
	//
	// The provider linter wants your imports to be in two groups: first,
	// standard library (i.e., "fmt" or "strings"), second, everything else.
	//
	// Also, AWS Go SDK v2 may handle nested structures differently than v1,
	// using the services/batch/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// awstypes.<Type Name>.
	"context"

	"github.com/aws/aws-sdk-go-v2/service/batch"
	batchtypes "github.com/aws/aws-sdk-go-v2/service/batch/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func newDataSourceJobDefinition(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceJobDefinition{}, nil
}

const (
	DSNameJobDefinition = "Job Definition Data Source"
)

func (r *resourceJobQueue) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("arn"),
			path.MatchRoot("name"),
		),
	}
}

type dataSourceJobDefinition struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceJobDefinition) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	resp.TypeName = "aws_batch_job_definition"
}

func (d *dataSourceJobDefinition) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": schema.StringAttribute{
				Optional:   true,
				CustomType: fwtypes.ARNType,
			},
			"id": framework.IDAttribute(),
			"name": schema.StringAttribute{
				Optional: true,
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
			"type": schema.StringAttribute{
				Computed: true,
			},
			"revision": schema.Int64Attribute{
				Optional: true,
			},
			"status": schema.StringAttribute{
				Optional: true,
				// Default: JobDefinitionStatusActive,
				// https://github.com/hashicorp/terraform-plugin-framework/issues/751#issuecomment-1799757575
				Validators: []validator.String{
					stringvalidator.OneOf(JobDefinitionStatus_Values()...),
				},
			},
		},
		Blocks: map[string]schema.Block{
			// "complex_argument": schema.ListNestedBlock{
			// 	NestedObject: schema.NestedBlockObject{
			// 		Attributes: map[string]schema.Attribute{
			// 			// TIP: Attributes that are required on a corresponding resource will be
			// 			// computed on the data source (unless required as part of the search criteria).
			// 			"nested_required": schema.StringAttribute{
			// 				Computed: true,
			// 			},
			// 			"nested_computed": schema.StringAttribute{
			// 				Computed: true,
			// 			},
			// 		},
			// 	},
			// },
		},
	}
}

func (d *dataSourceJobDefinition) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().BatchClient(ctx)

	var data dataSourceJobDefinitionData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	jd := batchtypes.JobDefinition{}

	if !data.ARN.IsNull() {
		out, err := FindJobDefinitionV2ByARN(ctx, conn, *flex.StringFromFramework(ctx, data.ARN))

		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Batch, create.ErrActionReading, DSNameJobDefinition, data.Name.String(), err),
				err.Error(),
			)
			return
		}
		jd = *out
	}

	// TODO: Paginate ListJobDefinitionsV2ByNameWithStatus
	if !data.Name.IsNull() {
		input := &batch.DescribeJobDefinitionsInput{
			JobDefinitionName: flex.StringFromFramework(ctx, data.Name),
		}

		if data.Status.IsNull() {
			active := JobDefinitionStatusActive
			input.Status = &active
		} else {
			input.Status = flex.StringFromFramework(ctx, data.Status)
		}

		jds, err := ListJobDefinitionsV2ByNameWithStatus(ctx, conn, input)

		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Batch, create.ErrActionReading, DSNameJobDefinition, data.Name.String(), err),
				err.Error(),
			)
		}

		if !data.Revision.IsNull() {
			for _, _jd := range jds {
				if *_jd.Revision == int32(*data.Revision.ValueInt64Pointer()) {
					jd = _jd
				}
			}

			if jd.JobDefinitionArn == nil {
				resp.Diagnostics.AddError(
					create.ProblemStandardMessage(names.Batch, create.ErrActionReading, DSNameJobDefinition, data.Name.String(), err),
					err.Error(),
				)
			}
		}

		if data.Revision.IsNull() {
			var latestRevision int32 = 0
			for _, _jd := range jds {
				if *_jd.Revision > latestRevision {
					latestRevision = *_jd.Revision
					jd = _jd
				}
			}
		}
	}

	data.ARN = flex.StringToFrameworkARN(ctx, jd.JobDefinitionArn)
	data.ID = flex.StringToFramework(ctx, jd.JobDefinitionArn)
	data.Name = flex.StringToFramework(ctx, jd.JobDefinitionName)
	data.Type = flex.StringToFramework(ctx, jd.Type)
	data.Revision = flex.Int32ToFramework(ctx, jd.Revision)

	// TIP: Setting a complex type.
	// complexArgument, diag := flattenComplexArgument(ctx, out.ComplexArgument)
	// resp.Diagnostics.Append(diag...)
	// data.ComplexArgument = complexArgument

	// TIP: -- 5. Set the tags
	// ignoreTagsConfig := d.Meta().IgnoreTagsConfig
	// tags := KeyValueTags(ctx, jd.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)
	// data.Tags = flex.FlattenFrameworkStringValueMapLegacy(ctx, tags.Map())

	// TIP: -- 6. Set the state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// TIP: ==== DATA STRUCTURES ====
// With Terraform Plugin-Framework configurations are deserialized into
// Go types, providing type safety without the need for type assertions.
// These structs should match the schema definition exactly, and the `tfsdk`
// tag value should match the attribute name.
//
// Nested objects are represented in their own data struct. These will
// also have a corresponding attribute type mapping for use inside flex
// functions.
//
// See more:
// https://developer.hashicorp.com/terraform/plugin/framework/handling-data/accessing-values
type dataSourceJobDefinitionData struct {
	ARN      fwtypes.ARN  `tfsdk:"arn"`
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Revision types.Int64  `tfsdk:"revision"`
	Status   types.String `tfsdk:"status"`
	Tags     types.Map    `tfsdk:"tags"`
	Type     types.String `tfsdk:"type"`
}

// type complexArgumentData struct {
// 	NestedRequired types.String `tfsdk:"nested_required"`
// 	NestedOptional types.String `tfsdk:"nested_optional"`
// }
