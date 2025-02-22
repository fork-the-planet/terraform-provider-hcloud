package snapshot

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util"
	"github.com/hetznercloud/terraform-provider-hcloud/internal/util/hcloudutil"
)

// ResourceType is the type name of the Hetzner Cloud Snapshot resource.
const ResourceType = "hcloud_snapshot"

// Resource creates a new Terraform schema for the hcloud_floating_ip resource.
func Resource() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSnapshotCreate,
		ReadContext:   resourceSnapshotRead,
		UpdateContext: resourceSnapshotUpdate,
		DeleteContext: resourceSnapshotDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(90 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"server_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
			},
		},
	}
}

func resourceSnapshotCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	serverID := util.CastInt64(d.Get("server_id"))
	opts := hcloud.ServerCreateImageOpts{
		Type:        hcloud.ImageTypeSnapshot,
		Description: hcloud.Ptr(d.Get("description").(string)),
	}

	if labels, ok := d.GetOk("labels"); ok {
		tmpLabels := make(map[string]string)
		for k, v := range labels.(map[string]interface{}) {
			tmpLabels[k] = v.(string)
		}
		opts.Labels = tmpLabels
	}

	res, _, err := client.Server.CreateImage(ctx, &hcloud.Server{ID: serverID}, &opts)
	if err != nil {
		return hcloudutil.ErrorToDiag(err)
	}

	d.SetId(util.FormatID(res.Image.ID))
	if res.Action != nil {
		if err := hcloudutil.WaitForAction(ctx, &client.Action, res.Action); err != nil {
			return hcloudutil.ErrorToDiag(err)
		}
	}

	return resourceSnapshotRead(ctx, d, m)
}

func resourceSnapshotRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	id, err := util.ParseID(d.Id())
	if err != nil {
		log.Printf("[WARN] invalid Snapshot id (%s), removing from state: %v", d.Id(), err)
		d.SetId("")
		return nil
	}

	snapshot, _, err := client.Image.GetByID(ctx, id)
	if err != nil {
		return hcloudutil.ErrorToDiag(err)
	}
	if snapshot == nil {
		log.Printf("[WARN] Snapshot (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if snapshot.Type != hcloud.ImageTypeSnapshot {
		log.Printf("[WARN] Image (%s) is not a snapshot, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	setSnapshotSchema(d, snapshot)
	return nil
}

func resourceSnapshotUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	id, err := util.ParseID(d.Id())
	if err != nil {
		log.Printf("[WARN] invalid Snapshot id (%s), removing from state: %v", d.Id(), err)
		d.SetId("")
		return nil
	}
	image := &hcloud.Image{ID: id}

	d.Partial(true)

	if d.HasChange("description") {
		description := d.Get("description").(string)
		_, _, err := client.Image.Update(ctx, image, hcloud.ImageUpdateOpts{
			Description: hcloud.Ptr(description),
		})
		if err != nil {
			if resourceSnapshotIsNotFound(err, d) {
				return nil
			}
			return hcloudutil.ErrorToDiag(err)
		}
	}

	if d.HasChange("labels") {
		labels := d.Get("labels")
		tmpLabels := make(map[string]string)
		for k, v := range labels.(map[string]interface{}) {
			tmpLabels[k] = v.(string)
		}
		_, _, err := client.Image.Update(ctx, image, hcloud.ImageUpdateOpts{
			Labels: tmpLabels,
		})
		if err != nil {
			if resourceSnapshotIsNotFound(err, d) {
				return nil
			}
			return hcloudutil.ErrorToDiag(err)
		}
	}
	d.Partial(false)

	return resourceSnapshotRead(ctx, d, m)
}

func resourceSnapshotDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hcloud.Client)

	imageID, err := util.ParseID(d.Id())
	if err != nil {
		log.Printf("[WARN] invalid Snapshot id (%s), removing from state: %v", d.Id(), err)
		d.SetId("")
		return nil
	}
	if _, err := client.Image.Delete(ctx, &hcloud.Image{ID: imageID}); err != nil {
		if hcloud.IsError(err, hcloud.ErrorCodeNotFound) {
			// server has already been deleted
			return nil
		}
		return hcloudutil.ErrorToDiag(err)
	}

	return nil
}

func resourceSnapshotIsNotFound(err error, d *schema.ResourceData) bool {
	if hcloud.IsError(err, hcloud.ErrorCodeNotFound) {
		log.Printf("[WARN] Snapshot (%s) not found, removing from state", d.Id())
		d.SetId("")
		return true
	}
	return false
}

func setSnapshotSchema(d *schema.ResourceData, s *hcloud.Image) {
	d.SetId(util.FormatID(s.ID))
	if s.CreatedFrom != nil {
		d.Set("server_id", util.CastInt(s.CreatedFrom.ID))
	}
	d.Set("description", s.Description)
	d.Set("labels", s.Labels)
}
