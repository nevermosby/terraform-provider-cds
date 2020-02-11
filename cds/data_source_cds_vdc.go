package cds

import (
	"context"
	"log"
	"time"

	"terraform-provider-cds/cds-sdk-go/common"
	"terraform-provider-cds/cds-sdk-go/vdc"

	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceCdsVdc() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceCdsVdcRead,

		Schema: map[string]*schema.Schema{
			// "id": {
			// 	Type:     schema.TypeString,
			// 	Computed: true,
			// },
			"vdc_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "vdc ID.",
			},
			"vdc_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "vdc name.",
			},
			"result_output_file": {
				Type: schema.TypeString,
				// Required:    true,
				Optional:    true,
				Description: "Used to save results.",
			},
			"davidtest": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vdc": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"vdc_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"vdc_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"region_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						// "private_network": {
						// 	Type:     schema.TypeList,
						// 	Computed: true,
						// 	Elem: &schema.Resource{
						// 		Schema: map[string]*schema.Schema{
						// 			"PrivateId": {
						// 				Type: schema.TypeString,
						// 				Computed: true,
						// 			},
						// 			"Status": {
						// 				Type: schema.TypeString,
						// 				Computed: true,
						// 			},
						// 			"Status": {
						// 				Type: schema.TypeString,
						// 				Computed: true,
						// 			},
						// 		},
						// 	},
						// },
					},
				},
			},
		},
	}
}

func dataSourceCdsVdcRead(d *schema.ResourceData, meta interface{}) error {
	defer logElapsed("data_source.vdc.read")()

	logId := getLogId(contextNil)
	ctx := context.WithValue(context.TODO(), "logId", logId)

	vdcService := VdcService{client: meta.(*CdsClient).apiConn}
	descRequest := vdc.DescribeVdcRequest()
	if v, ok := d.GetOk("vdc_id"); ok {
		descRequest.VdcId = common.StringPtr(v.(string))
	}
	if v, ok := d.GetOk("vdc_name"); ok {
		descRequest.Keyword = common.StringPtr(v.(string))
	}

	result, err := vdcService.DescribeVdc(ctx, descRequest)
	if err != nil {
		return err
	}

	return vdcDescriptionAttributes(d, result)
}

func vdcDescriptionAttributes(d *schema.ResourceData, result vdc.DescVdcResponse) error {
	var names []string
	var out []map[string]interface{}

	for _, vdc := range result.Data {
		mapping := map[string]interface{}{
			"vdc_id":    vdc.VdcId,
			"vdc_name":  vdc.VdcName,
			"region_id": vdc.RegionId,
			// "private_network": vdc.PrivateNetwork,
			// "public_network":  vdc.PublicNetwork,
		}
		names = append(names, *vdc.VdcName)
		out = append(out, mapping)
	}

	log.Printf("[DEBUG] Received CDS vdc names: %q", names)
	d.SetId(time.Now().UTC().String())
	// TODO: sort by region id
	if err := d.Set("vdc", out); err != nil {
		return err
	}

	output, ok := d.GetOk("result_output_file")
	if ok && output.(string) != "" {
		if err := writeToFile(output.(string), result.Data); err != nil {
			return err
		}
	}

	return nil
}
