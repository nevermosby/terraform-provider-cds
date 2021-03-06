package cds

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"terraform-provider-cds/cds-sdk-go/common"
	"terraform-provider-cds/cds-sdk-go/vdc"
	u "terraform-provider-cds/cds/utils"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceCdsVdc() *schema.Resource {
	return &schema.Resource{
		Create: resourceCdsVdcCreate,
		Read:   resourceCdsVdcRead,
		Update: resourceCdsVdcUpdate,
		Delete: resourceCdsVdcDelete,
		Schema: map[string]*schema.Schema{
			"vdc_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     false,
				ValidateFunc: u.ValidateStringLengthInRange(1, 36),
			},
			"region_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			"public_network": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Public Network info.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ipnum": &schema.Schema{
							Type:     schema.TypeInt,
							Required: true,
						},
						"name": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"qos": &schema.Schema{
							Type:     schema.TypeInt,
							Required: true,
						},
						"floatbandwidth": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"billingmethod": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"autorenew": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"type": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"public_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Public Network id.",
			},
		},
	}
}

func resourceCdsVdcCreate(d *schema.ResourceData, meta interface{}) error {
	defer logElapsed("resource.cds_vdc.create")()
	fmt.Println("create vdc")
	logId := getLogId(contextNil)
	ctx := context.WithValue(context.TODO(), "logId", logId)

	vdcService := VdcService{client: meta.(*CdsClient).apiConn}
	taskService := TaskService{client: meta.(*CdsClient).apiConn}
	name := d.Get("vdc_name").(string)
	region := d.Get("region_id").(string)
	var publicNetwork = make(map[string]interface{})
	if v, ok := d.GetOk("public_network"); ok {
		publicNetwork = v.(map[string]interface{})
	}

	taskId, err := vdcService.CreateVdc(ctx, name, region, publicNetwork)
	if err != nil {
		return nil
	}

	detail, err := taskService.DescribeTask(ctx, taskId)
	if err != nil {
		return err
	}
	d.SetId(*detail.Data.ResourceID)

	return resourceCdsVdcRead(d, meta)
}

func resourceCdsVdcRead(d *schema.ResourceData, meta interface{}) error {
	fmt.Println("read vdc")
	defer logElapsed("resource.cds_vdc.read")()
	logId := getLogId(contextNil)
	ctx := context.WithValue(context.TODO(), "logId", logId)

	id := d.Id()
	vdcService := VdcService{client: meta.(*CdsClient).apiConn}

	request := vdc.DescribeVdcRequest()
	result, errRet := vdcService.DescribeVdc(ctx, request)
	if errRet != nil {
		return errRet
	}
	for _, value := range result.Data {
		if *value.VdcId == id {
			d.Set("vdc_name", *value.VdcName)
			d.Set("region_id", *value.RegionId)
			if len(value.PublicNetwork) > 0 {
				d.Set("public_id", *value.PublicNetwork[0].PublicId)
			}
		}
	}

	return nil
}
func resourceCdsVdcUpdate(d *schema.ResourceData, meta interface{}) error {
	fmt.Println("update vdc")
	defer logElapsed("resource.cds_vdc.read")()
	logId := getLogId(contextNil)
	ctx := context.WithValue(context.TODO(), "logId", logId)

	id := d.Id()
	vdcService := VdcService{client: meta.(*CdsClient).apiConn}

	if d.HasChange("vdc_name") {
		return errors.New("vdc_name 不支持修改")
	}

	if d.HasChange("region_id") {
		return errors.New("region_id 不支持修改")
	}

	if d.HasChange("public_network") {

		oi, ni := d.GetChange("public_network")

		if oi == nil {
			oi = new(map[string]interface{})
		}
		if ni == nil {
			ni = new(map[string]interface{})
		}
		ois := oi.(map[string]interface{})
		nis := ni.(map[string]interface{})
		// Add public network
		if len(ois) == 0 && len(nis) > 0 {

			request := vdc.NewAddPublicNetworkRequest()
			request.VdcId = common.StringPtr(id)
			terformErr := u.Mapstructure(nis, request)
			if terformErr != nil {
				return terformErr
			}
			_, err := vdcService.client.UseVdcClient().AddPublicNetwork(request)
			if err != nil {
				return err
			}
			return nil
		}
		// Delete public network
		if len(nis) == 0 && len(ois) > 0 {
			if publicId, ok := d.GetOk("public_id"); ok {
				publicId := publicId.(string)
				if len(publicId) > 0 {
					request := vdc.NewDeletePublicNetworkRequest()
					request.PublicId = common.StringPtr(publicId)
					_, errRet := vdcService.DeletePublicNetwork(ctx, request)
					if errRet != nil {
						return errRet
					}
				}
			}
			return nil
		}
		publicId := d.Get("public_id").(string)

		// Update public network
		result := u.Merge(ois, nis)

		for key, value := range result {

			if len(value) != 2 {
				continue
			}

			switch key {
			case "ipnum":
				{
					continue
				}
			case "name":
				continue
			case "qos":
				{
					request := vdc.NewModifyPublicNetworkRequest()
					request.PublicId = common.StringPtr(publicId)
					request.Qos = common.StringPtr(value[1].(string))
					_, errRet := vdcService.ModifyPublicNetwork(ctx, request)
					if errRet != nil {
						return errRet
					}
				}
			case "floatbandwidth":
				continue
			case "billingmethod":
				continue
			case "autorenew":
				{
					i, _ := strconv.Atoi(value[0].(string))
					request := vdc.NewRenewPublicNetworkRequest()
					request.PublicId = common.StringPtr(publicId)
					request.AutoRenew = common.IntPtr(i)
					_, errRet := vdcService.RenewPublicNetwork(ctx, request)
					if errRet != nil {
						return errRet
					}
				}
			case "type":
				continue
			}
		}

		d.SetPartial("public_network")
	}
	time.Sleep(20 * time.Second)
	return nil
}

func resourceCdsVdcDelete(d *schema.ResourceData, meta interface{}) error {
	defer logElapsed("resource.cds_vdc.delete")()
	fmt.Println("delete vdc")
	logId := getLogId(contextNil)
	ctx := context.WithValue(context.TODO(), "logId", logId)

	vdcService := VdcService{client: meta.(*CdsClient).apiConn}
	id := d.Id()
	if publicId, ok := d.GetOk("public_id"); ok {
		publicId := publicId.(string)
		if len(publicId) > 0 {
			request := vdc.NewDeletePublicNetworkRequest()
			request.PublicId = common.StringPtr(publicId)
			_, errRet := vdcService.DeletePublicNetwork(ctx, request)
			if errRet != nil {
				return errRet
			}
		}
	}
	time.Sleep(60 * time.Second)

	request := vdc.NewDeleteVdcRequest()
	request.VdcId = common.StringPtr(id)

	_, errRet := vdcService.DeleteVdc(ctx, request)

	if errRet != nil {
		return errRet
	}
	time.Sleep(10 * time.Second)
	return nil
}
